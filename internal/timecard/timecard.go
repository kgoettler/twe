package timecard

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/kgoettler/twe/internal/styles"
	timew "github.com/kgoettler/twe/pkg/timewarrior"

	"github.com/charmbracelet/lipgloss"
	tableFormatter "github.com/charmbracelet/lipgloss/table"
)

var (
	DayFormat = "Mon 01/02"
	Day       = time.Hour * 24
	EmptyChar = "-"
)

type TimecardOptions struct {
	Filters      []string
	Groups       []string
	OutputFormat string
	InputFile    string

	// If true, includes a column for tag totals
	IncludeTotalCol bool

	// If true, includes a row for daily totals
	IncludeTotalRow bool

	// increment (in minutes) up to which each duration will be rounded.
	Increment int
}

// TimecardData contains tabular timecard data.
// The internal `data` field is a nested map with the structure:
//
//	{
//	  "row": {
//	     "column": time.Duration
//	  }
//	}
type TimecardData struct {
	data    map[string]timecardCol
	rows    []string
	columns []time.Time

	// Contains daily totals of hours logged
	totals timecardCol

	// Contains tag-wise totals of hours logged
	rowTotals map[string]time.Duration

	// Options
	options TimecardOptions

	round func(d time.Duration) time.Duration
}
type timecardCol = map[time.Time]time.Duration

// Run the timecard report
func Run(tw *timew.Report, options TimecardOptions) (string, error) {
	var err error

	data, err := NewTimecardData(tw, options)
	if err != nil {
		return "", fmt.Errorf("generating data: %w", err)
	}

	// Get table format
	var dataString string
	switch options.OutputFormat {
	case "table":
		dataString, err = data.StringTable()
	default:
		return "", fmt.Errorf("unrecognized table format: %s", options.OutputFormat)
	}
	if err != nil {
		return "", err
	}
	return dataString, nil
}

func NewTimecardData(tw *timew.Report, options TimecardOptions) (TimecardData, error) {
	// Localize intervals
	intervals := localizeIntervals(tw.Intervals)

	// Filter intervals
	var err error
	if len(options.Filters) > 0 {
		intervals, err = filterIntervals(intervals, options.Filters)
		if err != nil {
			return TimecardData{}, fmt.Errorf("filtering intervals: %w", err)
		}
	}

	data := TimecardData{
		data:      make(map[string]map[time.Time]time.Duration),
		totals:    make(map[time.Time]time.Duration),
		rowTotals: make(map[string]time.Duration),
		options:   options,
		round:     getRoundingFunc(options.Increment),
	}

	// rows := []string{}
	// columns := []time.Time{}
	for _, interval := range intervals {
		// Loop over each day this interval overlaps with, calculate the
		// amount of overlap on that day, and add it to the data structure.
		var iStart, iEnd, dateCur, dateEnd time.Time
		iStart = interval.Start.Time
		if interval.End != nil {
			iEnd = interval.End.Time
		} else {
			iEnd = time.Now()
		}
		dateCur = midnightLocal(iStart)
		dateEnd = midnightLocal(iEnd)
		for dateCur.Compare(dateEnd) <= 0 {
			// overlapStart is midnight of the current day or the interval start time, whichever is later
			// overlapEnd is midnight of the next day or the interval end time, whichever is earlier
			overlapStart := maxTime(dateCur, iStart.In(time.Local))
			overlapEnd := minTime(dateCur.Add(Day), iEnd.In(time.Local))
			// overlapStart will be before overlapEnd until we've reached a day
			// with which the interval no longer overlaps.
			if overlapStart.Before(overlapEnd) {
				duration := data.round(overlapEnd.Sub(overlapStart))
				data.AddDateTotal(dateCur, duration)
				for _, tag := range interval.Tags {
					data.Add(tag, dateCur, duration)
					data.AddTagTotal(tag, duration)
				}
			}
			dateCur = dateCur.Add(Day)
		}
	}

	slices.Sort(data.rows)
	slices.SortFunc(data.columns, func(a, b time.Time) int { return a.Compare(b) })

	if len(data.rows) == 0 && len(data.columns) == 0 {
		return data, fmt.Errorf("no data in range %s - %s", tw.Config["temp.report.start"], tw.Config["temp.report.end"])
	}

	return data, nil
}

// Add time for the given tag + date
func (td *TimecardData) Add(tag string, date time.Time, duration time.Duration) time.Duration {
	_, ok := td.data[tag]
	if !ok {
		td.data[tag] = make(map[time.Time]time.Duration)
		td.rows = append(td.rows, tag)
	}
	_, ok = td.data[tag][date]
	if !ok {
		td.data[tag][date] = 0
	}
	td.data[tag][date] += duration

	if !slices.Contains(td.columns, date) {
		td.columns = append(td.columns, date)
	}

	return td.data[tag][date]
}

func (td *TimecardData) AddTagTotal(tag string, duration time.Duration) {
	_, ok := td.rowTotals[tag]
	if !ok {
		td.rowTotals[tag] = 0
	}
	td.rowTotals[tag] += duration
}

func (td *TimecardData) AddDateTotal(date time.Time, duration time.Duration) {
	_, ok := td.totals[date]
	if !ok {
		td.totals[date] = duration
	} else {
		td.totals[date] += duration
	}
}

func (td TimecardData) Rows() int {
	// Tag rows + total row
	var extra int
	if td.options.IncludeTotalRow {
		extra++
	}
	return len(td.rows) + extra
}

func (td TimecardData) Columns() int {
	// Header column + date columns + total column
	var extra int
	if td.options.IncludeTotalCol {
		extra++
	}
	return 1 + len(td.columns) + extra
}

func (td TimecardData) atHeaderColumn(row int) string {
	rowN := td.Rows() - 1
	if row == rowN && td.options.IncludeTotalRow {
		return "TOTAL"
	}
	return td.rows[row]
}

func (td TimecardData) atTotalsColumn(row int) string {
	if row == td.Rows()-1 && td.options.IncludeTotalRow {
		return EmptyChar
	}
	rowName := td.rows[row]
	return formatDurationDecimal(td.rowTotals[rowName])
}

func (td TimecardData) atTotalsRow(cell int) string {
	if cell == td.Columns()-1 && td.options.IncludeTotalCol {
		return EmptyChar
	}
	return formatDurationDecimal(td.totals[td.columns[cell-1]])
}

func (td TimecardData) At(row, cell int) string {
	col0 := 0
	rowN := td.Rows() - 1
	colN := td.Columns() - 1

	if cell == col0 {
		return td.atHeaderColumn(row)
	}

	if row == rowN && td.options.IncludeTotalRow {
		return td.atTotalsRow(cell)
	}

	if cell == colN && td.options.IncludeTotalCol {
		return td.atTotalsColumn(row)
	}

	val, err := td.Get(td.rows[row], td.columns[cell-1])
	if err != nil {
		return EmptyChar
	}
	return formatDurationDecimal(val)
}

// Get hours logged for given tag on the given date.
func (td TimecardData) Get(tag string, date time.Time) (time.Duration, error) {
	col, ok := td.data[tag]
	if !ok {
		return 0, fmt.Errorf("[%s,:] undefined", tag)
	}
	val, ok := col[date]
	if !ok {
		return 0, fmt.Errorf("[%s,%s] undefined", tag, date)
	}
	return val, nil
}

func (td TimecardData) StringTable() (string, error) {
	// Format table
	var extra int
	if td.options.IncludeTotalCol {
		extra++
	}
	colNames := make([]string, len(td.columns)+extra)
	for i, col := range td.columns {
		colNames[i] = col.Format(DayFormat)
	}
	if td.options.IncludeTotalCol {
		colNames[len(colNames)-1] = "TOTAL"
	}
	t := tableFormatter.New().
		Data(td).
		Headers(append([]string{"Tag"}, colNames...)...).
		Border(lipgloss.NormalBorder()).
		BorderStyle(styles.BorderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == -1 && col == (td.Columns()-1) && td.options.IncludeTotalCol:
				return styles.TotalRowStyle
			case row == -1:
				return styles.HeaderStyle
			case row == (td.Rows()-1) && td.options.IncludeTotalRow:
				return styles.TotalRowStyle
			case col == (td.Columns()-1) && td.options.IncludeTotalCol:
				return styles.TotalRowStyle
			case row%2 == 0:
				return styles.EvenRowStyle
			default:
				return styles.OddRowStyle
			}
		})
	ts := t.Render()
	return ts, nil
}

func (td TimecardData) String() string {
	var builder strings.Builder
	for tag, col := range td.data {
		for date, duration := range col {
			builder.WriteString(fmt.Sprintf("%s,%s = %.2f\n", tag, date.Format("Jan 02"), duration.Hours()))
		}
	}
	return builder.String()
}

// Localize the start and end times in the provided intervals.
func localizeIntervals(intervals []timew.Interval) []timew.Interval {
	out := make([]timew.Interval, len(intervals))
	for i, interval := range intervals {
		out[i] = interval.Localize()
	}
	return out
}

// Filter out intervals without tags matching at least one of the provided filter regexps.
func filterIntervals(intervals []timew.Interval, filters []string) ([]timew.Interval, error) {
	out := []timew.Interval{}
	filterPatterns := make([]*regexp.Regexp, len(filters))
	for i, filter := range filters {
		p, err := regexp.Compile(filter)
		if err != nil {
			return out, fmt.Errorf("filter %s failed to compile as regex: %w", filter, err)
		}
		filterPatterns[i] = p
	}

	// Apply filters
	for _, interval := range intervals {
		for _, tag := range interval.Tags {
			if matchAny(tag, filterPatterns) {
				out = append(out, interval)
				break
			}
		}
	}
	return out, nil
}

func matchAny(s string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(s) {
			return true
		}
	}
	return false
}

// Helper function to get the maximum of two time values
func maxTime(t1, t2 time.Time) time.Time {
	if t1.After(t2) {
		return t1
	}
	return t2
}

// Helper function to get the minimum of two time values
func minTime(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t1
	}
	return t2
}

// Return midnight for the provided date + location
func midnightLocal(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func formatDurationDecimal(d time.Duration) string {
	dstr := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", d.Hours()), "0"), ".")
	if dstr == "0" {
		return EmptyChar
	}
	return dstr
}

func getRoundingFunc(increment int) func(time.Duration) time.Duration {
	if increment == 0 {
		return func(d time.Duration) time.Duration {
			return d
		}
	}
	m := time.Minute * time.Duration(increment)
	return func(d time.Duration) time.Duration {
		if m <= 0 {
			return d
		}
		// Calculate the remainder (modulo)
		remainder := d % m
		if remainder == 0 {
			return d
		}
		// Add the difference to round up
		return d + (m - remainder)
	}
}
