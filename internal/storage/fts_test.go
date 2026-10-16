package storage

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func legacySearchDB(t *testing.T, metadata string) string {
	t.Helper()
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, Filename))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(migrations[0]); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("PRAGMA user_version=1"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO games(scope,id,position,source_id,title,metadata,player) VALUES ('standalone','one',0,'steam','Océano profundo',?,'{}')", []byte(metadata)); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestFTSMigrationBackfillsAndTracksUpdatesDeletesAndRebuild(t *testing.T) {
	dir := legacySearchDB(t, `{"description":"Planetas olvidados","genres":["Acción RPG"]}`)
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	count := func(query string, want int) {
		t.Helper()
		var got int
		err := db.QueryRow("SELECT count(*) FROM library_documents_fts WHERE library_documents_fts MATCH ?", query).Scan(&got)
		if err != nil || got != want {
			t.Fatalf("%s = %d want %d: %v", query, got, want, err)
		}
	}
	count(`"oceano profundo"`, 1)
	count("accion", 1)
	count("planetas", 1)
	if _, err = db.Exec("UPDATE games SET title='Bosque',metadata=? WHERE id='one'", []byte(`{"description":"Senderos","genres":["Aventura"]}`)); err != nil {
		t.Fatal(err)
	}
	count("oceano", 0)
	count("planetas", 0)
	count("senderos", 1)
	count("aventura", 1)
	if _, err = db.Exec("INSERT INTO library_documents_fts(library_documents_fts) VALUES ('rebuild')"); err != nil {
		t.Fatal(err)
	}
	count("bosque", 1)
	if _, err = db.Exec("DELETE FROM games WHERE id='one'"); err != nil {
		t.Fatal(err)
	}
	count("bosque", 0)
	count("senderos", 0)
	if _, err = db.Exec("INSERT INTO library_documents_fts(library_documents_fts,rank) VALUES ('integrity-check',1)"); err != nil {
		t.Fatal(err)
	}
}

func TestFailedFTSMigrationRollsBackAndCanRetry(t *testing.T) {
	dir := legacySearchDB(t, "invalid JSON")
	if db, err := Open(dir); err == nil {
		db.Close()
		t.Fatal("migrated invalid metadata")
	}
	db, err := sql.Open("sqlite", filepath.Join(dir, Filename))
	if err != nil {
		t.Fatal(err)
	}
	var version, count int
	if err = db.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != 1 {
		t.Fatal(version, err)
	}
	if err = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE name LIKE 'library_documents%'").Scan(&count); err != nil || count != 0 {
		t.Fatal("partial schema", count, err)
	}
	if _, err = db.Exec("UPDATE games SET metadata='{}'"); err != nil {
		t.Fatal(err)
	}
	db.Close()
	db, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != 3 {
		t.Fatal(version, err)
	}
}
