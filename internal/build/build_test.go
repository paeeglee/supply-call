package build

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadExampleKeepsValuesVerbatim(t *testing.T) {
	b, err := Load(filepath.Join("testdata", "terran_bio.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if b.Name != "Terran Bio + Cyclone + Medivac" || b.Race != "Terran" {
		t.Fatalf("name/race = %q/%q", b.Name, b.Race)
	}
	if len(b.Steps) != 22 {
		t.Fatalf("steps = %d, want 22", len(b.Steps))
	}
	s := b.Steps[1]
	if s.Time != 48 || s.Action != "Refinaria" || s.Note != "3 SCVs no gás quando terminar" {
		t.Errorf("step 1 = %+v", s)
	}
	if b.Steps[7].Action != "2º Command Center" || b.Steps[7].Time != 171 {
		t.Errorf("step 7 = %+v", b.Steps[7])
	}
	if len(b.Reminders) != 1 {
		t.Fatalf("reminders = %d", len(b.Reminders))
	}
	r := b.Reminders[0]
	if r.Text != "SCV!" || r.EverySeconds != 12 || r.Until != 420 {
		t.Errorf("reminder = %+v", r)
	}
	if b.File != "terran_bio.yml" {
		t.Errorf("File = %q", b.File)
	}
}

func TestParseSortsByTimeStable(t *testing.T) {
	b, err := Parse([]byte(`
name: x
race: Zerg
steps:
  - {time: "1:00", action: C}
  - {time: "0:10", action: A}
  - {time: "1:00", action: D}
  - {time: "0:30", action: B}
`), "x.yml")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, s := range b.Steps {
		got = append(got, s.Action)
	}
	if strings.Join(got, ",") != "A,B,C,D" {
		t.Errorf("order = %v, want A,B,C,D", got)
	}
}

func TestParseIgnoresUnknownKeys(t *testing.T) {
	b, err := Parse([]byte(`
name: x
race: Protoss
author: someone
replay_hints: {supply: 14}
steps:
  - {time: "0:20", action: Pylon, supply: 14, chrono: true}
reminders:
  - {text: "Probe!", every_seconds: 12, until: "5:00", sound: beep}
`), "x.yml")
	if err != nil {
		t.Fatalf("unknown keys must be ignored: %v", err)
	}
	if b.Steps[0].Action != "Pylon" || b.Reminders[0].Text != "Probe!" {
		t.Errorf("parsed = %+v", b)
	}
}

func TestParseBrokenYAMLNamesFile(t *testing.T) {
	_, err := Parse([]byte("name: [unclosed\nsteps: {"), "quebrado.yml")
	if err == nil || !strings.Contains(err.Error(), "quebrado.yml") {
		t.Fatalf("err = %v, want mention of file", err)
	}
}

func TestParseInvalidStepTime(t *testing.T) {
	_, err := Parse([]byte(`
race: Terran
steps:
  - {time: "0:35", action: Depot}
  - {time: "1:5", action: Barracks}
`), "b.yml")
	if err == nil {
		t.Fatal("want error")
	}
	for _, want := range []string{"b.yml", "passo 2", "1:5"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err %q missing %q", err, want)
		}
	}
}

func TestParseValidationErrors(t *testing.T) {
	cases := map[string]string{
		"invalid race":    "race: Human\nsteps: [{time: \"0:10\", action: A}]",
		"missing race":    "steps: [{time: \"0:10\", action: A}]",
		"invalid vs":      "race: Zerg\nvs: Human\nsteps: [{time: \"0:10\", action: A}]",
		"no steps":        "race: Zerg\nsteps: []",
		"every_seconds 0": "race: Zerg\nsteps: [{time: \"0:10\", action: A}]\nreminders: [{text: x, every_seconds: 0, until: \"1:00\"}]",
		"bad until":       "race: Zerg\nsteps: [{time: \"0:10\", action: A}]\nreminders: [{text: x, every_seconds: 5, until: \"1:0\"}]",
	}
	for name, src := range cases {
		if _, err := Parse([]byte(src), "v.yml"); err == nil {
			t.Errorf("%s: want error", name)
		} else if !strings.Contains(err.Error(), "v.yml") {
			t.Errorf("%s: err %q does not name the file", name, err)
		}
	}
}

func TestParseAcceptsAllRacesAndVs(t *testing.T) {
	for _, race := range []string{"Terran", "Protoss", "Zerg", "Random"} {
		src := "race: " + race + "\nvs: " + race + "\nsteps: [{time: \"0:10\", action: A}]"
		b, err := Parse([]byte(src), "r.yml")
		if err != nil {
			t.Errorf("%s: %v", race, err)
			continue
		}
		if b.Race != race || b.Vs != race {
			t.Errorf("race/vs = %q/%q, want %q", b.Race, b.Vs, race)
		}
	}
}

func TestDisplayNameFallsBackToFileName(t *testing.T) {
	b, err := Parse([]byte("race: Zerg\nsteps: [{time: \"0:10\", action: A}]"), "ling_bane.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if got := b.DisplayName(); got != "ling_bane" {
		t.Errorf("DisplayName = %q, want ling_bane", got)
	}
	b.Name = "Ling Bane"
	if got := b.DisplayName(); got != "Ling Bane" {
		t.Errorf("DisplayName = %q, want Ling Bane", got)
	}
}
