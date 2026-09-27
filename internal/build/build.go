// Package build loads build order files and computes what to show at a given
// game time.
package build

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Races accepted in the race and vs keys.
var Races = []string{"Terran", "Protoss", "Zerg", "Random"}

// Build is one build order file. Unknown keys are ignored.
type Build struct {
	Name      string     `yaml:"name"`
	Race      string     `yaml:"race"`
	Vs        string     `yaml:"vs"`
	Steps     []Step     `yaml:"steps"`
	Reminders []Reminder `yaml:"reminders"`
	File      string     `yaml:"-"` // base file name
}

// Step is one timed action.
type Step struct {
	TimeText string `yaml:"time"`
	Action   string `yaml:"action"`
	Note     string `yaml:"note"`
	Time     int    `yaml:"-"` // seconds
}

// Reminder repeats every EverySeconds until Until.
type Reminder struct {
	Text         string `yaml:"text"`
	EverySeconds int    `yaml:"every_seconds"`
	UntilText    string `yaml:"until"`
	Until        int    `yaml:"-"` // seconds
}

// DisplayName is the name shown in menus: name, or the file name without extension.
func (b *Build) DisplayName() string {
	if b.Name != "" {
		return b.Name
	}
	return strings.TrimSuffix(b.File, filepath.Ext(b.File))
}

// Load reads and parses a build file.
func Load(path string) (*Build, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data, filepath.Base(path))
}

// Parse parses and validates a build. Steps come back sorted by time; steps
// with the same time keep their order from the file.
func Parse(data []byte, fileName string) (*Build, error) {
	var b Build
	if err := yaml.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("%s: YAML inválido: %v", fileName, err)
	}
	b.File = fileName
	fail := func(format string, args ...any) (*Build, error) {
		return nil, fmt.Errorf("%s: "+format, append([]any{fileName}, args...)...)
	}

	race, ok := normalizeRace(b.Race)
	if !ok {
		return fail("race %q inválida, use Terran, Protoss, Zerg ou Random", b.Race)
	}
	b.Race = race
	if b.Vs != "" {
		vs, ok := normalizeRace(b.Vs)
		if !ok {
			return fail("vs %q inválido, use Terran, Protoss, Zerg ou Random", b.Vs)
		}
		b.Vs = vs
	}
	if len(b.Steps) == 0 {
		return fail("a build não tem steps")
	}
	for i := range b.Steps {
		t, err := ParseTime(b.Steps[i].TimeText)
		if err != nil {
			return fail("passo %d (%s): %v", i+1, b.Steps[i].Action, err)
		}
		b.Steps[i].Time = t
	}
	for i := range b.Reminders {
		r := &b.Reminders[i]
		if r.EverySeconds <= 0 {
			return fail("lembrete %d (%s): every_seconds deve ser maior que 0", i+1, r.Text)
		}
		u, err := ParseTime(r.UntilText)
		if err != nil {
			return fail("lembrete %d (%s): until: %v", i+1, r.Text, err)
		}
		r.Until = u
	}
	sort.SliceStable(b.Steps, func(i, j int) bool { return b.Steps[i].Time < b.Steps[j].Time })
	return &b, nil
}

func normalizeRace(s string) (string, bool) {
	for _, r := range Races {
		if strings.EqualFold(strings.TrimSpace(s), r) {
			return r, true
		}
	}
	return "", false
}
