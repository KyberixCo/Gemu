package library

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"kyberix.com/gemu/internal/domain"
)

func TestLegacyLibraryMigrationAndFailedWriteRecovery(t *testing.T) {
	dir := t.TempDir()
	legacy := Snapshot{Version: 1, Sources: []SourceConfig{{ID: "steam", Root: "C:/Steam", ProfileID: "steam-user"}}, Games: []Entry{{SourceID: "steam", Game: domain.Game{ID: "steam:42", Title: "Título", Installations: []domain.Installation{{Provider: "steam", ExternalID: "42", Path: "C:/Steam/Game", Installed: true}}}, Metadata: Metadata{Description: "Descripción", ManualArt: []string{"icon"}, Icon: "/library-assets/example.img"}, Player: PlayerState{Profile: Profile{ID: "steam-user"}, Achievements: []Achievement{{ID: "first", Unlocked: true}}}}}, ArtworkUndo: &ArtworkUndo{Before: map[string]ArtworkState{"steam:42": {}}, After: map[string]ArtworkState{}}}
	data, _ := json.Marshal(legacy)
	path := filepath.Join(dir, "library.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	s := New(dir, nil)
	got, err := s.Load()
	if err != nil || !samePublishedSnapshot(got, legacy) {
		t.Fatalf("migration: %+v %v", got, err)
	}
	stale, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	got.Games[0].Game.Title = "Edited"
	if err = s.writeSnapshot(got); err != nil {
		t.Fatal(err)
	}
	if err = s.writeSnapshot(stale); err == nil {
		t.Fatal("stale snapshot overwrote committed data")
	}
	got, err = s.Load()
	if err != nil {
		t.Fatal(err)
	}
	got.Games = append(got.Games, got.Games[0])
	if err = s.writeSnapshot(got); err == nil {
		t.Fatal("duplicate game accepted")
	}
	loaded, err := New(dir, nil).Load()
	if err != nil || len(loaded.Games) != 1 || loaded.Games[0].Game.Title != "Edited" || loaded.ArtworkUndo == nil {
		t.Fatal(loaded, err)
	}
	backup, err := os.ReadFile(path)
	if err != nil || string(backup) != string(data) {
		t.Fatal("legacy backup modified", err)
	}
	if err = os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = New(dir, nil).Load(); err != nil {
		t.Fatal("reimported an already migrated file", err)
	}
}

func TestCorruptLegacyLibraryCanBeRepairedAndRetried(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "library.json")
	for _, data := range []string{"broken", `{"version":99}`, `{"version":1,"games":[{"game":{"ID":"duplicate"}},{"game":{"ID":"duplicate"}}]}`} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := New(dir, nil).Load(); err == nil {
			t.Fatal("accepted corrupt or unsupported legacy library")
		}
	}
	if err := os.WriteFile(path, []byte(`{"version":1,"games":[],"sources":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := New(dir, nil).Load()
	if err != nil || len(got.Games) != 0 {
		t.Fatal(got, err)
	}
}
