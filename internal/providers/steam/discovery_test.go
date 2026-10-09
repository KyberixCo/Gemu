package steam

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverMultipleLibraries(t *testing.T) {
	root, extra := t.TempDir(), t.TempDir()
	for _, dir := range []string{filepath.Join(root, "steamapps"), filepath.Join(extra, "steamapps", "common", "Game")} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	folders := `"libraryfolders" { "0" { "path" "` + filepath.ToSlash(root) + `" } "1" { "path" "` + filepath.ToSlash(extra) + `" } }`
	if err := os.WriteFile(filepath.Join(root, "steamapps", "libraryfolders.vdf"), []byte(folders), 0600); err != nil {
		t.Fatal(err)
	}
	acf := `"AppState" { "appid" "42" "name" "Example" "installdir" "Game" "StateFlags" "4" }`
	if err := os.WriteFile(filepath.Join(extra, "steamapps", "appmanifest_42.acf"), []byte(acf), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "steamapps", "appmanifest_bad.acf"), []byte(`"AppState" {`), 0600); err != nil {
		t.Fatal(err)
	}
	games, diagnostics := Discover(root)
	if len(games) != 1 || !games[0].Installations[0].Installed || len(diagnostics) != 1 {
		t.Fatalf("games=%+v diagnostics=%v", games, diagnostics)
	}
}
func TestVDF(t *testing.T) {
	tree, err := parse("// comment\n\"libraryfolders\" {\"0\" \"C:\\\\Games\"}")
	if err != nil || tree["libraryfolders"].children["0"].value != `C:\Games` {
		t.Fatalf("%v %v", tree, err)
	}
	for _, s := range []string{`"x" {`, `"x"`, `}`, `"x" "unfinished`} {
		if _, err := parse(s); err == nil {
			t.Fatalf("accepted malformed VDF %q", s)
		}
	}
}
func TestAppID(t *testing.T) {
	for _, id := range []string{"", "0", "-1", "1?x", "4294967296"} {
		if ValidAppID(id) {
			t.Fatalf("accepted %s", id)
		}
	}
}
