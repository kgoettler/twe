package timewarrior

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type TWReportSuite struct {
	suite.Suite
	tw *Report
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestTWReportSuite(t *testing.T) {
	suite.Run(t, new(TWReportSuite))
}

func (suite *TWReportSuite) SetupTest() {
	reader := strings.NewReader(sampleInput)
	tw, err := NewReport(reader)
	suite.Require().NoError(err)
	suite.tw = tw
}

//go:embed testdata/sample.input
var sampleInput string

func (suite *TWReportSuite) TestNewTimewarrior() {
	suite.Len(suite.tw.Intervals, 3)
}

func (suite *TWReportSuite) TestLast() {
	last := suite.tw.Last()
	suite.Equal("20240106T120000Z", last)
}

func (suite *TWReportSuite) TestGetUniqueTags() {
	require := suite.Require()
	tags := suite.tw.GetUniqueTags()
	require.NotEmpty(tags)
}

func (suite *TWReportSuite) TestDateBounds() {
	require := suite.Require()
	ti, tf, err := suite.tw.GetDateRange()
	require.NoError(err)
	require.Greater(tf.Time, ti.Time)
}

func (suite *TWReportSuite) TestDates() {
	require := suite.Require()
	dates, err := suite.tw.GetDates(nil)
	require.NoError(err)
	require.Len(dates, 1)
}

func (suite *TWReportSuite) TestIsSingleWeek_OK() {
	require := suite.Require()
	ok := suite.tw.IsSingleWeek()
	require.True(ok)
}

func (suite *TWReportSuite) TestJSONString() {
	require := suite.Require()

	// Dump to JSON string
	str, err := suite.tw.JSONString()
	require.NoError(err)
	require.NotEmpty(str)
}
