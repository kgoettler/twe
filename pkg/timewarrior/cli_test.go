package timewarrior

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CLISuite struct {
	suite.Suite

	dataDir string
}

func TestCLISuite(t *testing.T) {
	suite.Run(t, new(CLISuite))
}

func (suite *CLISuite) SetupSuite() {
	// Create temporary data directory
	dataDir, err := os.MkdirTemp("", "twe-*")
	suite.NoError(err)
	suite.dataDir = dataDir

	// Set environment variable for Timewarrior database
	os.Setenv("TIMEWARRIORDB", dataDir)

	// Bootstrap
	err = bootstrapTestDB()
	suite.NoError(err)
}

func (suite *CLISuite) TearDownSuite() {
	var err error
	err = os.RemoveAll(suite.dataDir)
	suite.NoError(err)
}

func (suite *CLISuite) TestCLI() {
	cli := NewCLI()
	suite.T().Logf("TIMEWARRIORDB: %s", os.Getenv("TIMEWARRIORDB"))
	suite.Equal("timew", cli.baseCmd)
}

func (suite *CLISuite) TestTrack() {
	now := time.Now()
	interval := Interval{
		ID:    1,
		Start: &Datetime{now.Add(-time.Hour * 2)},
		End:   &Datetime{now.Add(-time.Hour)},
		Tags:  []string{"Foo"},
	}
	cli := NewCLI()
	err := cli.Track(interval)
	suite.Require().NoError(err)
	err = cli.Delete(interval.ID)
	suite.Require().NoError(err)
}
func (suite *CLISuite) TestReport_OK() {
	cli := NewCLI()
	reader, err := cli.Report("echo")
	suite.Require().NoError(err)
	report, err := NewReport(reader)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(report)
}

func (suite *CLISuite) TestReport_Err() {
	cli := NewCLI()
	_, err := cli.Report("foo")
	suite.Require().Error(err)
}

func (suite *CLISuite) TestExport() {
	cli := NewCLI()
	intervals, err := cli.Export()
	suite.Require().NoError(err)
	suite.NotEmpty(intervals)
	suite.Equal(35, len(intervals))
}

func (suite *CLISuite) TestGetIntervalByID() {
	cli := NewCLI()
	interval, err := cli.GetIntervalByID(1)
	suite.Require().NoError(err)
	suite.Contains(interval.Tags, "Work")
}

func (suite *CLISuite) TestModify_OK() {
	cli := NewCLI()
	interval, err := cli.GetIntervalByID(1)
	suite.Require().NoError(err)

	// Add one hour to the end time
	newEndTime := Datetime{
		interval.End.Add(time.Hour),
	}
	err = cli.Modify(interval.ID, "end", newEndTime.LocalString())
	suite.Require().NoError(err)

	interval, err = cli.GetIntervalByID(1)
	suite.Require().NoError(err)
	suite.Equal(newEndTime.LocalString(), interval.End.LocalString())
}

func (suite *CLISuite) TestRetag_OK() {
	cli := NewCLI()
	interval, err := cli.GetIntervalByID(1)
	suite.Require().NoError(err)
	suite.Contains(interval.Tags, "Work")

	err = cli.Retag(interval.ID, []string{"Foo"})
	suite.Require().NoError(err)

	interval, err = cli.GetIntervalByID(1)
	suite.Require().NoError(err)
	suite.Contains(interval.Tags, "Foo")
}

func bootstrapTestDB() error {
	file, err := os.Open("testdata/sample.data")
	if err != nil {
		return fmt.Errorf("opening test data: %w", err)
	}
	defer file.Close()

	var intervals []Interval
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&intervals); err != nil {
		return fmt.Errorf("decoding test data: %w", err)
	}

	cli := NewCLI()
	for _, interval := range intervals {
		err = cli.Track(interval)
		if err != nil {
			return fmt.Errorf("recording interval: %w", err)
		}
	}

	// Create the extension directory
	extensionDir := fmt.Sprintf("%s/extensions", os.Getenv("TIMEWARRIORDB"))
	err = os.MkdirAll(extensionDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("creating extension directory: %w", err)
	}

	// Write the echo script
	echoScriptPath := fmt.Sprintf("%s/echo", extensionDir)
	echoScriptContent := []byte("#!/usr/bin/env bash\ncat -\n")
	err = os.WriteFile(echoScriptPath, echoScriptContent, 0755)
	if err != nil {
		return fmt.Errorf("writing echo script: %w", err)
	}

	return nil
}
