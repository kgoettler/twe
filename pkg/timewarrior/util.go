package timewarrior

import (
	"fmt"
	"math"
	"strings"
	"time"
)

var emptyChar = "-"

func FormatDurationTime(d time.Duration) string {
	h := int(math.Floor(d.Hours()))
	m := int(math.Floor((d - (time.Duration(h) * time.Hour)).Minutes()))
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}

func FormatDurationDecimal(d time.Duration) string {
	dstr := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", d.Hours()), "0"), ".")
	if dstr == "0" {
		return emptyChar
	}
	return dstr
}
