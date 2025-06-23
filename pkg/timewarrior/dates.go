package timewarrior

import (
	"fmt"
	"regexp"
	"time"
)

var timeFormats = []string{
	"2006-01-02",
	"20060102",
}

type DateFunc = func(time.Time, string) (time.Time, error)

type datePattern struct {
	pattern string
	fn      DateFunc
}

var weekdayMap = map[string]time.Weekday{
	"mon":       time.Monday,
	"monday":    time.Monday,
	"tue":       time.Tuesday,
	"tuesday":   time.Tuesday,
	"wed":       time.Wednesday,
	"wednesday": time.Wednesday,
	"thu":       time.Thursday,
	"thursday":  time.Thursday,
	"fri":       time.Friday,
	"friday":    time.Friday,
	"sat":       time.Saturday,
	"saturday":  time.Saturday,
	"sun":       time.Sunday,
	"sunday":    time.Sunday,
}

var datePatterns = []datePattern{
	{
		`^(mon|monday|tue|tuesday|wed|wednesday|thu|thursday|fri|friday|sat|saturday|sun|sunday)$`,
		func(now time.Time, match string) (time.Time, error) {
			day := weekdayMap[match]
			return now.AddDate(0, 0, -int(now.Weekday()-day)), nil
		},
	},
	{
		`^(today|now)$`,
		func(now time.Time, _ string) (time.Time, error) {
			return now, nil
		},
	},
	{
		`^(yesterday)$`,
		func(now time.Time, _ string) (time.Time, error) {
			return now.Add(-24 * time.Hour), nil
		},
	},
	{
		`^(tomorrow)$`,
		func(now time.Time, _ string) (time.Time, error) {
			return now.Add(24 * time.Hour), nil
		},
	},
	{
		`^\d{4}-\d{2}-\d{2}$`,
		func(_ time.Time, match string) (time.Time, error) {
			return time.Parse("2006-01-02", match)
		},
	},
	{
		`^\d{8}$`,
		func(_ time.Time, match string) (time.Time, error) {
			return time.Parse("20060102", match)
		},
	},
}

var dateFormats = func() map[*regexp.Regexp]DateFunc {
	m := make(map[*regexp.Regexp]DateFunc)
	for _, pat := range datePatterns {
		m[regexp.MustCompile(pat.pattern)] = pat.fn
	}
	return m
}()

func ConvertDateStringToTime(now time.Time, dateString string) (time.Time, error) {
	for regex, dateFunc := range dateFormats {
		m := regex.FindString(dateString)
		if m != "" {
			return dateFunc(now, dateString)
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date: %s", dateString)
}
