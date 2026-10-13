package main

import (
	"context"
	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/library"
	"kyberix.com/gemu/internal/profiles"
	"testing"
	"time"
)

type profileSource struct{}

func (profileSource) Name() string           { return "Test" }
func (profileSource) ArtworkHosts() []string { return nil }
func (profileSource) Discover(context.Context, string) (library.Preview, error) {
	return library.Preview{Games: []domain.Game{{ID: "test:1", Title: "Game", Installations: []domain.Installation{{Installed: true}}}}, Profiles: []library.Profile{{ID: "steam-a", Name: "Steam A"}, {ID: "steam-b", Name: "Steam B"}}}, nil
}

type blockingProfileSource struct {
	profileSource
	started chan struct{}
	release chan struct{}
}

func (s blockingProfileSource) Metadata(context.Context, domain.Game) (library.Metadata, error) {
	close(s.started)
	<-s.release
	return library.Metadata{Status: "ready"}, nil
}
func TestSwitchWhileImportIsInFlight(t *testing.T) {
	dir := t.TempDir()
	source := blockingProfileSource{started: make(chan struct{}), release: make(chan struct{})}
	app := &App{dir: dir, profiles: profiles.New(dir), sources: map[string]library.Source{"test": source}}
	a, _ := app.CreateProfile("A")
	b, _ := app.CreateProfile("B")
	if err := app.SelectProfile(a.ID); err != nil {
		t.Fatal(err)
	}
	result := make(chan library.Draft, 1)
	failure := make(chan error, 1)
	go func() {
		draft, err := app.PrepareLibraryImport(library.Request{SourceID: "test", ProfileID: "steam-a"})
		if err != nil {
			failure <- err
			return
		}
		result <- draft
	}()
	select {
	case <-source.started:
	case <-time.After(5 * time.Second):
		t.Fatal("import did not start")
	}
	switched := make(chan error, 1)
	go func() { switched <- app.SelectProfile(b.ID) }()
	select {
	case err := <-switched:
		close(source.release)
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		close(source.release)
		t.Fatal("profile switch blocked by metadata I/O")
	}
	select {
	case draft := <-result:
		if _, err := app.CommitLibraryImport(draft.Token); err == nil {
			t.Fatal("in-flight import leaked to new profile")
		}
	case err := <-failure:
		t.Fatal(err)
	case <-time.After(5 * time.Second):
		t.Fatal("import did not finish")
	}
	snapshot, err := app.GetLibrary()
	if err != nil || len(snapshot.Games) != 0 {
		t.Fatal(snapshot, err)
	}
}
func (profileSource) Metadata(context.Context, domain.Game) (library.Metadata, error) {
	return library.Metadata{Status: "ready"}, nil
}
func (profileSource) Player(_ context.Context, _ domain.Game, p library.Profile, _ string) library.PlayerState {
	return library.PlayerState{Profile: p, Status: "ready"}
}
func TestProfileLibraryIsolation(t *testing.T) {
	dir := t.TempDir()
	app := &App{dir: dir, profiles: profiles.New(dir), sources: map[string]library.Source{"test": profileSource{}}}
	if _, err := app.GetLibrary(); err == nil {
		t.Fatal("library accessible before selecting profile")
	}
	a, _ := app.CreateProfile("A")
	b, _ := app.CreateProfile("B")
	if err := app.SelectProfile(a.ID); err != nil {
		t.Fatal(err)
	}
	draft, err := app.PrepareLibraryImport(library.Request{SourceID: "test", ProfileID: "steam-a"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = app.CommitLibraryImport(draft.Token); err != nil {
		t.Fatal(err)
	}
	pending, err := app.PrepareLibraryImport(library.Request{SourceID: "test", ProfileID: "steam-a"})
	if err != nil {
		t.Fatal(err)
	}
	if err = app.SelectProfile(b.ID); err != nil {
		t.Fatal(err)
	}
	empty, err := app.GetLibrary()
	if err != nil || len(empty.Games) != 0 {
		t.Fatal(empty, err)
	}
	if _, err = app.CommitLibraryImport(pending.Token); err == nil {
		t.Fatal("published another player's draft")
	}
	draft, err = app.PrepareLibraryImport(library.Request{SourceID: "test", ProfileID: "steam-b"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = app.CommitLibraryImport(draft.Token); err != nil {
		t.Fatal(err)
	}
	if err = app.SelectProfile(a.ID); err != nil {
		t.Fatal(err)
	}
	saved, err := app.GetLibrary()
	if err != nil || len(saved.Games) != 1 || saved.Games[0].Player.Profile.ID != "steam-a" {
		t.Fatal(saved, err)
	}
	if err = app.SelectProfile("../unknown"); err == nil {
		t.Fatal("selected unknown profile")
	}
}
