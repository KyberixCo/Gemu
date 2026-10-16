package library

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/storage"
)

func (s *Service) database() (*sql.DB, error) {
	db, err := storage.Open(s.dbDir)
	if err != nil {
		return nil, err
	}
	err = storage.Update(db, func(tx *sql.Tx) error {
		done, err := storage.Imported(tx, s.scope, "library.json")
		if err != nil || done {
			return err
		}
		snapshot, err := s.loadLegacy()
		if err != nil {
			return err
		}
		if err = s.replace(tx, snapshot, false); err != nil {
			return err
		}
		return storage.MarkImported(tx, s.scope, "library.json")
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func (s *Service) loadSQLite() (Snapshot, error) {
	db, err := s.database()
	if err != nil {
		return Snapshot{}, err
	}
	defer db.Close()
	result := Snapshot{Version: 1, Sources: []SourceConfig{}, Games: []Entry{}}
	// All components must come from the same committed revision.
	err = storage.Update(db, func(tx *sql.Tx) error {
		var undo []byte
		if err := tx.QueryRow("SELECT revision,undo FROM libraries WHERE scope=?", s.scope).Scan(&result.revision, &undo); err != nil {
			return err
		}
		if len(undo) > 0 {
			if err := json.Unmarshal(undo, &result.ArtworkUndo); err != nil {
				return err
			}
		}
		rows, err := tx.Query("SELECT data FROM sources WHERE scope=? ORDER BY position", s.scope)
		if err != nil {
			return err
		}
		for rows.Next() {
			var data []byte
			var source SourceConfig
			if err = rows.Scan(&data); err == nil {
				err = json.Unmarshal(data, &source)
			}
			if err != nil {
				rows.Close()
				return err
			}
			result.Sources = append(result.Sources, source)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		rows, err = tx.Query("SELECT id,title,source_id,metadata,player FROM games WHERE scope=? ORDER BY position", s.scope)
		if err != nil {
			return err
		}
		for rows.Next() {
			entry := Entry{}
			var metadata, player []byte
			if err = rows.Scan(&entry.Game.ID, &entry.Game.Title, &entry.SourceID, &metadata, &player); err == nil {
				err = json.Unmarshal(metadata, &entry.Metadata)
			}
			if err == nil {
				err = json.Unmarshal(player, &entry.Player)
			}
			if err != nil {
				rows.Close()
				return err
			}
			result.Games = append(result.Games, entry)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		byID := map[string]int{}
		for i, game := range result.Games {
			byID[game.Game.ID] = i
		}
		rows, err = tx.Query("SELECT game_id,provider,external_id,path,installed FROM installations WHERE scope=? ORDER BY game_id,position", s.scope)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			var installation domain.Installation
			if err = rows.Scan(&id, &installation.Provider, &installation.ExternalID, &installation.Path, &installation.Installed); err != nil {
				return err
			}
			i := byID[id]
			result.Games[i].Game.Installations = append(result.Games[i].Game.Installations, installation)
		}
		return rows.Err()
	})
	return result, err
}

func (s *Service) saveSQLite(snapshot Snapshot) error {
	return s.saveSQLiteWith(snapshot, nil)
}

func (s *Service) saveSQLiteWith(snapshot Snapshot, extra func(*sql.Tx) error) error {
	db, err := s.database()
	if err != nil {
		return err
	}
	defer db.Close()
	return storage.Update(db, func(tx *sql.Tx) error {
		if err := s.replace(tx, snapshot, true); err != nil {
			return err
		}
		if extra != nil {
			return extra(tx)
		}
		return nil
	})
}

func (s *Service) replace(tx *sql.Tx, snapshot Snapshot, checkRevision bool) error {
	if snapshot.Version != 1 {
		return fmt.Errorf("versión de biblioteca no compatible")
	}
	var undo []byte
	var err error
	if snapshot.ArtworkUndo != nil {
		undo, err = json.Marshal(snapshot.ArtworkUndo)
		if err != nil {
			return err
		}
	}
	if checkRevision {
		result, err := tx.Exec("UPDATE libraries SET revision=revision+1,undo=? WHERE scope=? AND revision=?", undo, s.scope, snapshot.revision)
		if err != nil {
			return err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if n != 1 {
			return fmt.Errorf("library changed while saving; reload and retry")
		}
	} else {
		if _, err = tx.Exec("INSERT INTO libraries(scope,revision,undo) VALUES (?,0,?)", s.scope, undo); err != nil {
			return err
		}
	}
	for _, table := range []string{"installations", "games", "sources"} {
		if _, err = tx.Exec("DELETE FROM "+table+" WHERE scope=?", s.scope); err != nil {
			return err
		}
	}
	for i, source := range snapshot.Sources {
		data, err := json.Marshal(source)
		if err != nil {
			return err
		}
		if _, err = tx.Exec("INSERT INTO sources(scope,id,position,data) VALUES (?,?,?,?)", s.scope, source.ID, i, data); err != nil {
			return err
		}
	}
	for i, entry := range snapshot.Games {
		metadata, err := json.Marshal(entry.Metadata)
		if err != nil {
			return err
		}
		player, err := json.Marshal(entry.Player)
		if err != nil {
			return err
		}
		if _, err = tx.Exec("INSERT INTO games(scope,id,position,source_id,title,metadata,player) VALUES (?,?,?,?,?,?,?)", s.scope, entry.Game.ID, i, entry.SourceID, entry.Game.Title, metadata, player); err != nil {
			return err
		}
		for j, installation := range entry.Game.Installations {
			if _, err = tx.Exec("INSERT INTO installations(scope,game_id,position,provider,external_id,path,installed) VALUES (?,?,?,?,?,?,?)", s.scope, entry.Game.ID, j, installation.Provider, installation.ExternalID, installation.Path, installation.Installed); err != nil {
				return err
			}
		}
	}
	return nil
}
