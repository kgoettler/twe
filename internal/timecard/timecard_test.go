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

//func (suite *TimecardTestSuite) TestTableData_StringCSV_OK() {
//	data, err := getTableData()
//	suite.Require().NoError(err)
//
//	expected := `,Mon 01/01,Tue 01/02,Wed 01/03,Thu 01/04,Fri 01/05,Sat 01/06,Sun 01/07
//Business,9,9,9,9,9,9,9
//Morning Prep,2,2,2,2,2,2,2
//Sleeping,6,6,6,6,6,6,6
//TOTAL,17,17,17,17,17,17,17
//`
//	actual, err := data.StringCSV()
//	suite.Require().NoError(err)
//	suite.Require().Equal(expected, actual)
//
//}

func (suite *TimecardTestSuite) TestTableDataGet_OK() {
	data, err := getTableData()
	suite.Require().NoError(err)

	value, err := data.Get("Sleep", time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC))
	suite.Require().NoError(err)
	suite.Require().Equal(time.Hour*6, value)
}

func (suite *TimecardTestSuite) TestTableData_OK() {
	reader := strings.NewReader(sampleInput)
	tw, err := timew.NewReport(reader)
	suite.Require().NoError(err)
	data, err := NewTimecardData(tw, []string{"Sleep"})
	suite.Require().NoError(err)
	suite.T().Logf("%s", data.String())
}

func getTableData() (TimecardData, error) {
	reader := strings.NewReader(sampleInput)
	tw, err := timew.NewReport(reader)
	if err != nil {
		return TimecardData{}, err
	}
	data, err := NewTimecardData(tw, nil)
	if err != nil {
		return TimecardData{}, err
	}
	return data, nil
}
