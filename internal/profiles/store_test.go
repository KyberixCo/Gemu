package profiles

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"testing"
)

func TestProfilesPersistAndValidate(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	items, err := s.List()
	if err != nil || len(items) != 0 {
		t.Fatal(items, err)
	}
	a, err := s.Create("  Manuel  ")
	if err != nil || a.Name != "Manuel" {
		t.Fatal(a, err)
	}
	b, err := s.Create("Invitado")
	if err != nil || a.ID == b.ID {
		t.Fatal(b, err)
	}
	items, err = New(dir).List()
	if err != nil || len(items) != 2 {
		t.Fatal(items, err)
	}
	for _, name := range []string{"", "   ", "MANUEL", "a\nb"} {
		if _, err = s.Create(name); err == nil {
			t.Fatalf("accepted %q", name)
		}
	}

	da, _ := s.Directory(a.ID)
	db, _ := s.Directory(b.ID)
	if da == db {
		t.Fatal("shared profile directory")
	}
	if _, err = s.Directory("../../library.json"); err == nil {
		t.Fatal("accepted unknown profile")
	}
}

func TestProfileAvatarPersistsAndRejectsInvalidImages(t *testing.T) {
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	avatar := "data:image/png;base64," + base64.StdEncoding.EncodeToString(data.Bytes())
	store := New(t.TempDir())
	created, err := store.CreateWithAvatar("Player", avatar)
	if err != nil || created.Avatar != avatar {
		t.Fatalf("profile=%+v err=%v", created, err)
	}
	loaded, err := store.List()
	if err != nil || len(loaded) != 1 || loaded[0].Avatar != avatar {
		t.Fatalf("profiles=%+v err=%v", loaded, err)
	}
	for _, invalid := range []string{"https://example.test/avatar.png", "data:image/png;base64,bm90LWltYWdl"} {
		if _, err := store.CreateWithAvatar("Other", invalid); err == nil {
			t.Fatalf("accepted invalid avatar %q", invalid)
		}
	}
}
