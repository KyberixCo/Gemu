package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

var playerKeys = map[string]bool{"gemu.favorites": true, "gemu.feedback": true, "gemu.management.v1": true, "gemu.setup.v1": true, "keyboard.v1": true}
var globalKeys = map[string]bool{"gemu.language.v1": true, "gemu.controllers.v1": true}

func ValidatePreference(scope, key, value string) error {
	allowed := playerKeys
	if scope == Global {
		allowed = globalKeys
	}
	if !allowed[key] || len(value) > 4*1024*1024 {
		return fmt.Errorf("invalid preference key or size")
	}
	if key == "gemu.language.v1" {
		if value != "es" && value != "en" {
			return fmt.Errorf("invalid language preference")
		}
		return nil
	}
	if key == "gemu.setup.v1" {
		if value != "complete" {
			return fmt.Errorf("invalid setup preference")
		}
		return nil
	}
	var decoded any
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return fmt.Errorf("invalid preference JSON: %w", err)
	}
	if key == "gemu.favorites" {
		items, ok := decoded.([]any)
		if !ok {
			return fmt.Errorf("invalid favorites")
		}
		for _, v := range items {
			if _, ok := v.(string); !ok {
				return fmt.Errorf("invalid favorites")
			}
		}
	}
	if key == "gemu.controllers.v1" {
		if _, ok := decoded.([]any); !ok {
			return fmt.Errorf("invalid controller registry")
		}
	}
	if key == "gemu.feedback" || key == "gemu.management.v1" || key == "keyboard.v1" {
		if _, ok := decoded.(map[string]any); !ok {
			return fmt.Errorf("invalid preference object")
		}
	}
	return rejectSecrets(decoded)
}

func rejectSecrets(value any) error {
	switch v := value.(type) {
	case map[string]any:
		for key, child := range v {
			switch strings.ToLower(key) {
			case "apikey", "password", "credentials", "devpassword", "secret", "token":
				return fmt.Errorf("credentials must use the secure vault")
			}
			if err := rejectSecrets(child); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range v {
			if err := rejectSecrets(child); err != nil {
				return err
			}
		}
	}
	return nil
}

// LoadPreferences imports a WebView's old settings once. The marker and values
// commit together, and INSERT OR IGNORE never overwrites a newer SQLite value.
type Preferences struct {
	Values         map[string]string `json:"values"`
	LegacyImported bool              `json:"legacyImported"`
}

// A nil legacy map only reads: it must not mark migration complete before the
// WebView has supplied its settings. An empty non-nil map is a fresh install.
func LoadPreferences(dir, scope string, legacy map[string]string) (Preferences, error) {
	db, err := Open(dir)
	if err != nil {
		return Preferences{}, err
	}
	defer db.Close()
	result := Preferences{Values: map[string]string{}}
	err = Update(db, func(tx *sql.Tx) error {
		done, err := Imported(tx, scope, "webview-preferences-v1")
		if err != nil {
			return err
		}
		if !done && legacy != nil {
			for key, value := range legacy {
				if err = ValidatePreference(scope, key, value); err != nil {
					return err
				}
				if _, err = tx.Exec("INSERT OR IGNORE INTO preferences(scope,key,value) VALUES (?,?,?)", scope, key, value); err != nil {
					return err
				}
			}
			if err = MarkImported(tx, scope, "webview-preferences-v1"); err != nil {
				return err
			}
			done = true
		}
		result.LegacyImported = done
		rows, err := tx.Query("SELECT key,value FROM preferences WHERE scope=?", scope)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var key, value string
			if err = rows.Scan(&key, &value); err != nil {
				return err
			}
			result.Values[key] = value
		}
		return rows.Err()
	})
	return result, err
}

func SavePreference(dir, scope, key, value string) error {
	if err := ValidatePreference(scope, key, value); err != nil {
		return err
	}
	db, err := Open(dir)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("INSERT INTO preferences(scope,key,value) VALUES (?,?,?) ON CONFLICT(scope,key) DO UPDATE SET value=excluded.value", scope, key, value)
	return err
}
