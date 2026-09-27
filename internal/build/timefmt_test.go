package build

import "testing"

func TestParseTimeValid(t *testing.T) {
	cases := map[string]int{
		"0:00":  0,
		"0:35":  35,
		"1:06":  66,
		"7:00":  420,
		"10:00": 600,
		"12:59": 779,
		" 2:08": 128, // surrounding spaces are tolerated
	}
	for in, want := range cases {
		got, err := ParseTime(in)
		if err != nil {
			t.Errorf("ParseTime(%q) error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseTime(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestParseTimeInvalid(t *testing.T) {
	for _, in := range []string{"", "35", "1:5", "1:60", "-1:00", "abc", "1:0a", "1:005", "1:2:03", "100:00"} {
		if _, err := ParseTime(in); err == nil {
			t.Errorf("ParseTime(%q) accepted, want error", in)
		}
	}
}

func TestFormatTime(t *testing.T) {
	cases := map[int]string{0: "0:00", 5: "0:05", 66: "1:06", 600: "10:00", -3: "0:00"}
	for in, want := range cases {
		if got := FormatTime(in); got != want {
			t.Errorf("FormatTime(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestCountdownRoundsUp(t *testing.T) {
	cases := map[float64]int{5.0: 5, 4.01: 5, 4.0: 4, 0.2: 1, 0: 0, -1: 0}
	for in, want := range cases {
		if got := Countdown(in); got != want {
			t.Errorf("Countdown(%v) = %d, want %d", in, got, want)
		}
	}
}
