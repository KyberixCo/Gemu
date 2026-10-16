package profiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLegacyProfilesImportOnceAndPreserveBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.json")
	data := `[{"id":"00112233445566778899aabbccddeeff","name":"Jugador ñ"}]`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	s := New(dir)
	items, err := s.List()
	if err != nil || len(items) != 1 || items[0].Name != "Jugador ñ" {
		t.Fatal(items, err)
	}
	if _, err = s.Create("Nuevo"); err != nil {
		t.Fatal(err)
	}
	backup, err := os.ReadFile(path)
	if err != nil || string(backup) != data {
		t.Fatal("modified legacy file", err)
	}
	if err = os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	items, err = New(dir).List()
	if err != nil || len(items) != 2 || items[0].ID != "00112233445566778899aabbccddeeff" {
		t.Fatal(items, err)
	}
}

func TestInvalidProfilesMigrationRetriesWithoutPartialRegistry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.json")
	for _, data := range []string{"broken", `[{"id":"00112233445566778899aabbccddeeff","name":"Valid"},{"id":"../../escape","name":"Invalid"}]`} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := New(dir).List(); err == nil {
			t.Fatal("invalid registry imported")
		}
	}
	if err := os.WriteFile(path, []byte("[]"), 0600); err != nil {
		t.Fatal(err)
	}
	items, err := New(dir).List()
	if err != nil || len(items) != 0 {
		t.Fatal(items, err)
	}
}
