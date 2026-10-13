package profiles

import "testing"

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
