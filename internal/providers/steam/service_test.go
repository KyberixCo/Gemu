package steam

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectSnapshot(t *testing.T) {
	root := t.TempDir()
	apps := filepath.Join(root, "steamapps")
	if err := os.MkdirAll(filepath.Join(apps, "common", "Ready"), 0755); err != nil {
		t.Fatal(err)
	}
	for id, manifest := range map[string]string{
		"42":     `"AppState" { "appid" "42" "name" "Ready" "installdir" "Ready" "StateFlags" "4" }`,
		"43":     `"AppState" { "appid" "43" "name" "Missing" "installdir" "Missing" "StateFlags" "4" }`,
		"44":     `"AppState" { "appid" "44" "name" "Downloading" "installdir" "Ready" "StateFlags" "2" }`,
		"45":     `"AppState" { "appid" "45" "name" "Invalid" "installdir" "../escape" "StateFlags" "4" }`,
		"228980": `"AppState" { "appid" "228980" "name" "Steamworks Common Redistributables" "installdir" "Ready" "StateFlags" "4" }`,
	} {
		if err := os.WriteFile(filepath.Join(apps, "appmanifest_"+id+".acf"), []byte(manifest), 0600); err != nil {
			t.Fatal(err)
		}
	}
	report, err := Inspect(root)
	if err != nil || len(report.Games) != 3 || report.InstalledCount != 1 || len(report.Diagnostics) != 1 {
		t.Fatalf("report=%+v err=%v", report, err)
	}
}

func TestInspectInvalidAndEmptyLibrary(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{root, "relative"} {
		if _, err := Inspect(path); err == nil {
			t.Fatalf("accepted invalid library %q", path)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "steamapps"), 0755); err != nil {
		t.Fatal(err)
	}
	report, err := Inspect(root)
	if err != nil || report.Games == nil || report.Diagnostics == nil || report.InstalledCount != 0 {
		t.Fatalf("report=%+v err=%v", report, err)
	}
	if detected, err := findRoot([]string{"", "relative", filepath.Join(root, "missing"), root}); err != nil || detected != root {
		t.Fatalf("detected=%q err=%v", detected, err)
	}
}

func TestUnavailableLibraryKeepsHealthyGames(t *testing.T) {
	root := t.TempDir()
	apps := filepath.Join(root, "steamapps")
	if err := os.MkdirAll(filepath.Join(apps, "common", "Game"), 0755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"libraryfolders.vdf": `"libraryfolders" { "0" { "path" "` + filepath.ToSlash(root) + `" } "1" { "path" "` + filepath.ToSlash(filepath.Join(root, "offline")) + `" } "2" { "path" "relative" } }`,
		"appmanifest_42.acf": `"AppState" { "appid" "42" "name" "Example" "installdir" "Game" "StateFlags" "4" }`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(apps, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	report, err := Inspect(root)
	if err != nil || len(report.Games) != 1 || report.InstalledCount != 1 || len(report.Diagnostics) != 2 {
		t.Fatalf("report=%+v err=%v", report, err)
	}
}
