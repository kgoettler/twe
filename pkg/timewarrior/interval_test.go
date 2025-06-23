package timewarrior

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type IntervalSuite struct {
	suite.Suite
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestIntervalSuite(t *testing.T) {
	suite.Run(t, new(IntervalSuite))
}

type intervalCase struct {
	times  []string
	tags   []string
	result any
}

type overlapResult struct {
	ok     bool
	amount time.Duration
}

func (c intervalCase) getIntervals() []Interval {
	// Parse the string into a time.Time struct
	parsedTimes := make([]*Datetime, len(c.times))
	for i, s := range c.times {
		if len(s) == 0 {
			parsedTimes[i] = nil
			continue
		}
		t, err := time.Parse(datetimeLayout, s)
		if err != nil {
			panic(err)
		}
		parsedTimes[i] = &Datetime{t}
	}

	intervals := make([]Interval, len(c.times)/2)
	for i := 0; i < len(c.times); i += 2 {
		intervals[i/2] = Interval{
			ID:    i,
			Start: parsedTimes[i],
			End:   parsedTimes[i+1],
			Tags:  c.tags,
		}
	}
	return intervals
}

func (suite *IntervalSuite) TestInterval_Overlaps_True() {
	cases := []intervalCase{
		{
			[]string{
				"20240101T060000Z", "20240101T090000Z",
				"20240101T060000Z", "20240101T090000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 3},
		},
		{
			[]string{
				"20240101T060000Z", "20240101T090000Z",
				"20240101T070000Z", "20240101T080000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 1},
		},
		{
			[]string{
				"20240101T070000Z", "20240101T080000Z",
				"20240101T060000Z", "20240101T090000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 1},
		},
		{
			[]string{
				"20240101T060000Z", "20240101T090000Z",
				"20240101T070000Z", "20240101T100000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 2},
		},
		{
			[]string{
				"20240101T070000Z", "20240101T100000Z",
				"20240101T060000Z", "20240101T090000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 2},
		},
		{
			[]string{
				"20240101T060000Z", "",
				"20240101T070000Z", "20240101T080000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 1},
		},
		{
			[]string{
				"20240101T070000Z", "20240101T080000Z",
				"20240101T060000Z", "",
			},
			[]string{},
			overlapResult{true, time.Hour * 1},
		},
		{
			[]string{
				"20240101T070000Z", "",
				"20240101T060000Z", "20240101T080000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 1},
		},
		{
			[]string{
				"20240101T070000Z", "",
				"20240101T060000Z", "20240101T080000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 1},
		},
		{
			[]string{
				"20240101T060000Z", "",
				"20240101T070000Z", "",
			},
			[]string{},
			overlapResult{true, 0},
		},
		{
			[]string{
				"20240101T060000Z", "20240101T070000Z",
				"20240101T070000Z", "20240101T080000Z",
			},
			[]string{},
			overlapResult{false, 0},
		},
		{
			[]string{
				"20240101T070000Z", "20240101T080000Z",
				"20240101T060000Z", "20240101T070000Z",
			},
			[]string{},
			overlapResult{false, 0},
		},
	}
	for _, c := range cases {
		intervals := c.getIntervals()
		ok, amount := intervals[0].Overlaps(intervals[1])
		expected, _ := c.result.(overlapResult)
		suite.Equal(expected.ok, ok)
		if !expected.ok || expected.amount > 0 {
			suite.Equal(expected.amount, amount, fmt.Sprintf("%s %s", intervals[0], intervals[1]))
		} else {
			suite.Greater(amount, expected.amount)
		}
	}
}

func (suite *IntervalSuite) TestInterval_StartsBefore() {
	cases := []intervalCase{
		{
			[]string{
				"20240101T060000Z", "20240101T070000Z",
				"20240101T070000Z", "20240101T080000Z",
			},
			[]string{},
			true,
		},
		{
			[]string{
				"20240101T070000Z", "20240101T080000Z",
				"20240101T060000Z", "20240101T070000Z",
			},
			[]string{},
			false,
		},
	}
	for _, c := range cases {
		intervals := c.getIntervals()
		suite.Equal(intervals[0].StartsBefore(intervals[1]), c.result.(bool))
	}
}

func (suite *IntervalSuite) TestInterval_Parse() {
	var interval Interval
	data := []byte(`{"id":2,"start":"20221206T010000Z","end":"20221206T040000Z","tags":["Football"]}`)
	err := json.Unmarshal(data, &interval)
	suite.Require().NoError(err)
}

func (suite *IntervalSuite) TestInterval_Closed() {
	cases := []intervalCase{
		{
			[]string{
				"20240101T060000Z", "",
			},
			[]string{"golang"},
			false,
		},
		{
			[]string{
				"20240101T060000Z", "20240101T090000Z",
			},
			[]string{"golang"},
			true,
		},
	}
	for _, c := range cases {
		intervals := c.getIntervals()
		closed := intervals[0].IsClosed()
		expected, _ := c.result.(bool)
		suite.Equal(expected, closed)
	}
}

func (suite *IntervalSuite) TestInterval_DatabaseString() {
	cases := []intervalCase{
		{
			[]string{
				"20240101T060000Z", "",
			},
			[]string{"golang"},
			`inc 20240101T060000Z # golang`,
		},
		{
			[]string{
				"20240101T060000Z", "20240101T090000Z",
			},
			[]string{"golang"},
			`inc 20240101T060000Z - 20240101T090000Z # golang`,
		},
		{
			[]string{
				"20240101T060000Z", "20240101T090000Z",
			},
			[]string{"golang", "golang time", "foo"},
			`inc 20240101T060000Z - 20240101T090000Z # golang "golang time" foo`,
		},
	}
	for _, c := range cases {
		intervals := c.getIntervals()
		res := intervals[0].DatabaseString()
		expected, _ := c.result.(string)
		suite.Equal(expected, res)
	}
}

func (suite *IntervalSuite) TestDatetime_String() {
	date := &Datetime{}
	_ = date.UnmarshalJSON([]byte(`"20221206T090000Z"`))
	dateStr := date.TimeString()
	suite.Equal("09:00", dateStr)
}
