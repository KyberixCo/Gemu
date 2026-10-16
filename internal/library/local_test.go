package library

import (
	"database/sql"
	"errors"
	"testing"

	"kyberix.com/gemu/internal/domain"
)

func TestLocalSourceAndConfigurationRollbackTogether(t *testing.T) {
	s := New(t.TempDir(), nil)
	entries := []Entry{{SourceID: "retroarch", Game: domain.Game{ID: "one", Title: "Original"}}}
	if _, err := s.PublishLocal(SourceConfig{ID: "retroarch"}, entries, nil); err != nil {
		t.Fatal(err)
	}
	entries[0].Game.Title = "Uncommitted"
	_, err := s.PublishLocal(SourceConfig{ID: "retroarch"}, entries, func(tx *sql.Tx) error {
		if _, err := tx.Exec("INSERT INTO retroarch_setups(scope,data) VALUES ('standalone','{}')"); err != nil {
			return err
		}
		return errors.New("simulated config write failure")
	})
	if err == nil {
		t.Fatal("config failure accepted")
	}
	snapshot, err := s.Load()
	if err != nil || snapshot.Games[0].Game.Title != "Original" {
		t.Fatal(snapshot, err)
	}
	db, err := s.database()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var count int
	if err = db.QueryRow("SELECT count(*) FROM retroarch_setups").Scan(&count); err != nil || count != 0 {
		t.Fatal("partial configuration", count, err)
	}
	page, err := s.Search(SearchRequest{Query: "uncommitted"})
	if err != nil || page.Total != 0 {
		t.Fatal("partial FTS commit", page, err)
	}
}
