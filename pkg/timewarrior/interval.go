package timewarrior

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const datetimeLayout = "20060102T150405Z"

// Struct corresponding to a single Timewarrior interval.
type Interval struct {
	ID    int       `json:"id"`
	Start *Datetime `json:"start"`
	End   *Datetime `json:"end"`
	Tags  []string  `json:"tags"`
}

// Returns true if the interval is closed (i.e. end datetime is defined).
func (interval Interval) IsClosed() bool {
	return interval.End != nil
}

// Returns true if the interval is open (i.e. end datetime is not defined).
func (interval Interval) IsOpen() bool {
	return interval.End == nil
}

// Returns a string representation of the interval suitable for the Timewarrior database file(s)
func (interval Interval) DatabaseString() string {
	if interval.End == nil {
		return fmt.Sprintf(
			"inc %s # %s",
			interval.Start.String(),
			tagsToDatabaseString(interval.Tags),
		)
	}
	return fmt.Sprintf(
		"inc %s - %s # %s",
		interval.Start.String(),
		interval.End.String(),
		tagsToDatabaseString(interval.Tags),
	)
}

// Return a new Interval where the start and end time locations are set to the local timezone.
func (interval Interval) Localize() Interval {
	out := Interval{
		ID:   interval.ID,
		Tags: interval.Tags,
	}
	if interval.Start != nil {
		start := interval.Start.Local()
		out.Start = &start
	}
	if interval.End != nil {
		end := interval.End.Local()
		out.End = &end
	}
	return out
}

// Returns a JSON-like string representation of the interval.
func (interval Interval) String() string {
	return fmt.Sprintf(
		"{\n\tID: %d\n\tStart: %s\n\tEnd: %s\n\tTags: %s\n}\n",
		interval.ID,
		interval.Start,
		interval.End,
		interval.Tags,
	)
}

// Returns tags as a slice of strings. Any strings with spaces in them are enclosed in single quotes.
func (interval Interval) GetTags() []string {
	out := make([]string, len(interval.Tags))
	for i, tag := range interval.Tags {
		if strings.Count(tag, " ") > 0 {
			out[i] = fmt.Sprintf("%s", tag)
		} else {
			out[i] = tag
		}
	}
	return out
}

// Returns true if the interval starts before the other interval.
func (interval Interval) StartsBefore(other Interval) bool {
	return int(interval.Start.Time.Sub(other.Start.Time)) < 0
}

// Returns true if the interval overlaps with the other interval, and the duration of the overlap.
func (interval Interval) Overlaps(other Interval) (bool, time.Duration) {
	if interval.End == nil && other.End == nil {
		return overlapsBothOpen(interval, other)
	}

	if interval.End == nil {
		return overlapsOneOpen(interval, other)
	}

	if other.End == nil {
		return overlapsOneOpen(other, interval)
	}

	// Both intervals have end times, check if they overlap
	return overlapsBothClosed(interval, other)
}

// Returns true if the interval contains the given date/time.
func (interval Interval) Contains(date time.Time) bool {
	var start, end time.Time
	zone, _ := date.Zone()
	if zone != "UTC" {
		zoneName, offset := time.Now().Zone()
		localZone := time.FixedZone(zoneName, offset)
		start = interval.Start.In(localZone)
		end = interval.End.In(localZone)
	} else {
		start = interval.Start.Time
		end = interval.End.Time
	}
	return date.Equal(start) || (date.After(start) && date.Before(end))
}

func overlapsBothOpen(interval Interval, other Interval) (bool, time.Duration) {
	// Both intervals are open-ended -> they overlap.
	// Figure out how much overlap they have as of now.
	now := time.Now()
	if interval.Start.Time.Before(other.Start.Time) {
		return true, now.Sub(other.Start.Time)
	}
	return true, now.Sub(interval.Start.Time)
}

func overlapsOneOpen(openInterval Interval, closedInterval Interval) (bool, time.Duration) {
	// First interval is open-ended -> check if the second interval ends *after* the first starts
	if closedInterval.End.After(openInterval.Start.Time) {
		if closedInterval.Start.After(openInterval.Start.Time) {
			return true, closedInterval.End.Time.Sub(closedInterval.Start.Time)
		}
		return true, closedInterval.End.Time.Sub(openInterval.Start.Time)
	}
	return false, time.Duration(0)
}

func overlapsBothClosed(interval1 Interval, interval2 Interval) (bool, time.Duration) {
	var t1, t2 time.Time
	//nolint: nestif // simplify later (if possible)
	if interval1.Start.Before(interval2.Start.Time) && interval2.Start.Before(interval1.End.Time) {
		t1 = interval2.Start.Time
		if interval2.End.Before(interval1.End.Time) {
			t2 = interval2.End.Time
		} else {
			t2 = interval1.End.Time
		}
		return true, t2.Sub(t1)
	} else if interval1.Start.Before(interval2.End.Time) && interval2.Start.Before(interval1.End.Time) {
		t1 = interval1.Start.Time
		if interval2.End.Before(interval1.End.Time) {
			t2 = interval2.End.Time
		} else {
			t2 = interval1.End.Time
		}
		return true, t2.Sub(t1)
	}
	return false, 0
}

// Print a list of tags to a string, suitable for writing to an interval in the database
func tagsToDatabaseString(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	newTags := make([]string, len(tags))
	for i, tag := range tags {
		newTags[i] = tagToDatabaseString(tag)
	}
	return strings.Join(newTags, " ")
}

func tagToDatabaseString(tag string) string {
	if strings.Count(tag, " ") > 0 {
		return fmt.Sprintf("\"%s\"", tag)
	}
	return tag
}

// Datetime is a thin wrapper over time.Time which adds helper methods to assist
// with serializing/deserializing times to/from Timewarrior.
type Datetime struct {
	time.Time
}

// Construct a new Datetime from a string with the format used interanlly by Timewarrior (i.e. "20060102T150405Z")
func NewDatetimeFromString(s string) (Datetime, error) {
	parsedTime, err := time.Parse(datetimeLayout, s)
	if err != nil {
		return Datetime{}, err
	}
	return Datetime{
		parsedTime,
	}, nil
}

func (t *Datetime) UnmarshalJSON(data []byte) error {
	// Define the layout string based on the input format
	layout := "20060102T150405Z"

	// Parse the string into a time.Time struct
	parsedTime, err := time.Parse(`"`+layout+`"`, string(data))
	if err != nil {
		return err
	}

	// Set the parsed time to the CustomTime field
	t.Time = parsedTime
	return nil
}

func (t Datetime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time)
}

// Return a new Datetime with the location set to local time.
func (t Datetime) Local() Datetime {
	return Datetime{t.Time.Local()}
}

// Return a string representation of the datetime.
func (t Datetime) String() string {
	return t.Time.Format(datetimeLayout)
}

// Return a string representation of the date (YYYY-mm-dd).
func (t Datetime) DateString() string {
	return t.Time.Format("2006-01-02")
}

// Return a string representation of the time (HH:MM).
func (t Datetime) TimeString() string {
	return t.Time.Format("15:04")
}

// Return a string representation of the datetime in the current timezone.
func (t Datetime) LocalString() string {
	return t.Time.Local().Format("20060102T150405")
}

// Return a string representation of the date (YYYY-mm-dd) in the current timezone.
func (t Datetime) LocalDateString() string {
	return t.Time.Local().Format("2006-01-02")
}

// Return a string representation of the time (HH:MM) in the current timezone.
func (t Datetime) LocalTimeString() string {
	return t.Time.Local().Format("15:04")
}
