package edit

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/kgoettler/twe/internal/edit/mocks"
	"github.com/kgoettler/twe/pkg/timewarrior"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func xTestModel_AddRowAndFillIn(t *testing.T) {
	tm, backend := newBlankModel(t)
	backend.On("Track", mock.Anything).Return(nil)

	tm.Type("a")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.Type("09:00")
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Type("09:30")
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Type("Test")
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}

	fm := tm.FinalModel(t, teatest.WithFinalTimeout(0))

	m, ok := fm.(Model)
	if !ok {
		t.Fatal()
	}
	assert.Equal(t, true, m.isEditing)
}

type action struct {
	Msg     tea.KeyMsg
	Asserts func(t *testing.T, m Model)
}

func (a action) Run(t *testing.T, m Model) Model {
	mnew, _ := m.Update(a.Msg)
	m2, ok := mnew.(Model)
	assert.True(t, ok)
	if a.Asserts != nil {
		a.Asserts(t, m2)
	}
	return m2
}

func keyMsg(r rune) tea.KeyMsg {
	return tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{r},
	}
}

func keyEnter() tea.KeyMsg {
	return tea.KeyMsg{
		Type: tea.KeyEnter,
	}
}

func newBlankModel(t *testing.T) (*teatest.TestModel, *mocks.MockTimewarriorBackend) {
	date := time.Date(2026, time.January, 0, 0, 0, 0, 0, time.UTC)
	dateStr := date.Format("2006-01-02")
	backend := mocks.NewMockTimewarriorBackend(t)
	backend.On("Export", []string{dateStr}).Return([]timewarrior.Interval{}, nil)
	m, err := NewModel(backend, date, nil)
	assert.NoError(t, err)
	assert.Len(t, m.data, 0)

	tm := teatest.NewTestModel(t, m)

	return tm, backend
}
