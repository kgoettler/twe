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

func (suite *TimecardTestSuite) TestTimecardData_GetMultiDayInterval_UTC() {

	startTime := &timew.Datetime{time.Date(2024, 1, 1, 5, 0, 0, 0, time.UTC)}
	endTime := &timew.Datetime{startTime.Add(time.Hour * 30)}

	intervals := []timew.Interval{
		{
			ID:    1,
			Start: startTime,
			End:   endTime,
			Tags:  []string{"Sleep"},
		},
	}

	report := getReport(
		intervals,
		*startTime,
		timew.Datetime{startTime.Add(time.Hour * 24)},
	)

	data, err := NewTimecardData(&report, nil)
	suite.Require().NoError(err)

	// Ensure 24 hours on first day
	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	value, err := data.Get("Sleep", date)
	suite.Require().NoError(err)
	suite.Equal(time.Hour*24, value)

	// Ensure 6 hours on second day
	date = time.Date(2024, 1, 2, 0, 0, 0, 0, time.Local)
	value, err = data.Get("Sleep", date)
	suite.Require().NoError(err)
	suite.Equal(time.Hour*6, value)
}

func getReport(intervals []timew.Interval, startDate timew.Datetime, endDate timew.Datetime) timew.Report {

	config := map[string]string{
		"temp.report.start": startDate.String(),
		"temp.report.end":   endDate.String(),
	}

	return timew.Report{
		Config:    config,
		Intervals: intervals,
	}
}
