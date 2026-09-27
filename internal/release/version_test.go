package release

import "testing"

func TestNext(t *testing.T) {
	cases := []struct {
		name string
		tags []string
		bump string
		want string
	}{
		{"REL-01 no tags, patch", nil, "patch", "v1.0.0"},
		{"REL-01 no tags, minor", nil, "minor", "v1.0.0"},
		{"REL-01 no tags, major", nil, "major", "v1.0.0"},
		{"REL-01 only non-semver tags", []string{"latest", "v1.0", "1.2.3"}, "minor", "v1.0.0"},
		{"REL-02 patch", []string{"v1.2.3"}, "patch", "v1.2.4"},
		{"REL-03 minor", []string{"v1.2.3"}, "minor", "v1.3.0"},
		{"REL-04 major", []string{"v1.2.3"}, "major", "v2.0.0"},
		{"REL-05 numeric order", []string{"v1.9.0", "v1.10.0", "v1.2.0"}, "patch", "v1.10.1"},
		{"REL-05 numeric order across major", []string{"v10.0.0", "v9.9.9"}, "patch", "v10.0.1"},
		{"REL-06 prerelease ignored", []string{"v1.0.0", "v1.0.0-beta", "v2.0.0-rc1"}, "patch", "v1.0.1"},
		{"REL-06 malformed ignored", []string{"v1.0.0", "v01.2.3", "v1.2", "v1.2.3.4", "x1.5.0", " v3.0.0"}, "patch", "v1.0.1"},
		{"blank lines ignored", []string{"", "v1.0.0", ""}, "minor", "v1.1.0"},
	}
	for _, c := range cases {
		got, err := Next(c.tags, c.bump)
		if err != nil || got != c.want {
			t.Errorf("%s: Next(%q, %q) = %q, %v want %q", c.name, c.tags, c.bump, got, err, c.want)
		}
	}
}

// REL-07: an unknown bump is an error, even with no tags.
func TestNextInvalidBump(t *testing.T) {
	for _, bump := range []string{"", "Patch", "build", "1"} {
		for _, tags := range [][]string{nil, {"v1.2.3"}} {
			if got, err := Next(tags, bump); err == nil {
				t.Errorf("Next(%q, %q) = %q, want error", tags, bump, got)
			}
		}
	}
}
