package timewarrior

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

const (
	hoursInDay = 24
)

// Report contains the data passed to a Timewarrior report via the [Extension API].
// [Extension API]: https://timewarrior.net/docs/api/
type Report struct {
	Config    map[string]string
	Intervals []Interval
}

// Return a new Report from an io.Reader. This is the primary mechanism for
// creating Reports; extensions should pass os.Stdin as the input reader to this
// constructor.
func NewReport(reader io.Reader) (*Report, error) {
	scanner := bufio.NewScanner(reader)

	// Read config + intervals
	configPattern := regexp.MustCompile(`([a-z\.]*): (.*)`)
	jsonPattern := regexp.MustCompile(`({.*})`)
	intervals := make([]Interval, 0)
	config := map[string]string{}
	for scanner.Scan() {
		line := scanner.Text()
		if m := configPattern.FindStringSubmatch(line); len(m) > 0 {
			// Config line
			config[m[1]] = m[2]
		} else if m := jsonPattern.FindStringSubmatch(line); len(m) > 0 {
			// JSON line
			// Strip any trailing comma
			cleanLine := strings.Trim(line, ",")

			// Parse to JSON
			var tempInterval Interval
			err := json.Unmarshal([]byte(cleanLine), &tempInterval)
			if err != nil {
				return nil, err
			}

			// Remove singel quotes from tags
			for i, tag := range tempInterval.Tags {
				tempInterval.Tags[i] = strings.Trim(tag, "'")
			}

			// Save
			intervals = append(intervals, tempInterval)
		}
	}
	tw := Report{
		Config:    config,
		Intervals: intervals,
	}
	return &tw, nil
}

// Returns the last recorded interval in the report.
func (tw *Report) Last() (Datetime, error) {
	if len(tw.Intervals) == 0 {
		return Datetime{}, fmt.Errorf("report contains no intervals")
	}
	lastInterval := tw.Intervals[len(tw.Intervals)-1]
	if lastInterval.End != nil {
		return *lastInterval.End, nil
	}
	return *lastInterval.Start, nil
}

// Returns a slice containing the unique tags attached to intervals within the
// report.
func (tw *Report) GetUniqueTags() []string {
	// Extract unique tags and days
	tagSet := make(map[string]struct{})

	for _, interval := range tw.Intervals {
		for _, tag := range interval.Tags {
			tagSet[tag] = struct{}{}
		}
	}

	// Convert sets to sorted slices
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

// Returns `temp.report.start` as a Datetime. If it is not defined for the
// report, a non-nil error is returned.
func (tw *Report) GetStartDate() (Datetime, error) {
	return tw.getConfigAsDatetime("temp.report.start")
}

// Returns `temp.report.end` as a Datetime. If it is not defined for the report,
// a non-nil error is returned.
func (tw *Report) GetEndDate() (Datetime, error) {
	return tw.getConfigAsDatetime("temp.report.end")
}

// Returns `temp.report.start` and `temp.report.end` as Datetimes. If either is
// not defined for the report, a non-nil error is returned.
func (tw *Report) GetDateRange() (Datetime, Datetime, error) {
	ti, err := tw.GetStartDate()
	if err != nil {
		return Datetime{}, Datetime{}, err
	}
	tf, err := tw.GetEndDate()
	if err != nil {
		return Datetime{}, Datetime{}, err
	}
	return ti, tf, nil
}

func (tw *Report) getConfigAsDatetime(field string) (Datetime, error) {
	v, ok := tw.Config[field]
	if !ok {
		return Datetime{}, fmt.Errorf("%s not defined on report", field)
	}
	out, err := NewDatetimeFromString(v)
	if err != nil {
		return Datetime{}, err
	}
	return out, nil
}
