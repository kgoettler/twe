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
	last, err := suite.tw.Last()
	suite.NoError(err)
	expected, err := NewDatetimeFromString("20240106T120000Z")
	suite.NoError(err)
	suite.Equal(expected, last)
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
