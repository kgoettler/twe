package timewarrior

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	HoursInDay = 24
)

type Report struct {
	Config    map[string]string
	Intervals []Interval
}

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

func (tw *Report) Last() string {
	if len(tw.Intervals) == 0 {
		return ""
	}

	lastInterval := tw.Intervals[len(tw.Intervals)-1]
	if lastInterval.End != nil {
		return lastInterval.End.String()
	}
	return lastInterval.Start.String()
}

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

func (tw *Report) GetStartDate() (Datetime, error) {
	return tw.getConfigAsDatetime("temp.report.start")
}

func (tw *Report) GetEndDate() (Datetime, error) {
	return tw.getConfigAsDatetime("temp.report.end")
}

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

// Get array of dates in which the report intervals are present
func (tw *Report) GetDates(zone *time.Location) ([]Interval, error) {
	// Use UTC for default zone
	if zone == nil {
		zone = time.UTC
	}

	// Get report bounds
	dateStart, dateEnd, err := tw.GetDateRange()
	if err != nil {
		return nil, err
	}

	// Allocate output
	days := make([]Interval, 0)

	// Loop
	currentTime := dateStart.Time
	for currentTime.Before(dateEnd.Time) || currentTime.Equal(dateStart.Time) {
		interval := Interval{
			Start: &Datetime{currentTime.In(zone)},
		}
		interval.End = &Datetime{interval.Start.Add(time.Hour * HoursInDay)}
		days = append(days, interval)
		currentTime = currentTime.Add(time.Hour * HoursInDay) // Add 24 hours to move to the next day
	}
	return days, nil
}

func (tw *Report) IsSingleWeek() bool {
	ti, tf, err := tw.GetDateRange()
	if err != nil {
		return false
	}

	dt := tf.Time.Sub(ti.Time)
	days := dt.Hours() / 24

	if days > 7 {
		return false
	}
	return true
}

func (tw *Report) JSONString() (string, error) {
	intervalBytes, err := json.Marshal(tw.Intervals)
	if err != nil {
		return "", err
	}
	return string(intervalBytes), nil
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
