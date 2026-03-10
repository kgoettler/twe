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

func (suite *IntervalSuite) TestIntervalFromString_Closed() {
	value := `inc 20260101T000000Z - 20260101T010000Z # Test "Code Review"`
	interval, err := NewIntervalFromString(value)
	suite.NoError(err)
	suite.Equal(*interval.Start, Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
	suite.Equal(*interval.End, Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)})
	suite.Contains(interval.Tags, "Test", "Code Review")
	suite.Empty(interval.Annotation)
}

func (suite *IntervalSuite) TestIntervalFromString_Closed_WithAnnotation() {
	value := `inc 20260101T000000Z - 20260101T010000Z # Test "Code Review" # "This is \"my annotation\" you see"`
	interval, err := NewIntervalFromString(value)
	suite.NoError(err)
	suite.Equal(*interval.Start, Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
	suite.Equal(*interval.End, Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)})
	suite.Contains(interval.Tags, "Test", "Code Review")
	suite.Equal("This is \"my annotation\" you see", interval.Annotation)
}

func (suite *IntervalSuite) TestIntervalFromString_Open() {
	value := `inc 20260101T000000Z # Test "Code Review"`
	interval, err := NewIntervalFromString(value)
	suite.NoError(err)
	suite.Equal(*interval.Start, Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
	suite.Nil(interval.End)
	suite.Contains(interval.Tags, "Test", "Code Review")
	suite.Empty(interval.Annotation)
}

func (suite *IntervalSuite) TestIntervalFromString_Open_WithAnnotation() {
	value := `inc 20260101T000000Z # Test "Code Review" # "This is \"my annotation\" you see"`
	interval, err := NewIntervalFromString(value)
	suite.NoError(err)
	suite.Equal(*interval.Start, Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
	suite.Nil(interval.End)
	suite.Contains(interval.Tags, "Test", "Code Review")
	suite.Equal("This is \"my annotation\" you see", interval.Annotation)
}

func (suite *IntervalSuite) TestLocalize() {
	value := `inc 20260101T000000Z - 20260101T010000Z # Test "Code Review"`
	interval, err := NewIntervalFromString(value)
	suite.NoError(err)
	localInterval := interval.Localize()
	// Tests run in EST (America/New_York)
	suite.Equal(*localInterval.Start, Datetime{time.Date(2025, 12, 31, 19, 0, 0, 0, time.Local)})
	suite.Equal(*localInterval.End, Datetime{time.Date(2025, 12, 31, 20, 0, 0, 0, time.Local)})
}

func (suite *IntervalSuite) TestContains_True() {
	interval, err := NewIntervalFromString(`inc 20260101T000000Z - 20260101T010000Z # Test "Code Review"`)
	suite.NoError(err)
	res := interval.Contains(time.Date(2026, 1, 1, 0, 30, 0, 0, time.UTC))
	suite.True(res)
}

func (suite *IntervalSuite) TestContains_False() {
	interval, err := NewIntervalFromString(`inc 20260101T000000Z - 20260101T010000Z # Test "Code Review"`)
	suite.NoError(err)
	res := interval.Contains(time.Date(2026, 1, 1, 1, 30, 0, 0, time.UTC))
	suite.False(res)
}

func (suite *IntervalSuite) TestContains_True_MultiZone() {
	interval, err := NewIntervalFromString(`inc 20260101T000000Z - 20260101T010000Z # Test "Code Review"`)
	suite.NoError(err)
	res := interval.Contains(time.Date(2025, 12, 31, 19, 30, 0, 0, time.Local))
	suite.True(res)
}

func (suite *IntervalSuite) TestIsOpen_True() {
	value := `inc 20260101T000000Z # Test "Code Review"`
	interval, err := NewIntervalFromString(value)
	suite.NoError(err)
	res := interval.IsOpen()
	suite.True(res)
}

func (suite *IntervalSuite) TestIsOpen_False() {
	interval, err := NewIntervalFromString(`inc 20260101T000000Z - 20260101T010000Z # Test "Code Review"`)
	suite.NoError(err)
	res := interval.IsOpen()
	suite.False(res)
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

func (suite *IntervalSuite) TestOverlaps() {
	cases := []intervalCase{
		{
			[]string{
				"20260101T060000Z", "20260101T090000Z",
				"20260101T060000Z", "20260101T090000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 3},
		},
		{
			[]string{
				"20260101T060000Z", "20260101T090000Z",
				"20260101T070000Z", "20260101T080000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 1},
		},
		{
			[]string{
				"20260101T070000Z", "20260101T080000Z",
				"20260101T060000Z", "20260101T090000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 1},
		},
		{
			[]string{
				"20260101T060000Z", "20260101T090000Z",
				"20260101T070000Z", "20260101T100000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 2},
		},
		{
			[]string{
				"20260101T070000Z", "20260101T100000Z",
				"20260101T060000Z", "20260101T090000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 2},
		},
		{
			[]string{
				"20260101T060000Z", "",
				"20260101T070000Z", "20260101T080000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 1},
		},
		{
			[]string{
				"20260101T070000Z", "20260101T080000Z",
				"20260101T060000Z", "",
			},
			[]string{},
			overlapResult{true, time.Hour * 1},
		},
		{
			[]string{
				"20260101T070000Z", "",
				"20260101T060000Z", "20260101T080000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 1},
		},
		{
			[]string{
				"20260101T070000Z", "",
				"20260101T060000Z", "20260101T080000Z",
			},
			[]string{},
			overlapResult{true, time.Hour * 1},
		},
		{
			[]string{
				"20260101T060000Z", "",
				"20260101T070000Z", "",
			},
			[]string{},
			overlapResult{true, 0},
		},
		{
			[]string{
				"20260101T060000Z", "20260101T070000Z",
				"20260101T070000Z", "20260101T080000Z",
			},
			[]string{},
			overlapResult{false, 0},
		},
		{
			[]string{
				"20260101T070000Z", "20260101T080000Z",
				"20260101T060000Z", "20260101T070000Z",
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

func (suite *IntervalSuite) TestStartsBefore() {
	cases := []intervalCase{
		{
			[]string{
				"20260101T060000Z", "20260101T070000Z",
				"20260101T070000Z", "20260101T080000Z",
			},
			[]string{},
			true,
		},
		{
			[]string{
				"20260101T070000Z", "20260101T080000Z",
				"20260101T060000Z", "20260101T070000Z",
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

func (suite *IntervalSuite) TestUnmarshal() {
	var interval Interval
	data := []byte(`{"id":2,"start":"20221206T010000Z","end":"20221206T040000Z","tags":["Football"]}`)
	err := json.Unmarshal(data, &interval)
	suite.Require().NoError(err)
}

func (suite *IntervalSuite) TestIsClosed() {
	cases := []intervalCase{
		{
			[]string{
				"20260101T060000Z", "",
			},
			[]string{"golang"},
			false,
		},
		{
			[]string{
				"20260101T060000Z", "20260101T090000Z",
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

func (suite *IntervalSuite) TestDatabaseString() {
	cases := []intervalCase{
		{
			[]string{
				"20260101T060000Z", "",
			},
			[]string{"golang"},
			`inc 20260101T060000Z # golang`,
		},
		{
			[]string{
				"20260101T060000Z", "20260101T090000Z",
			},
			[]string{"golang"},
			`inc 20260101T060000Z - 20260101T090000Z # golang`,
		},
		{
			[]string{
				"20260101T060000Z", "20260101T090000Z",
			},
			[]string{"golang", "golang time", "foo"},
			`inc 20260101T060000Z - 20260101T090000Z # golang "golang time" foo`,
		},
	}
	for _, c := range cases {
		intervals := c.getIntervals()
		res := intervals[0].DatabaseString()
		expected, _ := c.result.(string)
		suite.Equal(expected, res)
	}
}

func (suite *IntervalSuite) TestDatetime_Local() {
	date := &Datetime{}
	_ = date.UnmarshalJSON([]byte(`"20221206T090000Z"`))
	local := date.Local()
	suite.Equal("09:00", date.TimeString())
	suite.Equal("04:00", local.TimeString())
}

func (suite *IntervalSuite) TestIntervalEqual() {
	testCases := []struct {
		name  string
		left  Interval
		right Interval
		equal bool
	}{
		{
			name:  "identical intervals (all fields)",
			left:  Interval{Start: &Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, End: &Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, Tags: []string{"foo", "bar"}, Annotation: "note"},
			right: Interval{Start: &Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, End: &Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, Tags: []string{"foo", "bar"}, Annotation: "note"},
			equal: true,
		},
		{
			name:  "different start",
			left:  Interval{Start: &Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, End: &Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, Tags: []string{"foo"}},
			right: Interval{Start: &Datetime{time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)}, End: &Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, Tags: []string{"foo"}},
			equal: false,
		},
		{
			name:  "different end",
			left:  Interval{Start: &Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, End: &Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, Tags: []string{"foo"}},
			right: Interval{Start: &Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, End: &Datetime{time.Date(2026, 1, 2, 1, 0, 0, 0, time.UTC)}, Tags: []string{"foo"}},
			equal: false,
		},
		{
			name:  "different tags",
			left:  Interval{Start: &Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, End: &Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, Tags: []string{"foo"}},
			right: Interval{Start: &Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, End: &Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, Tags: []string{"bar"}},
			equal: false,
		},
		{
			name:  "different annotation",
			left:  Interval{Start: &Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, End: &Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, Tags: []string{"foo"}, Annotation: "a"},
			right: Interval{Start: &Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, End: &Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, Tags: []string{"foo"}, Annotation: "b"},
			equal: false,
		},
		{
			name:  "nil start vs non-nil",
			left:  Interval{Start: nil, End: &Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, Tags: []string{"foo"}},
			right: Interval{Start: &Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, End: &Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, Tags: []string{"foo"}},
			equal: false,
		},
		{
			name:  "nil end vs non-nil",
			left:  Interval{Start: &Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, End: nil, Tags: []string{"foo"}},
			right: Interval{Start: &Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, End: &Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, Tags: []string{"foo"}},
			equal: false,
		},
		{
			name:  "both nil start",
			left:  Interval{Start: nil, End: &Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, Tags: []string{"foo"}},
			right: Interval{Start: nil, End: &Datetime{time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, Tags: []string{"foo"}},
			equal: true,
		},
		{
			name:  "both nil end",
			left:  Interval{Start: &Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, End: nil, Tags: []string{"foo"}},
			right: Interval{Start: &Datetime{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, End: nil, Tags: []string{"foo"}},
			equal: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			eq := tc.left.Equal(tc.right)
			suite.Equal(tc.equal, eq, "left: %#v, right: %#v", tc.left, tc.right)
		})
	}
}
