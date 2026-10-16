// Package storage owns the on-disk schema shared by profiles, libraries and UI
// preferences. Credentials and artwork files remain in their dedicated stores.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const Filename = "gemu.sqlite3"
const Global = "global"
const Standalone = "standalone"

var migrations = []string{`
CREATE TABLE scopes (id TEXT PRIMARY KEY);
INSERT INTO scopes VALUES ('global'), ('standalone');
CREATE TABLE profiles (id TEXT PRIMARY KEY REFERENCES scopes(id), name TEXT NOT NULL, avatar TEXT NOT NULL, position INTEGER NOT NULL);
CREATE TABLE legacy_imports (scope TEXT NOT NULL REFERENCES scopes(id), name TEXT NOT NULL, PRIMARY KEY(scope, name));
CREATE TABLE libraries (scope TEXT PRIMARY KEY REFERENCES scopes(id), revision INTEGER NOT NULL DEFAULT 0, undo BLOB);
CREATE TABLE sources (scope TEXT NOT NULL REFERENCES scopes(id), id TEXT NOT NULL, position INTEGER NOT NULL, data BLOB NOT NULL, PRIMARY KEY(scope,id));
CREATE TABLE games (scope TEXT NOT NULL REFERENCES scopes(id), id TEXT NOT NULL, position INTEGER NOT NULL, source_id TEXT NOT NULL, title TEXT NOT NULL, metadata BLOB NOT NULL, player BLOB NOT NULL, PRIMARY KEY(scope,id));
CREATE TABLE installations (scope TEXT NOT NULL, game_id TEXT NOT NULL, position INTEGER NOT NULL, provider TEXT NOT NULL, external_id TEXT NOT NULL, path TEXT NOT NULL, installed INTEGER NOT NULL, PRIMARY KEY(scope,game_id,position), FOREIGN KEY(scope,game_id) REFERENCES games(scope,id) ON DELETE CASCADE);
CREATE TABLE preferences (scope TEXT NOT NULL REFERENCES scopes(id), key TEXT NOT NULL, value TEXT NOT NULL, PRIMARY KEY(scope,key));
`, ftsMigration, `CREATE TABLE retroarch_setups (scope TEXT PRIMARY KEY REFERENCES scopes(id), data BLOB NOT NULL);`}

func Open(dir string) (*sql.DB, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	path, err := filepath.Abs(filepath.Join(dir, Filename))
	if err != nil {
		return nil, err
	}
	u := url.URL{Scheme: "file", Path: "/" + filepath.ToSlash(path)}
	if filepath.VolumeName(path) == "" {
		u.Path = filepath.ToSlash(path)
	}
	q := url.Values{"_pragma": {"busy_timeout(5000)", "foreign_keys(1)", "synchronous(FULL)"}, "_txlock": {"immediate"}}
	u.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	fail := func(err error) (*sql.DB, error) { db.Close(); return nil, err }
	if _, err = db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fail(err)
	}
	err = Update(db, func(tx *sql.Tx) error {
		var version int
		if err := tx.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
			return err
		}
		if version < 0 || version > len(migrations) {
			return fmt.Errorf("database schema is newer than this version of Gemu")
		}
		for i := version; i < len(migrations); i++ {
			if i == 1 {
				if err := verifyFTS5(tx); err != nil {
					return err
				}
			}
			if _, err := tx.Exec(migrations[i]); err != nil {
				return err
			}
			if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version=%d", i+1)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return db, nil
}

func Update(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func Imported(tx *sql.Tx, scope, name string) (bool, error) {
	var count int
	err := tx.QueryRow("SELECT count(*) FROM legacy_imports WHERE scope=? AND name=?", scope, name).Scan(&count)
	return count != 0, err
}

func MarkImported(tx *sql.Tx, scope, name string) error {
	_, err := tx.Exec("INSERT INTO legacy_imports(scope,name) VALUES (?,?)", scope, name)
	return err
}
