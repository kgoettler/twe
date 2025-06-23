package edit

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	timew "github.com/kgoettler/twe/pkg/timewarrior"

	curse "github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TimewarriorBackend interface {
	Delete(id int) error
	Export(args ...string) ([]timew.Interval, error)
	Modify(id int, field string, value string) error
	Retag(id int, tags []string) error
	Track(interval timew.Interval) error
	Undo() error
}

type Model struct {
	// Data to display in the table
	data []Row

	// Backend timewarrior interface to use
	backend TimewarriorBackend

	// Cursor indicates where the user is currently at in the table.
	cursor *cursor

	// Help message
	help help.Model

	// Keybindings
	keys    keyMap
	Logfile *os.File

	// Message to display below the table
	message string

	// Indicates whether a cell has been selected for editing
	selected bool

	// Indicates whether at least one keystroke has been processed in the current editing session.
	firstKeyProcessed bool

	// Date the user is currently editing
	date time.Time
}

func NewModel(backend TimewarriorBackend, date time.Time) (Model, error) {
	intervals, err := backend.Export(date.Format("2006-01-02"))
	if err != nil {
		return Model{}, err
	}
	data := make([]Row, len(intervals))
	for i, interval := range intervals {
		data[i] = NewRow(interval)
	}
	cursor := NewCursor(len(data), 3)
	return Model{
		data:     data,
		backend:  backend,
		cursor:   &cursor,
		help:     help.New(),
		keys:     keys,
		selected: false,
		date:     date,
	}, nil
}

func (m *Model) loadData() error {
	intervals, err := m.backend.Export(m.date.Format("2006-01-02"))
	if err != nil {
		return err
	}
	m.data = make([]Row, len(intervals))
	for i, interval := range intervals {
		m.data[i] = NewRow(interval)
	}
	return nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Return a mutable reference to the cell currently under the cursor.
func (m *Model) GetCurrentCell() *cell {
	i := m.cursor.GetRow()
	j := m.cursor.GetCol()
	return &(m.data[i].cells[j])
}

func (m Model) moveRightWhileEditing() (tea.Model, tea.Cmd) {
	// Blur current cell
	cell := m.GetCurrentCell()
	cell.Blur()

	// Move to right
	m.cursor.Right()

	// Focus cell
	nextCell := m.GetCurrentCell()
	focusCmd := nextCell.Focus()
	m.firstKeyProcessed = false
	return m, focusCmd
}

func (m Model) moveLeftWhileEditing() (tea.Model, tea.Cmd) {
	// Blur current cell
	cell := m.GetCurrentCell()
	cell.Blur()

	// Move to left
	m.cursor.Left()

	// Focus cell
	nextCell := m.GetCurrentCell()
	focusCmd := nextCell.Focus()
	m.firstKeyProcessed = false
	return m, focusCmd
}

// Update function called when a cell is selected
func (m Model) handleEditing(msg tea.Msg) (tea.Model, tea.Cmd) {
	cell := m.GetCurrentCell()

	// If esc key is pressed, cancel selection
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter":
			cell.Blur()
			m.firstKeyProcessed = false
			m.selected = false

			// Commit the row
			return m.UpdateRow(m.data[m.cursor.GetRow()])
		case "shift+tab":
			return m.moveLeftWhileEditing()
		case "tab":
			return m.moveRightWhileEditing()
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	if !m.firstKeyProcessed && cell.CharLimit > 0 && cell.CharLimit == len(cell.Value()) {
		cell.SetValue("")
		m.firstKeyProcessed = true
	}

	// Pass all other messages to textinput.Update method.
	newCell, cmd := cell.Update(msg)
	*cell = newCell
	return m, cmd
}

// Update function called when no cell is selected (i.e. table-level control)
func (m Model) handleTableNavigation(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			m.cursor.Up()

		case key.Matches(msg, m.keys.Down):
			m.cursor.Down()

		case key.Matches(msg, m.keys.Left):
			m.cursor.Left()

		case key.Matches(msg, m.keys.Right):
			m.cursor.Right()

		case key.Matches(msg, m.keys.Add):
			m.AddRow()

		case key.Matches(msg, m.keys.Reload):
			m, cmd = m.Reload()

		case key.Matches(msg, m.keys.Remove):
			m.RemoveRow()
			_, cmd = m.Reload()

		case key.Matches(msg, m.keys.Select):
			m.selected = true
			cell := m.GetCurrentCell()
			cmd = cell.Focus()

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, m.keys.Quit):
			cmd = tea.Quit

		case key.Matches(msg, m.keys.Undo):
			m.Undo()
			_, cmd = m.Reload()
		}
	case MsgData:
		m.data = msg.data
	case MsgError:
		if msg.err != nil {
			m.message = msg.err.Error()
			cmd = clearMessage()
		} else {
			m.message = ""
		}
	case MsgReload:
		m, cmd = m.Reload()
	}
	return m, cmd
}

func clearMessage() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(7 * time.Second)
		return MsgError{nil}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.selected {
		return m.handleEditing(msg)
	}
	return m.handleTableNavigation(msg)
}

func (m Model) View() string {
	format := " %-6s %-6s %-40s"
	s := HeaderBorderStyle.Render(
		fmt.Sprintf(
			format,
			CellBorderStyle.Render(lipgloss.PlaceHorizontal(6, lipgloss.Left, "Start")),
			CellBorderStyle.Render(lipgloss.PlaceHorizontal(6, lipgloss.Left, "End")),
			"Description",
		),
	)

	// Iterate over our choices
	ts := ""
	for i, row := range m.data {
		// format row
		formattedRow := make([]string, row.GetWidth())
		for j, cell := range row.cells {
			// Determine the style of the cell
			if cell.Err != nil && len(cell.Value()) == cell.CharLimit {
				cell.TextStyle = ErrStyle
			} else if i == m.cursor.GetRow() && j == m.cursor.GetCol() {
				if cell.Focused() {
					cell.TextStyle = FocusStyle
				} else {
					cell.TextStyle = HighlightStyle
				}
			} else {
				cell.TextStyle = DefaultStyle
			}
			// Determine the value to show in the cell.
			// Cells that are empty and not focused show a dimmed placeholder
			value := cell.Value()
			if value == "" && !cell.Focused() {
				value = cell.Placeholder
				cell.TextStyle = cell.TextStyle.Faint(true)
			}
			// Note: pad by one to make room for color escape chars.
			formattedRow[j] = lipgloss.PlaceHorizontal(cell.Width+1, lipgloss.Left, cell.TextStyle.Render(value))
		}
		// Render the row
		ts += fmt.Sprintf(format+"\n", CellBorderStyle.Render(formattedRow[0]), CellBorderStyle.Render(formattedRow[1]), formattedRow[2])
	}

	// Trim trailing newline
	if len(ts) > 0 {
		ts = ts[:len(ts)-1]
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinVertical(
			lipgloss.Center,
			"",
			m.date.Format("Mon 02-Jan-2006"),
			TableBorderStyle.Render(s+"\n"+ts),
		),
		ErrStyle.Render(m.message),
		m.help.View(m.keys),
	)
}

// Commands + messages
// MsgData is used to send data from the backend to the model.
type MsgData struct {
	data []Row
}

// MsgError is used to report errors for the application to handle.
type MsgError struct {
	err error
}

// MsgReload is a message to reload the data from the backend.
type MsgReload struct{}

type MsgRemoveRow struct {
	rowIndex int
}

// Add a new row to the table.
func (m *Model) AddRow() {
	row := newRowInternal(m.date, "", "", "")

	// If the current row has an end time, copy it as the start time of the new row.
	i := m.cursor.GetRow()

	if len(m.data) == 0 {
		m.data = append(m.data, row)
		m.cursor.AddRow()
		m.cursor.Down()
		return
	}

	currentRow := m.data[i]
	if currentRow.cells[1].Value() != "" {
		row.cells[0].SetValue(currentRow.cells[1].Value())
	}

	// If there is a next row and it has a start time, copy it as the end time of the new row.
	if i < len(m.data)-1 {
		nextRow := m.data[i+1]
		if nextRow.cells[0].Value() != "" {
			row.cells[1].SetValue(row.cells[0].Value())
		}
	}

	// Insert the new row between the two rows
	m.data = append(
		m.data[:i+1],
		append(
			[]Row{row},
			m.data[i+1:]...,
		)...,
	)

	m.cursor.AddRow()
}

// Commit the given row to the Timewarrior database via the backend.
func (m *Model) CommitRow(row Row) tea.Cmd {
	return func() tea.Msg {
		err := row.Commit(m.backend)
		if err != nil {
			return MsgError{err}
		}
		return MsgReload{}
	}
}

// Reload the data from the backend and update the cursor position.
func (m Model) Reload() (Model, tea.Cmd) {
	cmd := func() tea.Msg {
		err := m.loadData()
		if err != nil {
			return MsgError{fmt.Errorf("loading data: %w", err)}
		}
		m.cursor.nrows = len(m.data)

		if len(m.data) == 0 {
			m.cursor.pos.row = 0
		} else {
			// If cursor is beyond the last row, move it to the last row.
			for m.cursor.GetRow() >= len(m.data) {
				m.cursor.Up()
			}
		}
		return MsgData{
			data: m.data,
		}
	}
	return m, cmd
}

// Remove the current row/interval from the Timewarrior database.
func (m *Model) RemoveRow() {
	// Get the current row
	i := m.cursor.GetRow()
	if len(m.data) == 0 {
		return
	}

	// If the current interval exists in the Timewarrior database, delete it.
	if m.data[i].Interval.ID > 0 {
		err := m.backend.Delete(m.data[i].Interval.ID)
		if err != nil {
			m.message = fmt.Errorf("deleting interval: %w", err).Error()
		}
	}
}

// Undo the last action taken against the Timewarrior database via the backend.
func (m *Model) Undo() {
	err := m.backend.Undo()
	if err != nil {
		m.message = fmt.Errorf("undo error: %w", err).Error()
	}
}

// Update the given row in the Timewarrior database via the backend.
func (m Model) UpdateRow(row Row) (Model, tea.Cmd) {
	// Get current cursor position
	j := m.cursor.GetCol()
	cmd := func() tea.Msg {
		// If current interval does not exist in Timewarrior, write it
		if row.Interval.ID == 0 {
			if row.Ready() {
				err := row.Commit(m.backend)
				if err != nil {
					return MsgError{err}
				}
				return MsgReload{}
			} else {
				return nil
			}
		}

		var err error
		switch j {
		case 0:
			err = row.UpdateStart(m.backend)
		case 1:
			err = row.UpdateEnd(m.backend)
		case 2:
			err = row.UpdateTags(m.backend)
		}

		if err != nil {
			return MsgError{
				fmt.Errorf("modify error: %w", err),
			}
		}
		return MsgReload{}
	}
	return m, cmd
}

// Row
// A Row represents a single row of the table and corresponds to a single interval in the Timewarrior database.
type Row struct {
	timew.Interval
	cells []cell
	date  time.Time
}

// Create a new Row from a Timewarrior interval.
func NewRow(interval timew.Interval) Row {
	return Row{
		Interval: interval,
		cells: []cell{
			newCell(interval.Start.LocalTimeString(), true),
			newCell(interval.End.LocalTimeString(), true),
			newCell(strings.Join(interval.Tags, ","), false),
		},
		date: interval.Start.Time,
	}
}

func newRowInternal(date time.Time, startTime string, endTime string, desc string) Row {
	startDatetime, _ := timew.NewDatetimeFromString(
		fmt.Sprintf(
			"%s%s00",
			date.Format("20060102"),
			strings.ReplaceAll(startTime, ":", ""),
		),
	)
	endDatetime, _ := timew.NewDatetimeFromString(
		fmt.Sprintf(
			"%s%s00",
			date.Format("20060102"),
			strings.ReplaceAll(endTime, ":", ""),
		),
	)
	interval := timew.Interval{
		Start: &startDatetime,
		End:   &endDatetime,
		Tags:  []string{""},
	}
	return Row{
		Interval: interval,
		cells: []cell{
			newCell(startTime, true),
			newCell(endTime, true),
			newCell(desc, false),
		},
		date: date,
	}
}

// Commit the row to the Timewarrior database
func (r *Row) Commit(backend TimewarriorBackend) error {
	err := r.setStartInInterval(r.date, r.cells[0].Value())
	if err != nil {
		return fmt.Errorf("setting start time: %w", err)
	}
	err = r.setEndInInterval(r.date, r.cells[1].Value())
	if err != nil {
		return fmt.Errorf("setting end time: %w", err)
	}
	r.setTagsInInterval(r.cells[2].Value())

	// Commit to Timewarrior
	err = backend.Track(r.Interval)
	if err != nil {
		return fmt.Errorf("writing to timewarrior: %w", err)
	}
	return nil
}

func (r *Row) UpdateStart(backend TimewarriorBackend) error {
	err := r.setStartInInterval(r.date, r.cells[0].Value())
	if err != nil {
		return fmt.Errorf("setting start time: %w", err)
	}

	// Commit to Timewarrior
	err = backend.Modify(r.Interval.ID, "start", r.Interval.Start.LocalString())
	if err != nil {
		return fmt.Errorf("writing to timewarrior: %w", err)
	}
	return nil
}

func (r *Row) UpdateEnd(backend TimewarriorBackend) error {
	err := r.setEndInInterval(r.date, r.cells[1].Value())
	if err != nil {
		return fmt.Errorf("setting end time: %w", err)
	}

	// Commit to Timewarrior
	err = backend.Modify(r.Interval.ID, "end", r.Interval.End.LocalString())
	if err != nil {
		return fmt.Errorf("writing to timewarrior: %w", err)
	}
	return nil
}

func (r *Row) UpdateTags(backend TimewarriorBackend) error {
	r.setTagsInInterval(r.cells[2].Value())

	// Commit to Timewarrior
	err := backend.Retag(r.Interval.ID, r.GetTags())
	if err != nil {
		return fmt.Errorf("writing to timewarrior: %w", err)
	}
	return nil
}

func (r *Row) GetTags() []string {
	tags := strings.Split(r.cells[2].Value(), ",")
	for i, tag := range tags {
		tags[i] = strings.TrimLeft(tag, " ")
	}
	return tags
}

func (r Row) GetWidth() int {
	return len(r.cells)
}

// Returns true if a Row is ready to commit
func (r Row) Ready() bool {
	return r.cells[0].Err == nil && r.cells[1].Err == nil && r.cells[2].Err == nil
}

func (r *Row) setTimeInInterval(date time.Time, timeStr string, destination *timew.Datetime) error {
	datestr := fmt.Sprintf("%s%s", date.Format("20060102"), strings.ReplaceAll(timeStr, ":", ""))
	parsedTime, err := time.ParseInLocation("200601021504", datestr, time.Local)
	if err != nil {
		return fmt.Errorf("parsing time: %w", err)
	}
	*destination = timew.Datetime{Time: parsedTime.UTC()}
	return nil
}

func (r *Row) setStartInInterval(date time.Time, timeStr string) error {
	return r.setTimeInInterval(date, timeStr, r.Interval.Start)
}

func (r *Row) setEndInInterval(date time.Time, timeStr string) error {
	return r.setTimeInInterval(date, timeStr, r.Interval.End)
}

func (r *Row) setTagsInInterval(tagStr string) {
	r.Interval.Tags = strings.Split(tagStr, ",")
	for i, tag := range r.Interval.Tags {
		r.Interval.Tags[i] = strings.TrimLeft(tag, " ")
	}
}

// Cell represents a single editable text field in the table.
type cell struct {
	textinput.Model
}

func (c cell) WithModel(m textinput.Model) cell {
	return cell{m}
}

func (c cell) Update(msg tea.Msg) (cell, tea.Cmd) {
	newModel, cmd := c.Model.Update(msg)
	return c.WithModel(newModel), cmd
}

func newCell(value string, isTime bool) cell {
	m := textinput.New()
	m.Prompt = ""
	if isTime {
		m.Placeholder = "HH:MM"
		m.CharLimit = 5
		m.Width = 5
		m.Validate = func(text string) error {
			re := regexp.MustCompile(`^(?:[01]\d|2[0-3]):[0-5]\d$`)
			if !re.MatchString(text) {
				return fmt.Errorf("Invalid format")
			}
			return nil
		}
	} else {
		m.CharLimit = 40
		m.Width = 40
		m.Placeholder = "none"
		m.Validate = func(text string) error {
			if len(text) == 0 {
				return fmt.Errorf("Must provide a tag")
			}
			return nil
		}
	}
	m.SetValue(value)
	m.Cursor.Blink = false
	m.Cursor.SetMode(curse.CursorHide)
	return cell{m}
}
