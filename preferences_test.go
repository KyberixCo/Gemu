package main

import (
	"os"
	"path/filepath"
	"testing"

	"kyberix.com/gemu/internal/profiles"
)

func TestExistingProfileLibraryAndWebViewPreferencesMigrateTogether(t *testing.T) {
	dir := t.TempDir()
	id := "00112233445566778899aabbccddeeff"
	legacyDir := filepath.Join(dir, "profiles", id)
	if err := os.MkdirAll(legacyDir, 0700); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		filepath.Join(dir, "profiles.json"):      `[{"id":"00112233445566778899aabbccddeeff","name":"Existing player"}]`,
		filepath.Join(legacyDir, "library.json"): `{"version":1,"sources":[{"id":"steam","root":"C:/Steam","profileId":"steam-player"}],"games":[{"sourceId":"steam","game":{"ID":"steam:42","Title":"Existing game","Installations":[{"Provider":"steam","ExternalID":"42","Path":"C:/Steam/Game","Installed":true}]},"metadata":{"description":"Keep this","icon":"/library-assets/keep.img","manualArt":["icon"]},"player":{"profile":{"id":"steam-player"},"status":"ready","achievements":[{"id":"first","unlocked":true}]}}]}`,
	}
	for path, data := range files {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	app := &App{dir: dir, profiles: profiles.New(dir)}
	players, err := app.GetProfiles()
	if err != nil || len(players) != 1 || players[0].ID != id {
		t.Fatal(players, err)
	}
	if err = app.SelectProfile(id); err != nil {
		t.Fatal(err)
	}
	if _, err = app.LoadPreferences(id, map[string]string{"gemu.favorites": `["steam:42"]`, "gemu.feedback": `{"volume":35}`, "gemu.setup.v1": "complete", "gemu.management.v1": "{}", "keyboard.v1": "{}"}); err != nil {
		t.Fatal(err)
	}
	restarted := &App{dir: dir, profiles: profiles.New(dir)}
	if err = restarted.SelectProfile(id); err != nil {
		t.Fatal(err)
	}
	snapshot, err := restarted.GetLibrary()
	if err != nil || len(snapshot.Games) != 1 {
		t.Fatal(snapshot, err)
	}
	game := snapshot.Games[0]
	if game.Game.ID != "steam:42" || !game.Game.Installations[0].Installed || game.Metadata.Description != "Keep this" || game.Metadata.ManualArt[0] != "icon" || !game.Player.Achievements[0].Unlocked || snapshot.Sources[0].ProfileID != "steam-player" {
		t.Fatal(game)
	}
	prefs, err := restarted.LoadPreferences(id, nil)
	if err != nil || !prefs.LegacyImported || len(prefs.Values) != 5 || prefs.Values["gemu.favorites"] != `["steam:42"]` {
		t.Fatal(prefs, err)
	}
	for path, want := range files {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != want {
			t.Fatal("changed backup", path, err)
		}
	}
}

func TestSQLitePreferencesExplicitOwnerSurvivesPlayerSwitchAndRestart(t *testing.T) {
	dir := t.TempDir()
	app := &App{dir: dir, profiles: profiles.New(dir)}
	a, err := app.CreateProfile("A")
	if err != nil {
		t.Fatal(err)
	}
	b, err := app.CreateProfile("B")
	if err != nil {
		t.Fatal(err)
	}
	if err = app.SelectProfile(b.ID); err != nil {
		t.Fatal(err)
	}
	// A save started by A must retain its owner even when B is now active.
	if err = app.SavePreference(a.ID, "gemu.favorites", `["steam:1"]`); err != nil {
		t.Fatal(err)
	}
	if err = app.SavePreference(b.ID, "gemu.favorites", `["steam:2"]`); err != nil {
		t.Fatal(err)
	}
	restarted := &App{dir: dir, profiles: profiles.New(dir)}
	for id, want := range map[string]string{a.ID: `["steam:1"]`, b.ID: `["steam:2"]`} {
		got, err := restarted.LoadPreferences(id, map[string]string{})
		if err != nil || got.Values["gemu.favorites"] != want {
			t.Fatal(got, err)
		}
	}
	if err = app.SavePreference("unknown", "gemu.favorites", "[]"); err == nil {
		t.Fatal("unknown player accepted")
	}
	if err = app.SavePreference("", "gemu.favorites", "[]"); err == nil {
		t.Fatal("player favorites in global scope")
	}
	if _, err = os.Stat(filepath.Join(dir, "profiles", a.ID, "gemu.sqlite3")); !os.IsNotExist(err) {
		t.Fatal("created a separate database per player")
	}
	otherDir := t.TempDir()
	other := &App{dir: otherDir, profiles: profiles.New(otherDir)}
	if err = app.SavePreference("", "gemu.language.v1", "en"); err != nil {
		t.Fatal(err)
	}
	state, err := other.LoadPreferences("", nil)
	if err != nil || len(state.Values) != 0 {
		t.Fatal("environment data leaked", state, err)
	}
}
