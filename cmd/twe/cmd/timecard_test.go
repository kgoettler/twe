package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/kgoettler/twe/cmd/twe/cmd/mocks"
	"github.com/kgoettler/twe/internal/timecard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validTimewEchoOutput = `temp.report.start: 20260101T050000Z
temp.report.end: 20260108T050000Z

[
{"id":1,"start":"20260101T140000Z","end":"20260101T220000Z","tags":["Work"]},
]
`

func TestRunCmdTimecard_ReportError(t *testing.T) {
	backend := mocks.NewMockBackendCmdTimecard(t)
	backend.EXPECT().Report([]string{"echo", ":week"}).Return(nil, errors.New("timew failed"))

	result, err := RunCmdTimecard(backend, timecard.TimecardOptions{OutputFormat: "table"})

	assert.Empty(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timew failed")
}

func TestRunCmdTimecard_DefaultArg(t *testing.T) {
	backend := mocks.NewMockBackendCmdTimecard(t)
	backend.EXPECT().Report([]string{"echo", ":week"}).Return(strings.NewReader(validTimewEchoOutput), nil)

	result, err := RunCmdTimecard(backend, timecard.TimecardOptions{OutputFormat: "table"})

	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestRunCmdTimecard_WithArgs(t *testing.T) {
	backend := mocks.NewMockBackendCmdTimecard(t)
	backend.EXPECT().Report([]string{"echo", ":month"}).Return(strings.NewReader(validTimewEchoOutput), nil)

	result, err := RunCmdTimecard(backend, timecard.TimecardOptions{OutputFormat: "table"}, ":month")

	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestRunCmdTimecard_InvalidReportData(t *testing.T) {
	backend := mocks.NewMockBackendCmdTimecard(t)
	// {"id": [} matches the JSON line regex but is invalid JSON, causing NewReport to fail.
	backend.EXPECT().Report([]string{"echo", ":week"}).Return(strings.NewReader(`{"id": [}`), nil)

	result, err := RunCmdTimecard(backend, timecard.TimecardOptions{OutputFormat: "table"})

	assert.Empty(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing")
}

func TestRunCmdTimecard_InputFileNotFound(t *testing.T) {
	backend := mocks.NewMockBackendCmdTimecard(t)
	options := timecard.TimecardOptions{InputFile: "/nonexistent/path/to/file.json", OutputFormat: "table"}

	result, err := RunCmdTimecard(backend, options)

	assert.Empty(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "opening input file")
}

func TestRunCmdTimecard_InputFile(t *testing.T) {
	backend := mocks.NewMockBackendCmdTimecard(t)
	options := timecard.TimecardOptions{
		InputFile:    "../../../internal/timecard/testdata/sample.input",
		OutputFormat: "table",
	}

	result, err := RunCmdTimecard(backend, options)

	require.NoError(t, err)
	assert.NotEmpty(t, result)
}
