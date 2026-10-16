package storage

import (
	"database/sql"
	"fmt"
)

const ftsMigration = `
CREATE TABLE library_documents (
 id INTEGER PRIMARY KEY,
 scope TEXT NOT NULL,
 game_id TEXT NOT NULL,
 source_id TEXT NOT NULL,
 kind TEXT NOT NULL DEFAULT 'game' CHECK(kind='game'),
 title TEXT NOT NULL,
 description TEXT NOT NULL,
 genre TEXT NOT NULL,
 notes TEXT NOT NULL DEFAULT '',
 UNIQUE(scope,game_id),
 FOREIGN KEY(scope,game_id) REFERENCES games(scope,id) ON DELETE CASCADE
);
CREATE VIRTUAL TABLE library_documents_fts USING fts5(
 title, description, genre, notes,
 content='library_documents', content_rowid='id',
 tokenize='unicode61 remove_diacritics 2', prefix='2 3 4'
);
CREATE TRIGGER library_documents_ai AFTER INSERT ON library_documents BEGIN
 INSERT INTO library_documents_fts(rowid,title,description,genre,notes)
 VALUES(new.id,new.title,new.description,new.genre,new.notes);
END;
CREATE TRIGGER library_documents_ad AFTER DELETE ON library_documents BEGIN
 INSERT INTO library_documents_fts(library_documents_fts,rowid,title,description,genre,notes)
 VALUES('delete',old.id,old.title,old.description,old.genre,old.notes);
END;
CREATE TRIGGER library_documents_au AFTER UPDATE ON library_documents BEGIN
 INSERT INTO library_documents_fts(library_documents_fts,rowid,title,description,genre,notes)
 VALUES('delete',old.id,old.title,old.description,old.genre,old.notes);
 INSERT INTO library_documents_fts(rowid,title,description,genre,notes)
 VALUES(new.id,new.title,new.description,new.genre,new.notes);
END;
CREATE TRIGGER games_search_ai AFTER INSERT ON games BEGIN
 INSERT INTO library_documents(scope,game_id,source_id,title,description,genre)
 VALUES(new.scope,new.id,new.source_id,new.title,
 COALESCE(json_extract(CAST(new.metadata AS TEXT),'$.description'),''),
 COALESCE((SELECT group_concat(value,' ') FROM json_each(CAST(new.metadata AS TEXT),'$.genres')),''));
END;
CREATE TRIGGER games_search_au AFTER UPDATE OF title,metadata,source_id ON games BEGIN
 UPDATE library_documents SET title=new.title,source_id=new.source_id,
 description=COALESCE(json_extract(CAST(new.metadata AS TEXT),'$.description'),''),
 genre=COALESCE((SELECT group_concat(value,' ') FROM json_each(CAST(new.metadata AS TEXT),'$.genres')),'')
 WHERE scope=old.scope AND game_id=old.id;
END;
INSERT INTO library_documents(scope,game_id,source_id,title,description,genre)
 SELECT scope,id,source_id,title,
 COALESCE(json_extract(CAST(metadata AS TEXT),'$.description'),''),
 COALESCE((SELECT group_concat(value,' ') FROM json_each(CAST(metadata AS TEXT),'$.genres')),'')
 FROM games;
INSERT INTO library_documents_fts(library_documents_fts) VALUES('rebuild');
`

// Verify actual CREATE/INSERT/MATCH support, rather than compile option names.
func verifyFTS5(tx *sql.Tx) error {
	for _, query := range []string{
		"CREATE VIRTUAL TABLE temp.gemu_fts_probe USING fts5(value)",
		"INSERT INTO gemu_fts_probe(value) VALUES ('verification')",
	} {
		if _, err := tx.Exec(query); err != nil {
			return fmt.Errorf("FTS5 is required for library search: %w", err)
		}
	}
	var count int
	if err := tx.QueryRow("SELECT count(*) FROM gemu_fts_probe WHERE gemu_fts_probe MATCH 'verification'").Scan(&count); err != nil {
		return fmt.Errorf("FTS5 is required for library search: %w", err)
	}
	if count != 1 {
		return fmt.Errorf("FTS5 verification failed")
	}
	_, err := tx.Exec("DROP TABLE temp.gemu_fts_probe")
	return err
}
