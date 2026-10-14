package library

import (
	"context"
	"encoding/json"
	"fmt"
	"kyberix.com/gemu/internal/domain"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeSource struct {
	id   string
	fail bool
}

func (f *fakeSource) Name() string           { return f.id }
func (f *fakeSource) ArtworkHosts() []string { return []string{} }
func (f *fakeSource) Discover(context.Context, string) (Preview, error) {
	return Preview{Root: "library", Games: []domain.Game{{ID: f.id + ":42", Title: "Example", Installations: []domain.Installation{{Provider: f.id, Installed: true}}}}, Profiles: []Profile{{ID: "player", Name: "Local", Origin: "local"}}, Diagnostics: []string{}}, nil
}
func (f *fakeSource) Metadata(context.Context, domain.Game, string) (Metadata, error) {
	if f.fail {
		return Metadata{}, fmt.Errorf("offline")
	}
	return Metadata{Title: "Localized Example", Description: "Fetched", Genres: []string{"Adventure"}, Status: "ready", Origin: f.id}, nil
}
func (f *fakeSource) Player(_ context.Context, _ domain.Game, p Profile, key string) PlayerState {
	return PlayerState{Profile: p, Status: "pending", Message: "No credentials"}
}
func TestGenericImportPublicationAndReimport(t *testing.T) {
	dir := t.TempDir()
	service := New(dir, map[string]Source{"one": &fakeSource{id: "one"}, "two": &fakeSource{id: "two"}})
	for _, id := range []string{"one", "two", "one"} {
		draft, err := service.Prepare(context.Background(), Request{SourceID: id, ProfileID: "player", APIKey: "secret-never-persist"})
		if err != nil {
			t.Fatal(err)
		}
		if id == "one" {
			if draft.Games[0].Game.Title != "Localized Example" {
				t.Fatalf("localized title=%q", draft.Games[0].Game.Title)
			}
			snapshot, _ := service.Load()
			if len(snapshot.Games) > 2 {
				t.Fatal("prepared games published prematurely")
			}
		}
		if _, err = service.Commit(draft.Token); err != nil {
			t.Fatal(err)
		}
		if _, err = service.Commit(draft.Token); err == nil {
			t.Fatal("token reused")
		}
	}
	snapshot, err := New(dir, nil).Load()
	if err != nil || len(snapshot.Games) != 2 || len(snapshot.Sources) != 2 {
		t.Fatalf("snapshot=%+v err=%v", snapshot, err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "library.json"))
	if strings.Contains(string(data), "secret-never-persist") {
		t.Fatal("credential persisted")
	}
}
func TestCancelledAndPartialImport(t *testing.T) {
	service := New(t.TempDir(), map[string]Source{"one": &fakeSource{id: "one", fail: true}})
	draft, err := service.Prepare(context.Background(), Request{SourceID: "one", ProfileID: "player"})
	if err != nil || len(draft.Games) != 1 || draft.Games[0].Metadata.Status != "pending" {
		t.Fatalf("draft=%+v err=%v", draft, err)
	}
	service.Discard(draft.Token)
	if _, err = service.Commit(draft.Token); err == nil {
		t.Fatal("discarded import published")
	}
	snapshot, _ := service.Load()
	if len(snapshot.Games) != 0 {
		t.Fatal("cancel changed catalog")
	}
	if _, err = service.Prepare(context.Background(), Request{SourceID: "one", ProfileID: "unknown"}); err == nil {
		t.Fatal("unknown profile accepted")
	}
	if _, err = service.Preview(context.Background(), "unknown", ""); err == nil {
		t.Fatal("unknown source accepted")
	}
}
func TestArtworkHandlerRejectsPaths(t *testing.T) {
	for _, raw := range []string{"http://example.com/a", "https://example.com.evil/a", "https://example.com:443/a", "https://user@example.com/a"} {
		if artworkURL(raw, []string{"example.com"}) {
			t.Fatalf("accepted %s", raw)
		}
	}
	if !artworkURL("https://example.com/a", []string{"example.com"}) {
		t.Fatal("valid art rejected")
	}
	if assetName.MatchString("../library.json") {
		t.Fatal("path traversal accepted")
	}
}

func TestOfflineReimportRetainsOnlyMatchingProfileProgress(t *testing.T) {
	dir := t.TempDir()
	snapshot := Snapshot{Version: 1, Games: []Entry{{Game: domain.Game{ID: "one:42"}, SourceID: "one", Metadata: Metadata{Status: "ready", Description: "Cached metadata"}, Player: PlayerState{Profile: Profile{ID: "player"}, Status: "ready", Achievements: []Achievement{{ID: "first", Unlocked: true}}}}}, Sources: []SourceConfig{}}
	data, _ := json.Marshal(snapshot)
	if err := os.WriteFile(filepath.Join(dir, "library.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	service := New(dir, map[string]Source{"one": &fakeSource{id: "one", fail: true}})
	draft, err := service.Prepare(context.Background(), Request{SourceID: "one", ProfileID: "player"})
	if err != nil {
		t.Fatal(err)
	}
	if draft.Games[0].Metadata.Status != "cached" || draft.Games[0].Player.Status != "cached" || !draft.Games[0].Player.Achievements[0].Unlocked {
		t.Fatalf("draft=%+v", draft)
	}
	other, err := service.Prepare(context.Background(), Request{SourceID: "one"})
	if err != nil {
		t.Fatal(err)
	}
	if len(other.Games[0].Player.Achievements) != 0 {
		t.Fatal("progress leaked to a different profile")
	}
}

func TestOfflineReimportKeepsPreviouslyLocalizedSteamTitle(t *testing.T) {
	dir := t.TempDir()
	snapshot := Snapshot{
		Version: 1,
		Sources: []SourceConfig{},
		Games: []Entry{{
			Game:     domain.Game{ID: "one:42", Title: "Localized Store Title", Installations: []domain.Installation{{Installed: true}}},
			SourceID: "one",
			Metadata: Metadata{Origin: "steam-store", Status: "ready"},
		}},
	}
	if err := New(dir, nil).writeSnapshot(snapshot); err != nil {
		t.Fatal(err)
	}
	service := New(dir, map[string]Source{"one": &fakeSource{id: "one", fail: true}})
	draft, err := service.Prepare(context.Background(), Request{SourceID: "one"})
	if err != nil {
		t.Fatal(err)
	}
	if got := draft.Games[0].Game.Title; got != "Localized Store Title" {
		t.Fatalf("title=%q, want stored localized title", got)
	}
}

func TestCachedArtBypassesFrontendFallback(t *testing.T) {
	dir := t.TempDir()
	service := New(dir, nil)
	name := strings.Repeat("a", 64) + ".img"
	os.Mkdir(filepath.Join(dir, "art"), 0700)
	image := []byte{0x89, 'P', 'N', 'G', 13, 10, 26, 10}
	if err := os.WriteFile(filepath.Join(dir, "art", name), image, 0600); err != nil {
		t.Fatal(err)
	}
	handler := service.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("HTML fallback")) }))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/library-assets/"+name, nil))
	if response.Code != 200 || response.Header().Get("Content-Type") != "image/png" || response.Body.String() != string(image) {
		t.Fatalf("response=%v %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/library-assets/../library.json", nil))
	if response.Code != 404 {
		t.Fatal("catalog exposed")
	}
}
