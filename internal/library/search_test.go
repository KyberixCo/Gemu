package library

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/profiles"
	"kyberix.com/gemu/internal/storage"
)

func searchEntry(id, title, description, genre, source string) Entry {
	return Entry{Game: domain.Game{ID: id, Title: title}, SourceID: source, Metadata: Metadata{Description: description, Genres: []string{genre}}}
}

func searchIDs(t *testing.T, s *Service, request SearchRequest) []string {
	t.Helper()
	page, err := s.Search(request)
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{}
	for _, hit := range page.Hits {
		ids = append(ids, hit.GameID)
		if hit.Kind != "game" || hit.SourceID == "" {
			t.Fatal("missing provenance", hit)
		}
	}
	return ids
}

func TestSearchAccentsPhrasesRankingAndTransactionalChanges(t *testing.T) {
	s := New(t.TempDir(), nil)
	snapshot := Snapshot{Version: 1, Sources: []SourceConfig{}, Games: []Entry{
		searchEntry("title", "Océano profundo", "Viaje espacial", "Exploración", "steam"),
		searchEntry("description", "Otro juego", "Océano profundo y recuerdos", "Acción RPG", "manual"),
		searchEntry("split", "Océano azul muy profundo", "", "Arcade", "steam"),
	}}
	if err := s.writeSnapshot(snapshot); err != nil {
		t.Fatal(err)
	}
	ids := searchIDs(t, s, SearchRequest{Query: `"OCEANO PROFUNDO"`})
	if len(ids) != 2 || ids[0] != "title" {
		t.Fatal("phrase/rank", ids)
	}
	ids = searchIDs(t, s, SearchRequest{Query: "explorac"})
	if len(ids) != 1 || ids[0] != "title" {
		t.Fatal("genre prefix", ids)
	}
	ids = searchIDs(t, s, SearchRequest{Query: "oceano profundo", SourceID: "manual"})
	if len(ids) != 1 || ids[0] != "description" {
		t.Fatal("source", ids)
	}
	loaded, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	loaded.Games[0].Game.Title = "Bosque"
	loaded.Games[0].Metadata.Description = "Senderos nuevos"
	if err = s.writeSnapshot(loaded); err != nil {
		t.Fatal(err)
	}
	if ids = searchIDs(t, s, SearchRequest{Query: "viaje"}); len(ids) != 0 {
		t.Fatal("stale description", ids)
	}
	if ids = searchIDs(t, s, SearchRequest{Query: "senderos"}); len(ids) != 1 {
		t.Fatal("missing update", ids)
	}
	loaded, err = s.Load()
	if err != nil {
		t.Fatal(err)
	}
	loaded.Games = append(loaded.Games, loaded.Games[0])
	loaded.Games[0].Game.Title = "Uncommitted"
	if err = s.writeSnapshot(loaded); err == nil {
		t.Fatal("duplicate accepted")
	}
	if ids = searchIDs(t, s, SearchRequest{Query: "uncommitted"}); len(ids) != 0 {
		t.Fatal("partial index write", ids)
	}
	loaded, err = s.Load()
	if err != nil {
		t.Fatal(err)
	}
	loaded.Games = loaded.Games[1:]
	if err = s.writeSnapshot(loaded); err != nil {
		t.Fatal(err)
	}
	if ids = searchIDs(t, New(s.dir, nil), SearchRequest{Query: "senderos"}); len(ids) != 0 {
		t.Fatal("deleted game in index", ids)
	}
}

func TestSearchIsolatesPlayersAndFiltersFavoritesBeforePagination(t *testing.T) {
	dir := t.TempDir()
	store := profiles.New(dir)
	a, err := store.Create("A")
	if err != nil {
		t.Fatal(err)
	}
	b, err := store.Create("B")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []profiles.Profile{a, b} {
		s := New(filepath.Join(dir, "profiles", p.ID), nil).WithPlayerStorage(dir, p.ID)
		games := []Entry{}
		for i := 0; i < 60; i++ {
			games = append(games, searchEntry(fmt.Sprintf("game:%02d", i), fmt.Sprintf("Orbit %02d", i), p.Name+" private", "Arcade", "steam"))
		}
		if err = s.writeSnapshot(Snapshot{Version: 1, Games: games}); err != nil {
			t.Fatal(err)
		}
	}
	s := New(filepath.Join(dir, "profiles", a.ID), nil).WithPlayerStorage(dir, a.ID)
	if ids := searchIDs(t, s, SearchRequest{Query: "B private"}); len(ids) != 0 {
		t.Fatal("another player leaked", ids)
	}
	if err = storage.SavePreference(dir, a.ID, "gemu.favorites", `["game:59"]`); err != nil {
		t.Fatal(err)
	}
	page, err := s.Search(SearchRequest{Query: "orbit", FavoritesOnly: true, Limit: 1})
	if err != nil || page.Total != 1 || len(page.Hits) != 1 || page.Hits[0].GameID != "game:59" || page.HasMore {
		t.Fatal(page, err)
	}
	page, err = s.Search(SearchRequest{Query: "orbit", Limit: 7, Offset: 7})
	if err != nil || page.Total != 60 || len(page.Hits) != 7 || page.Hits[0].GameID != "game:07" || !page.HasMore {
		t.Fatal(page, err)
	}
	for _, query := range []string{"", "***", `"`, `orbit OR nonexistent`, `title:missing`, `'; DROP TABLE games; --`, `orbit NEAR(foo)`} {
		page, err = s.Search(SearchRequest{Query: query})
		if err != nil || page.Total != 0 {
			t.Fatalf("literal %q: %+v %v", query, page, err)
		}
	}
	for _, request := range []SearchRequest{{Query: strings.Repeat("x", 2049)}, {Query: strings.Repeat("x ", 33)}, {Query: "orbit", Limit: 101}, {Query: "orbit", Offset: -1}} {
		if _, err = s.Search(request); err == nil {
			t.Fatal("invalid request accepted", request)
		}
	}
}
