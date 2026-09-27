package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const valid = "name: %s\nrace: Terran\nsteps: [{time: \"0:10\", action: A}]\n"

func TestScanMixedFolder(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "b_pvz.yaml", strings.Replace(valid, "%s", "PvZ Macro", 1))
	write(t, dir, "A_tvt.yml", strings.Replace(valid, "%s", "TvT Bio", 1))
	write(t, dir, "sem_nome.yml", strings.Replace(valid, "name: %s\n", "", 1))
	write(t, dir, "quebrado.yml", "name: [x\n")
	write(t, dir, "notas.txt", "ignore me")
	os.Mkdir(filepath.Join(dir, "sub.yml"), 0o755) // directories are ignored

	entries := Scan(dir)
	var got []string
	for _, e := range entries {
		got = append(got, e.File+"="+e.Label())
	}
	want := []string{
		"A_tvt.yml=TvT Bio",
		"b_pvz.yaml=PvZ Macro",
		"quebrado.yml=quebrado.yml (erro)",
		"sem_nome.yml=sem_nome",
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("entries = %v\nwant      %v", got, want)
	}
	for _, e := range entries {
		switch e.File {
		case "quebrado.yml":
			if e.Err == nil || e.Build != nil || !strings.Contains(e.Err.Error(), "quebrado.yml") {
				t.Errorf("invalid entry = %+v", e)
			}
		default:
			if e.Err != nil || e.Build == nil {
				t.Errorf("valid entry = %+v", e)
			}
		}
	}
}

func TestScanEmptyAndMissingFolder(t *testing.T) {
	if got := Scan(t.TempDir()); len(got) != 0 {
		t.Errorf("empty folder = %v", got)
	}
	if got := Scan(filepath.Join(t.TempDir(), "nao-existe")); len(got) != 0 {
		t.Errorf("missing folder = %v", got)
	}
}
