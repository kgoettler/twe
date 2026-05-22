package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/kgoettler/twe/cmd/twe/cmd/mocks"
	timew "github.com/kgoettler/twe/pkg/timewarrior"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustDatetime(t *testing.T, s string) timew.Datetime {
	t.Helper()
	dt, err := timew.NewDatetimeFromString(s)
	require.NoError(t, err)
	return dt
}

func TestRunCmdImport_InvalidJSON(t *testing.T) {
	backend := mocks.NewMockBackendCmdImport(t)

	err := RunCmdImport(backend, strings.NewReader("not json"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "decoding intervals")
}

func TestRunCmdImport_EmptyInput(t *testing.T) {
	backend := mocks.NewMockBackendCmdImport(t)

	err := RunCmdImport(backend, strings.NewReader("[]"))

	require.NoError(t, err)
}

func TestRunCmdImport_SingleInterval(t *testing.T) {
	start := mustDatetime(t, "20260101T090000Z")
	end := mustDatetime(t, "20260101T170000Z")

	interval := timew.Interval{
		Start: &start,
		End:   &end,
		Tags:  []string{"Work"},
	}

	backend := mocks.NewMockBackendCmdImport(t)
	backend.EXPECT().Track(interval).Return(nil)

	input := `[{"id":0,"start":"20260101T090000Z","end":"20260101T170000Z","tags":["Work"]}]`
	err := RunCmdImport(backend, strings.NewReader(input))

	require.NoError(t, err)
}

func TestRunCmdImport_MultipleIntervalsTrackedInOrder(t *testing.T) {
	early := mustDatetime(t, "20260101T090000Z")
	earlyEnd := mustDatetime(t, "20260101T120000Z")
	late := mustDatetime(t, "20260101T130000Z")
	lateEnd := mustDatetime(t, "20260101T170000Z")

	first := timew.Interval{Start: &early, End: &earlyEnd, Tags: []string{"Work"}}
	second := timew.Interval{Start: &late, End: &lateEnd, Tags: []string{"Meeting"}}

	backend := mocks.NewMockBackendCmdImport(t)
	// Expect calls in chronological order regardless of JSON ordering
	backend.EXPECT().Track(first).Return(nil).Once()
	backend.EXPECT().Track(second).Return(nil).Once()

	// JSON has intervals in reverse order to verify sorting
	input := `[
		{"id":0,"start":"20260101T130000Z","end":"20260101T170000Z","tags":["Meeting"]},
		{"id":0,"start":"20260101T090000Z","end":"20260101T120000Z","tags":["Work"]}
	]`
	err := RunCmdImport(backend, strings.NewReader(input))

	require.NoError(t, err)
}

func TestRunCmdImport_TrackError(t *testing.T) {
	start := mustDatetime(t, "20260101T090000Z")
	end := mustDatetime(t, "20260101T170000Z")

	interval := timew.Interval{
		Start: &start,
		End:   &end,
		Tags:  []string{"Work"},
	}

	backend := mocks.NewMockBackendCmdImport(t)
	backend.EXPECT().Track(interval).Return(errors.New("timew failed"))

	input := `[{"id":0,"start":"20260101T090000Z","end":"20260101T170000Z","tags":["Work"]}]`
	err := RunCmdImport(backend, strings.NewReader(input))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to import interval")
	assert.Contains(t, err.Error(), "timew failed")
}

func TestRunCmdImport_StopsOnFirstTrackError(t *testing.T) {
	start1 := mustDatetime(t, "20260101T090000Z")
	end1 := mustDatetime(t, "20260101T120000Z")
	start2 := mustDatetime(t, "20260101T130000Z")
	end2 := mustDatetime(t, "20260101T170000Z")

	first := timew.Interval{Start: &start1, End: &end1, Tags: []string{"Work"}}
	second := timew.Interval{Start: &start2, End: &end2, Tags: []string{"Meeting"}}

	backend := mocks.NewMockBackendCmdImport(t)
	backend.EXPECT().Track(first).Return(errors.New("disk full"))
	// second interval should never be tracked
	backend.EXPECT().Track(second).Maybe().Return(nil)

	input := `[
		{"id":0,"start":"20260101T090000Z","end":"20260101T120000Z","tags":["Work"]},
		{"id":0,"start":"20260101T130000Z","end":"20260101T170000Z","tags":["Meeting"]}
	]`
	err := RunCmdImport(backend, strings.NewReader(input))

	require.Error(t, err)
}
