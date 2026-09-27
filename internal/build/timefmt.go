package build

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var timeRe = regexp.MustCompile(`^(\d{1,2}):([0-5]\d)$`)

// ParseTime converts "m:ss" (or "mm:ss") into seconds.
func ParseTime(s string) (int, error) {
	m := timeRe.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, fmt.Errorf("tempo %q inválido, use m:ss (ex.: 1:06)", s)
	}
	min, _ := strconv.Atoi(m[1])
	sec, _ := strconv.Atoi(m[2])
	return min*60 + sec, nil
}

// FormatTime renders seconds as "m:ss". Negative values render as "0:00".
func FormatTime(sec int) string {
	if sec < 0 {
		sec = 0
	}
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}

// Countdown converts a remaining duration in seconds into whole seconds,
// rounding up so "0:01" is shown until the moment is reached.
func Countdown(remaining float64) int {
	if remaining <= 0 {
		return 0
	}
	return int(math.Ceil(remaining))
}
