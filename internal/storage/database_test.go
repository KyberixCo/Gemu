package storage

import (
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestSchemaDurabilityAndConcurrentOpen(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "spaces and # unicode ñ")
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	for pragma, want := range map[string]string{"journal_mode": "wal", "synchronous": "2", "foreign_keys": "1", "user_version": "3"} {
		var got string
		if err = db.QueryRow("PRAGMA " + pragma).Scan(&got); err != nil || got != want {
			t.Fatalf("%s=%s want %s: %v", pragma, got, want, err)
		}
	}
	db.Close()
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db, err := Open(dir)
			if err != nil {
				t.Error(err)
				return
			}
			defer db.Close()
			if _, err = db.Exec("INSERT OR IGNORE INTO preferences VALUES ('global','gemu.language.v1','en')"); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	db, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var check string
	if err = db.QueryRow("PRAGMA quick_check").Scan(&check); err != nil || check != "ok" {
		t.Fatal(check, err)
	}
}

func TestRejectFutureSchema(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("PRAGMA user_version=999"); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if db, err = Open(dir); err == nil {
		db.Close()
		t.Fatal("accepted newer schema")
	}
}

func TestInterruptedTransactionRecovery(t *testing.T) {
	if os.Getenv("GEMU_TEST_INTERRUPTED_DB") != "" {
		db, err := Open(os.Getenv("GEMU_TEST_INTERRUPTED_DB"))
		if err != nil {
			t.Fatal(err)
		}
		tx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}
		for _, statement := range []string{
			"UPDATE preferences SET value='es' WHERE scope='global'",
			"INSERT INTO legacy_imports VALUES ('global','interrupted')",
			"CREATE TABLE interrupted_schema (id INTEGER)",
			"PRAGMA user_version=999",
		} {
			if _, err = tx.Exec(statement); err != nil {
				t.Fatal(err)
			}
		}
		// Abrupt exit: neither rollback nor connection cleanup can run.
		os.Exit(77)
	}
	dir := t.TempDir()
	if err := SavePreference(dir, Global, "gemu.language.v1", "en"); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(os.Args[0], "-test.run=^TestInterruptedTransactionRecovery$")
	command.Env = append(os.Environ(), "GEMU_TEST_INTERRUPTED_DB="+dir)
	output, err := command.CombinedOutput()
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 77 {
		t.Fatalf("helper: %s %v", output, err)
	}
	db, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var value string
	if err = db.QueryRow("SELECT value FROM preferences WHERE scope='global'").Scan(&value); err != nil || value != "en" {
		t.Fatal(value, err)
	}
	var count int
	if err = db.QueryRow("SELECT count(*) FROM sqlite_master WHERE name='interrupted_schema'").Scan(&count); err != nil || count != 0 {
		t.Fatal(count, err)
	}
	if err = Update(db, func(tx *sql.Tx) error {
		done, err := Imported(tx, Global, "interrupted")
		if done {
			t.Fatal("uncommitted migration survived")
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestPreferenceMigrationIsAtomicAndDoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	state, err := LoadPreferences(dir, Standalone, nil)
	if err != nil || state.LegacyImported {
		t.Fatal(state, err)
	}
	if err = SavePreference(dir, Standalone, "gemu.favorites", `["new"]`); err != nil {
		t.Fatal(err)
	}
	_, err = LoadPreferences(dir, Standalone, map[string]string{"gemu.favorites": `["old"]`, "gemu.feedback": "broken"})
	if err == nil {
		t.Fatal("accepted corrupt preferences")
	}
	state, err = LoadPreferences(dir, Standalone, nil)
	if err != nil || state.LegacyImported || len(state.Values) != 1 {
		t.Fatal(state, err)
	}
	state, err = LoadPreferences(dir, Standalone, map[string]string{"gemu.favorites": `["old"]`, "gemu.feedback": `{"volume":25}`})
	if err != nil || !state.LegacyImported || state.Values["gemu.favorites"] != `["new"]` {
		t.Fatal(state, err)
	}
	state, err = LoadPreferences(dir, Standalone, map[string]string{"gemu.feedback": "broken"})
	if err != nil || state.Values["gemu.feedback"] != `{"volume":25}` {
		t.Fatal(state, err)
	}
	for key, value := range map[string]string{"unknown": "{}", "gemu.favorites": "[1]", "gemu.management.v1": `{"llm":{"apiKey":"secret"}}`} {
		if err = SavePreference(dir, Standalone, key, value); err == nil {
			t.Fatal("accepted", key)
		}
	}
	if err = SavePreference(dir, Global, "gemu.favorites", "[]"); err == nil {
		t.Fatal("player preference in global scope")
	}
	if err = SavePreference(dir, "missing", "gemu.favorites", "[]"); err == nil || !strings.Contains(err.Error(), "FOREIGN KEY") {
		t.Fatal("unknown owner", err)
	}
}
