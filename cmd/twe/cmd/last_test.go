package cmd

import (
	"errors"
	"testing"
	"time"

	"github.com/kgoettler/twe/cmd/twe/cmd/mocks"
	timew "github.com/kgoettler/twe/pkg/timewarrior"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCmdLast_ExportError(t *testing.T) {
	backend := mocks.NewMockBackendCmdLast(t)
	backend.EXPECT().Export([]string{"@1"}).Return(nil, errors.New("timew failed"))

	result, err := RunCmdLast(backend)

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timew failed")
}

func TestRunCmdLast_NoIntervals(t *testing.T) {
	backend := mocks.NewMockBackendCmdLast(t)
	backend.EXPECT().Export([]string{"@1"}).Return([]timew.Interval{}, nil)

	result, err := RunCmdLast(backend)

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "interval @1 not found")
}

func TestRunCmdLast_ClosedInterval(t *testing.T) {
	end, err := timew.NewDatetimeFromString("20260101T120000Z")
	require.NoError(t, err)

	backend := mocks.NewMockBackendCmdLast(t)
	backend.EXPECT().Export([]string{"@1"}).Return([]timew.Interval{
		{Start: nil, End: &end, Tags: []string{"Work"}},
	}, nil)

	result, err := RunCmdLast(backend)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, end, *result)
}

func TestRunCmdLast_OpenInterval(t *testing.T) {
	start, err := timew.NewDatetimeFromString("20260101T100000Z")
	require.NoError(t, err)

	backend := mocks.NewMockBackendCmdLast(t)
	backend.EXPECT().Export([]string{"@1"}).Return([]timew.Interval{
		{Start: &start, End: nil, Tags: []string{"Work"}},
	}, nil)

	before := time.Now()
	result, err := RunCmdLast(backend)
	after := time.Now()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Before(before), "returned time should be >= time before call")
	assert.False(t, result.After(after), "returned time should be <= time after call")
}
