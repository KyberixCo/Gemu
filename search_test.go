package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/library"
	"kyberix.com/gemu/internal/profiles"
)

func TestSearchLibraryKeepsExplicitOwnerAfterPlayerSwitch(t *testing.T) {
	dir := t.TempDir()
	app := &App{dir: dir, profiles: profiles.New(dir)}
	players := []profiles.Profile{}
	for _, name := range []string{"Alpha", "Beta"} {
		p, err := app.CreateProfile(name)
		if err != nil {
			t.Fatal(err)
		}
		players = append(players, p)
		legacyDir, err := app.profiles.Directory(p.ID)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(legacyDir, 0700); err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(library.Snapshot{Version: 1, Games: []library.Entry{{SourceID: "steam", Game: domain.Game{ID: "same-id", Title: name + " private"}}}})
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(legacyDir, "library.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := app.SelectProfile(players[1].ID); err != nil {
		t.Fatal(err)
	}
	page, err := app.SearchLibrary(players[0].ID, library.SearchRequest{Query: "alpha"})
	if err != nil || page.Total != 1 || page.Hits[0].Title != "Alpha private" {
		t.Fatal(page, err)
	}
	page, err = app.SearchLibrary(players[0].ID, library.SearchRequest{Query: "beta"})
	if err != nil || page.Total != 0 {
		t.Fatal("cross-player search", page, err)
	}
	if _, err = app.SearchLibrary("../unknown", library.SearchRequest{Query: "private"}); err == nil {
		t.Fatal("unknown player accepted")
	}
}
