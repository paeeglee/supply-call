// Package release computes the next version tag for the release workflow.
package release

import (
	"fmt"
	"regexp"
	"strconv"
)

// First is the version published when the repo has no version tag yet.
const First = "v1.0.0"

var tagRE = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

// Next returns the tag that follows the highest vMAJOR.MINOR.PATCH tag in
// tags, bumped by "patch", "minor" or "major". Tags in any other format are
// ignored; with none left it returns First.
func Next(tags []string, bump string) (string, error) {
	if bump != "patch" && bump != "minor" && bump != "major" {
		return "", fmt.Errorf("bump inválido %q: use patch, minor ou major", bump)
	}
	var best [3]int
	found := false
	for _, tag := range tags {
		m := tagRE.FindStringSubmatch(tag)
		if m == nil {
			continue
		}
		var v [3]int
		for i := range v {
			n, err := strconv.Atoi(m[i+1])
			if err != nil { // out of int range
				return "", fmt.Errorf("tag %q: %w", tag, err)
			}
			v[i] = n
		}
		if !found || less(best, v) {
			best, found = v, true
		}
	}
	if !found {
		return First, nil
	}
	switch bump {
	case "major":
		best = [3]int{best[0] + 1, 0, 0}
	case "minor":
		best = [3]int{best[0], best[1] + 1, 0}
	default:
		best[2]++
	}
	return fmt.Sprintf("v%d.%d.%d", best[0], best[1], best[2]), nil
}

func less(a, b [3]int) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}
