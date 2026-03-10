package timecard

import (
	_ "embed"
	"strings"
	"testing"
	"time"

	timew "github.com/kgoettler/twe/pkg/timewarrior"

	"github.com/stretchr/testify/suite"
)

type TimecardTestSuite struct {
	suite.Suite
}

func TestTimecardTestSuite(t *testing.T) {
	suite.Run(t, new(TimecardTestSuite))
}

//go:embed testdata/sample.input
var sampleInput string

func (suite *TimecardTestSuite) TestReport() {
	reader := strings.NewReader(sampleInput)
	tw, err := timew.NewReport(reader)
	options := TimecardOptions{
		OutputFormat: "table",
	}
	suite.Require().NoError(err)
	_, err = Run(tw, options)
	suite.Require().NoError(err)
}

func (suite *TimecardTestSuite) TestNewTimecardData_NoIntervals() {
	startTime := &timew.Datetime{time.Date(2026, 1, 1, 5, 0, 0, 0, time.UTC)}
	report := getReport(
		suite.T(),
		"",
		startTime,
		&timew.Datetime{startTime.Add(time.Hour * 24)},
	)
	_, err := NewTimecardData(&report, TimecardOptions{})
	suite.Error(err)
}

func (suite *TimecardTestSuite) TestNewTimecardData_OpenInterval() {
	// Morning interval is open but should have duration of 18 hours on 2026/01/01
	report := getReport(
		suite.T(),
		`inc 20260101T050000Z - 20260101T110000Z # Sleep
inc 20260101T110000Z # Morning`,
		nil,
		nil,
	)
	data, err := NewTimecardData(&report, TimecardOptions{})
	suite.NoError(err)
	suite.Len(data.rows, 2)
	val, err := data.Get("Morning", time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
	suite.NoError(err, "received error: %s", err)
	suite.Equal(18*time.Hour, val)
}

func (suite *TimecardTestSuite) TestNewTimecardData_WithTotals_Increment() {
	report := getReport(
		suite.T(),
		`
inc 20260101T082208Z - 20260101T141606Z # TASK1 
inc 20260101T141606Z - 20260101T151005Z # TASK2
inc 20260101T151005Z - 20260101T155151Z # TASK3
`,
		nil,
		nil,
	)
	tcs := []struct {
		increment int
		durations []string
	}{
		{
			increment: 0,
			durations: []string{"5.899", "0.9", "0.696", "7.495"},
		},
		{
			increment: 6,
			durations: []string{"5.9", "0.9", "0.7", "7.5"},
		},
		{
			increment: 15,
			durations: []string{"6", "1", "0.75", "7.75"},
		},
	}
	for _, tc := range tcs {
		data, err := NewTimecardData(
			&report,
			TimecardOptions{IncludeTotalRow: true, Increment: tc.increment},
		)
		suite.NoError(err)
		for i, expected := range tc.durations {
			suite.Equal(expected, data.At(i, 1))
		}
	}
}

func (suite *TimecardTestSuite) TestNewTimecardData_Filters() {
	intervalString := `
inc 20260101T000000Z - 20260101T060000Z # Sleep
inc 20260101T060000Z - 20260101T090000Z # Morning
inc 20260101T090000Z - 20260101T170000Z # Work
`
	report := getReport(
		suite.T(),
		intervalString,
		nil,
		nil,
	)

	data, err := NewTimecardData(&report, TimecardOptions{Filters: []string{"Sleep"}})
	suite.NoError(err)
	suite.Len(data.rows, 1)
}

func (suite *TimecardTestSuite) TestGet_NoDataForTag() {
	report := getReport(
		suite.T(),
		"inc 20260101T050000Z - 20260102T110000Z # Sleep",
		nil,
		nil,
	)

	data, err := NewTimecardData(&report, TimecardOptions{})
	suite.Require().NoError(err)

	value, err := data.Get("Foo", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	suite.Error(err)
	suite.Zero(value)
}

func (suite *TimecardTestSuite) TestTimecardData_GetMultiDayInterval_UTC() {
	report := getReport(
		suite.T(),
		"inc 20260101T050000Z - 20260102T110000Z # Sleep",
		nil,
		nil,
	)

	data, err := NewTimecardData(&report, TimecardOptions{})
	suite.Require().NoError(err)

	// Ensure 24 hours on first day
	date := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	value, err := data.Get("Sleep", date)
	suite.Require().NoError(err)
	suite.Equal(time.Hour*24, value)

	// Ensure 6 hours on second day
	date = time.Date(2026, 1, 2, 0, 0, 0, 0, time.Local)
	value, err = data.Get("Sleep", date)
	suite.Require().NoError(err)
	suite.Equal(time.Hour*6, value)
}

func (suite *TimecardTestSuite) TestFormatDurationDecimal() {
	testCases := []struct {
		duration time.Duration
		expected string
	}{
		{0, EmptyChar},
		{15 * time.Minute, "0.25"},
		{30 * time.Minute, "0.5"},
		{45 * time.Minute, "0.75"},
		{1 * time.Hour, "1"},
	}
	for _, tc := range testCases {
		suite.Equal(tc.expected, formatDurationDecimal(tc.duration),
			"duration: %v", tc.duration)
	}
}

func (suite *TimecardTestSuite) TestRoundingFunc_6MinuteIncrement() {
	round := getRoundingFunc(6)
	tcs := [][]time.Duration{
		{time.Minute * 3, time.Minute * 6},
		{time.Minute * 6, time.Minute * 6},
		{time.Minute * 7, time.Minute * 12},
	}
	for _, tc := range tcs {
		expected := tc[1]
		actual := round(tc[0])
		suite.Equal(expected, actual)
	}
}

func getReport(t *testing.T, intervalString string, startDate *timew.Datetime, endDate *timew.Datetime) timew.Report {

	intervals := getIntervals(t, intervalString)

	if startDate == nil {
		startDate = intervals[0].Start
	}
	if endDate == nil {
		endDate = intervals[len(intervals)-1].End
		if endDate == nil {
			endDate = &timew.Datetime{intervals[len(intervals)-1].Start.Add(time.Hour)}
		}
	}

	config := map[string]string{
		"temp.report.start": startDate.String(),
		"temp.report.end":   endDate.String(),
	}

	return timew.Report{
		Config:    config,
		Intervals: intervals,
	}
}

func getIntervals(t *testing.T, intervalString string) []timew.Interval {
	values := strings.Split(intervalString, "\n")
	out := make([]timew.Interval, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		interval, err := timew.NewIntervalFromString(value)
		if err != nil {
			t.Fatalf("failed to parse interval %s: %s", value, err)
		}
		out = append(out, interval)
	}
	return out
}
