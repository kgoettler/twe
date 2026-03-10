package edit

import (
	"fmt"
	"io"
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

const (
	FieldText = iota
	FieldTime
)

type ColumnSpec struct {
	Label  string
	Width  int
	Get    func(interval timew.Interval) string
	Action func(r *Row, backend TimewarriorBackend) error
	Type   int
}

var COLUMNS = []ColumnSpec{
	{
		Label: "Start",
		Width: 5,
		Action: func(r *Row, backend TimewarriorBackend) error {
			return r.UpdateStart(backend)
		},
		Get: func(interval timew.Interval) string {
			if interval.Start != nil {
				return interval.Start.Local().TimeString()
			}
			return ""
		},
		Type: FieldTime,
	},
	{
		Label: "End",
		Width: 5,
		Action: func(r *Row, backend TimewarriorBackend) error {
			return r.UpdateEnd(backend)
		},
		Get: func(interval timew.Interval) string {
			if interval.End != nil {
				return interval.End.Local().TimeString()
			}
			return ""
		},
		Type: FieldTime,
	},
	{
		Label: "Tags",
		Width: 30,
		Action: func(r *Row, backend TimewarriorBackend) error {
			return r.UpdateTags(backend)
		},
		Get: func(interval timew.Interval) string {
			return strings.Join(interval.Tags, ",")
		},
	},
	{
		Label: "Annotation",
		Width: 30,
		Action: func(r *Row, backend TimewarriorBackend) error {
			return r.UpdateAnnotation(backend)
		},
		Get: func(interval timew.Interval) string {
			return interval.Annotation
		},
	},
}

var COLUMN_FORMAT = getFormatString(COLUMNS)

func getFormatString(columns []ColumnSpec) string {
	formatParts := make([]string, len(columns))
	for i, column := range columns {
		formatParts[i] = fmt.Sprintf("%%-%ds", column.Width)
	}
	return " " + strings.Join(formatParts, " ")
}

type TimewarriorBackend interface {
	Annotate(id int, annotation string) error
	Delete(id int) error
	Export(args ...string) ([]timew.Interval, error)
	Modify(id int, field string, value string) error
	Retag(id int, tags []string) error
	Stop(stopTime *string) error
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
	keys     keyMap
	editKeys editModeKeys

	logfile io.Writer

	log func(format string, a ...any)

	// Message to display below the table
	message string

	// Indicates whether a cell has been isEditing for editing
	isEditing bool

	// Indicates whether at least one keystroke has been processed in the current editing session.
	firstKeyProcessed bool

	// Date the user is currently editing
	date time.Time
}

func NewModel(backend TimewarriorBackend, date time.Time, logfile io.Writer) (Model, error) {
	intervals, err := backend.Export(date.Format("2006-01-02"))
	if err != nil {
		return Model{}, err
	}
	data := make([]Row, len(intervals))
	for i, interval := range intervals {
		data[i] = NewRowFromInterval(interval)
	}
	cursor := NewCursor(len(data), len(COLUMNS))
	m := Model{
		data:      data,
		backend:   backend,
		cursor:    &cursor,
		help:      help.New(),
		keys:      keys,
		editKeys:  editKeys,
		isEditing: false,
		date:      date,
		logfile:   logfile,
	}

	if logfile != nil {
		m.log = func(format string, a ...any) { fmt.Fprintf(logfile, format, a...) }
	} else {
		m.log = func(format string, a ...any) {}
	}

	return m, nil
}

func (m *Model) loadData() error {
	intervals, err := m.backend.Export(m.date.Format("2006-01-02"))
	if err != nil {
		return err
	}
	m.data = make([]Row, len(intervals))
	for i, interval := range intervals {
		m.data[i] = NewRowFromInterval(interval)
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
	case curse.BlinkMsg:
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.editKeys.Esc):
			cell.Blur()
			m.firstKeyProcessed = false
			m.isEditing = false
			return m.UpdateRow(m.data[m.cursor.GetRow()])
		case key.Matches(msg, m.editKeys.Left):
			return m.moveLeftWhileEditing()
		case key.Matches(msg, m.editKeys.Right):
			return m.moveRightWhileEditing()
		case key.Matches(msg, m.editKeys.Quit):
			return m, tea.Quit
		}
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
			m, cmd = m.AddRow()

		case key.Matches(msg, m.keys.Reload):
			m, cmd = m.Reload()

		case key.Matches(msg, m.keys.Remove):
			m, cmd = m.RemoveRow()

		case key.Matches(msg, m.keys.Select):
			m.isEditing = true
			cell := m.GetCurrentCell()
			cmd = cell.Focus()

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, m.keys.Quit):
			cmd = tea.Quit

		case key.Matches(msg, m.keys.Undo):
			m, cmd = m.Undo()
		}
	case MsgError:
		if msg.err != nil {
			m.message = msg.err.Error()
			cmd = clearMessage()
		} else {
			m.message = ""
		}
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
	if m.isEditing {
		return m.handleEditing(msg)
	}
	return m.handleTableNavigation(msg)
}

func (m Model) View() string {

	// Make header
	headerCols := make([]string, len(COLUMNS))
	for i, column := range COLUMNS {
		if i < len(COLUMNS)-1 {
			headerCols[i] = CellStyle.Render(lipgloss.PlaceHorizontal(column.Width+1, lipgloss.Left, column.Label))
		} else {
			headerCols[i] = CellStyle.BorderRight(false).Render(lipgloss.PlaceHorizontal(column.Width+1, lipgloss.Left, column.Label))
		}
	}
	headerString := TableBorderStyle.BorderBottom(true).Render(strings.Join(headerCols, ""))

	// Iterate over our choices
	dataString := ""
	for i, row := range m.data {
		// format row
		formattedRow := make([]string, row.GetWidth())
		for j, cell := range row.cells {
			if i == m.cursor.GetRow() && j == m.cursor.GetCol() {
				cell.highlight = true
			} else {
				cell.highlight = false
			}
			// Note: pad by one to make room for color escape chars.
			cellStr := lipgloss.PlaceHorizontal(COLUMNS[j].Width, lipgloss.Left, cell.TextStyle.Render(cell.View()))
			if j < len(row.cells)-1 {
				formattedRow[j] = CellStyle.Render(cellStr)
			} else {
				formattedRow[j] = CellStyle.BorderRight(false).Render(cellStr)
			}
		}
		dataString += strings.Join(formattedRow, "") + "\n"
	}

	// Trim trailing newline
	if len(dataString) > 0 {
		dataString = dataString[:len(dataString)-1]
	}

	tableString := TableBorderStyle.Render(headerString + "\n" + dataString)

	// Error message
	errString := ErrStyle.Render(m.message)

	var helpString string
	if !m.isEditing {
		helpString = lipgloss.NewStyle().PaddingLeft(1).Render(m.help.View(m.keys))
	} else {
		m.help.ShowAll = false
		helpString = m.help.Styles.ShortDesc.PaddingLeft(1).Render(m.help.View(m.editKeys))
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinVertical(
			lipgloss.Center,
			"",
			m.date.Format("Mon 02-Jan-2006"),
			tableString,
		),
		errString,
		helpString,
	)
}

// Commands + messages
// MsgError is used to report errors for the application to handle.
type MsgError struct {
	err error
}

// Add a new row to the table.
func (m Model) AddRow() (Model, tea.Cmd) {
	row := NewRow(m.date)

	// If the current row has an end time, copy it as the start time of the new row.
	i := m.cursor.GetRow()

	if len(m.data) == 0 {
		m.data = append(m.data, row)
		m.cursor.AddRow()
		m.cursor.Down()
		return m, nil
	}

	currentRow := m.data[i]
	if currentRow.cells[1].Value() != "" {
		row.cells[0].SetValue(currentRow.cells[1].Value())
	}

	// If there is a next row and it has a start time, copy it as the end time of the new row.
	if i < len(m.data)-1 {
		nextRow := m.data[i+1]
		nextRowStart := nextRow.cells[0].Value()
		if nextRowStart != "" && nextRowStart != row.cells[0].Value() {
			row.cells[1].SetValue(nextRowStart)
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
	m.cursor.Down()
	return m, nil
}

// Reload the data from the backend and update the cursor position.
func (m Model) Reload() (Model, tea.Cmd) {
	err := m.loadData()
	if err != nil {
		return m.setError(fmt.Errorf("loading data: %w", err))
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
	return m, nil
}

// Remove the current row/interval from the Timewarrior database.
func (m Model) RemoveRow() (Model, tea.Cmd) {
	i := m.cursor.GetRow()
	if len(m.data) > 0 {
		reload := false
		if m.data[i].Interval.ID > 0 {
			err := m.backend.Delete(m.data[i].Interval.ID)
			if err != nil {
				return m.setError(err)
			}
			reload = true
		}
		m.data = append(m.data[:i], m.data[i+1:]...)
		m.cursor.RemoveRow()
		for m.cursor.GetRow() >= len(m.data) {
			m.cursor.Up()
		}
		if reload {
			return m.Reload()
		}
	}
	return m, nil
}

func (m Model) setError(err error) (Model, tea.Cmd) {
	var cmd tea.Cmd
	if err != nil {
		m.message = err.Error()
		cmd = clearMessage()
	} else {
		m.message = ""
	}
	return m, cmd
}

// Undo the last action taken against the Timewarrior database via the backend.
func (m *Model) Undo() (Model, tea.Cmd) {
	err := m.backend.Undo()
	if err != nil {
		return m.setError(fmt.Errorf("undo error: %w", err))
	}
	return m.Reload()
}

// Update the given row in the Timewarrior database via the backend.
func (m Model) UpdateRow(row Row) (Model, tea.Cmd) {
	// Get current cursor position
	j := m.cursor.GetCol()
	// If current interval does not exist in Timewarrior, write it
	if row.Interval.ID == 0 {
		if row.Ready() {
			err := row.Commit(m.backend)
			if err != nil {
				return m.setError(fmt.Errorf("commit error: %w", err))
			}
			return m.Reload()
		}
	} else {
		err := COLUMNS[j].Action(&row, m.backend)
		if err != nil {
			return m.setError(fmt.Errorf("modify error: %w", err))
		}
	}
	return m, nil
}

// Row
// A Row represents a single row of the table and corresponds to a single interval in the Timewarrior database.
type Row struct {
	timew.Interval
	cells []cell
	date  time.Time
}

func NewRow(date time.Time) Row {
	cells := make([]cell, len(COLUMNS))
	interval := timew.Interval{} // dummy
	for i, column := range COLUMNS {
		cells[i] = newCell(column.Get(interval), column.Type, column.Width)
	}
	return Row{
		// OK for these to be nil because they'll be set before they're ever committed
		Interval: timew.Interval{
			Start: &timew.Datetime{},
			End:   &timew.Datetime{},
			Tags:  []string{},
		},
		cells: cells,
		date:  date,
	}
}

// Create a new Row from a Timewarrior interval.
func NewRowFromInterval(interval timew.Interval) Row {
	cells := make([]cell, len(COLUMNS))
	for i, column := range COLUMNS {
		cells[i] = newCell(column.Get(interval), column.Type, column.Width)
	}
	return Row{
		Interval: interval,
		cells:    cells,
		date:     interval.Start.Time,
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

	annotation := r.cells[3].Value()
	if len(annotation) > 0 {
		// Only way to annotate is to get all the data again and find the ID of
		// the new interval.
		intervals, err := backend.Export(r.date.Format("2006-01-02"))
		if err != nil {
			return fmt.Errorf("reading timewarrior datat: %w", err)
		}

		var id int
		for _, interval := range intervals {
			if r.Interval.Equal(interval) {
				id = interval.ID
			}
		}
		if id == 0 {
			return fmt.Errorf("cannot find id of newly created interval")
		}

		err = backend.Annotate(id, annotation)
		if err != nil {
			return fmt.Errorf("creating annotation: %w", err)
		}
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
	// Handle the case where the interval is open
	isIntervalOpen := false
	if r.Interval.End == nil {
		r.Interval.End = &timew.Datetime{}
		isIntervalOpen = true
	}
	err := r.setEndInInterval(r.date, r.cells[1].Value())
	if err != nil {
		return fmt.Errorf("setting end time: %w", err)
	}

	// Commit to Timewarrior
	if isIntervalOpen {
		stopTime := r.Interval.End.LocalString()
		err = backend.Stop(&stopTime)
	} else {
		err = backend.Modify(r.Interval.ID, "end", r.Interval.End.LocalString())
	}
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

func (r *Row) UpdateAnnotation(backend TimewarriorBackend) error {
	r.Interval.Annotation = r.cells[3].Value()

	// Commit to Timewarrior
	err := backend.Annotate(r.Interval.ID, r.Interval.Annotation)
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
	for _, cell := range r.cells[:len(r.cells)-1] {
		if cell.Err != nil {
			return false
		}
	}
	return true
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

	BaseStyle      lipgloss.Style
	HighlightStyle lipgloss.Style
	FocusStyle     lipgloss.Style
	highlight      bool
}

func (c cell) WithHighlight(highlight bool) cell {
	c.highlight = highlight
	return c
}

func (c cell) WithModel(m textinput.Model) cell {
	c.Model = m
	return c
}

func (c cell) Update(msg tea.Msg) (cell, tea.Cmd) {
	newModel, cmd := c.Model.Update(msg)
	return c.WithModel(newModel), cmd
}

func (c cell) View() string {
	c.PlaceholderStyle = c.BaseStyle.Faint(true)
	if c.highlight {
		c.TextStyle = c.HighlightStyle
		c.PlaceholderStyle = c.HighlightStyle
	} else if c.Focused() {
		c.TextStyle = c.FocusStyle
	} else {
		c.TextStyle = c.BaseStyle
	}
	return c.Model.View()
}

func newCell(value string, fieldType int, width int) cell {
	m := textinput.New()
	m.Prompt = ""
	if fieldType == FieldTime {
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
		// m.CharLimit = 40
		m.Width = width
		m.Placeholder = "none"
		m.Validate = func(text string) error {
			if len(text) == 0 {
				return fmt.Errorf("Must provide a tag")
			}
			return nil
		}
	}
	m.SetValue(value)
	m.Cursor.Blink = true
	m.Cursor.SetMode(curse.CursorBlink)
	return cell{
		Model:          m,
		BaseStyle:      BaseStyle,
		FocusStyle:     FocusStyle,
		HighlightStyle: HighlightStyle,
		highlight:      false,
	}
}
