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

type TimecardData struct {
	data     map[string]timecardCol
	rowNames []string
	colNames []time.Time
}
type timecardCol = map[time.Time]time.Duration

// Run the timecard report
func Run(tw *timew.Report, options TimecardOptions) (string, error) {
	// lipgloss.SetColorProfile(termenv.Ascii)
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
	// filter intervals (if needed)
	var intervals []timew.Interval
	var err error
	if len(filters) > 0 {
		intervals, err = filterIntervals(tw.Intervals, filters)
		if err != nil {
			return TimecardData{}, fmt.Errorf("filtering intervals: %w", err)
		}
	} else {
		intervals = tw.Intervals
	}

	data := TimecardData{
		data: make(map[string]map[time.Time]time.Duration),
	}

	rowNames := []string{}
	colNames := []time.Time{}
	for _, interval := range intervals {
		dateCur := interval.Start.Truncate(Day)
		dateEnd := interval.End.Truncate(Day)
		for dateCur.Compare(dateEnd) <= 0 {
			overlapStart := maxTime(dateCur, interval.Start.Time)
			overlapEnd := minTime(dateCur.Add(Day), interval.End.Time)
			if overlapStart.Before(overlapEnd) {
				duration := overlapEnd.Sub(overlapStart)
				for _, tag := range interval.Tags {
					date := overlapStart.Truncate(Day)
					data.Add(tag, date, duration)
					if !slices.Contains(rowNames, tag) {
						rowNames = append(rowNames, tag)
					}
					if !slices.Contains(colNames, date) {
						colNames = append(colNames, date)
					}
				}
			}
			dateCur = dateCur.Add(Day)
		}
	}

	slices.Sort(rowNames)
	slices.SortFunc(colNames, func(a, b time.Time) int { return a.Compare(b) })
	data.rowNames = rowNames
	data.colNames = colNames

	if len(data.rowNames) == 0 && len(data.colNames) == 0 {
		return data, fmt.Errorf("no data in range %s - %s", tw.Config["temp.report.start"], tw.Config["temp.report.end"])
	}

	return data, nil
}

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
	return len(td.rowNames)
}

func (td TimecardData) Columns() int {
	return len(td.colNames) + 1
}

func (td TimecardData) At(row, cell int) string {
	if cell == 0 {
		return td.rowNames[row]
	}
	val, err := td.Get(td.rowNames[row], td.colNames[cell-1])
	if err != nil {
		return ""
	}
	return formatDurationDecimal(val)
}

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
	colNames := make([]string, len(td.colNames))
	for i, col := range td.colNames {
		colNames[i] = col.Format(DayFormat)
	}
	t := tableFormatter.New().
		Data(td).
		Headers(append([]string{"Tag"}, colNames...)...).
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == td.Rows() && col >= len(td.rowNames):
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

func formatDurationDecimal(d time.Duration) string {
	dstr := fmt.Sprintf("%g", d.Round(time.Hour/4).Hours())
	if dstr == "0" {
		return ""
	}
	return dstr
}
