package timecard

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	timew "github.com/kgoettler/twe/pkg/timewarrior"

	"github.com/charmbracelet/lipgloss"
	tableFormatter "github.com/charmbracelet/lipgloss/table"
)

var (
	DayFormat = "Mon 01/02"
	Day       = time.Hour * 24
)

type TimecardOptions struct {
	Filters      []string
	Groups       []string
	OutputFormat string
	InputFile    string
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
}
type timecardCol = map[time.Time]time.Duration

// Run the timecard report
func Run(tw *timew.Report, options TimecardOptions) (string, error) {
	var err error

	data, err := NewTimecardData(tw, options.Filters)
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

func NewTimecardData(tw *timew.Report, filters []string) (TimecardData, error) {
	// Localize intervals
	intervals := localizeIntervals(tw.Intervals)

	// Filter intervals
	var err error
	if len(filters) > 0 {
		intervals, err = filterIntervals(intervals, filters)
		if err != nil {
			return TimecardData{}, fmt.Errorf("filtering intervals: %w", err)
		}
	}

	data := TimecardData{
		data: make(map[string]map[time.Time]time.Duration),
	}

	rows := []string{}
	columns := []time.Time{}
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
				duration := overlapEnd.Sub(overlapStart)
				for _, tag := range interval.Tags {
					data.Add(tag, dateCur, duration)
					if !slices.Contains(rows, tag) {
						rows = append(rows, tag)
					}
					if !slices.Contains(columns, dateCur) {
						columns = append(columns, dateCur)
					}
				}
			}
			dateCur = dateCur.Add(Day)
		}
	}

	slices.Sort(rows)
	slices.SortFunc(columns, func(a, b time.Time) int { return a.Compare(b) })
	data.rows = rows
	data.columns = columns

	if len(data.rows) == 0 && len(data.columns) == 0 {
		return data, fmt.Errorf("no data in range %s - %s", tw.Config["temp.report.start"], tw.Config["temp.report.end"])
	}

	return data, nil
}

// Add time for the given tag + date
func (td TimecardData) Add(tag string, date time.Time, duration time.Duration) time.Duration {
	_, ok := td.data[tag]
	if !ok {
		td.data[tag] = make(map[time.Time]time.Duration)
	}
	_, ok = td.data[tag][date]
	if !ok {
		td.data[tag][date] = 0
	}
	td.data[tag][date] += duration

	// TODO: update rowNames and colNames (if necessary)
	return td.data[tag][date]
}

func (td TimecardData) Rows() int {
	return len(td.rows)
}

func (td TimecardData) Columns() int {
	return len(td.columns) + 1
}

func (td TimecardData) At(row, cell int) string {
	if cell == 0 {
		return td.rows[row]
	}
	val, err := td.Get(td.rows[row], td.columns[cell-1])
	if err != nil {
		return ""
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
	colNames := make([]string, len(td.columns))
	for i, col := range td.columns {
		colNames[i] = col.Format(DayFormat)
	}
	t := tableFormatter.New().
		Data(td).
		Headers(append([]string{"Tag"}, colNames...)...).
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == td.Rows() && col >= len(td.rows):
				if len(td.At(row-1, col)) > 0 {
					return TotalRowStyle
				} else {
					return OddRowStyle
				}
			case row%2 == 0:
				return EvenRowStyle
			default:
				return OddRowStyle
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
	dstr := fmt.Sprintf("%g", d.Round(time.Hour/4).Hours())
	if dstr == "0" {
		return ""
	}
	return dstr
}
