package library

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"kyberix.com/gemu/internal/domain"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeContent struct {
	result ContentResult
	err    error
	calls  int
}

func (*fakeContent) Name() string                                     { return "Content" }
func (*fakeContent) ArtworkHosts() []string                           { return []string{"example.com"} }
func (f *fakeContent) Check(context.Context, map[string]string) error { return f.err }
func (f *fakeContent) Fetch(context.Context, domain.Game, map[string]string) (ContentResult, error) {
	f.calls++
	return f.result, f.err
}

type languageAwareContent struct {
	fakeContent
	language string
}

func (f *languageAwareContent) FetchLocalized(_ context.Context, _ domain.Game, _ map[string]string, language string) (ContentResult, error) {
	f.language = language
	return f.result, f.err
}

func TestMetadataEnrichmentForwardsApplicationLanguage(t *testing.T) {
	s := New(t.TempDir(), nil)
	provider := &languageAwareContent{}
	s.WithContentProviders(map[string]ContentProvider{"test": provider})
	s.enrichMetadata(context.Background(), domain.Game{Title: "One"}, &Metadata{}, Metadata{}, []ContentRequest{{ID: "test", Language: "es"}}, map[string]bool{}, &[]string{})
	if provider.language != "es" {
		t.Fatalf("provider language = %q, want es", provider.language)
	}
}

func cachedTestArt(t *testing.T, service *Service, address string, width, height int) string {
	t.Helper()
	if err := os.MkdirAll(service.artDir, 0700); err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("%x", sha256.Sum256([]byte(address))) + ".img"
	file, err := os.Create(filepath.Join(service.artDir, name))
	if err != nil {
		t.Fatal(err)
	}
	if err = png.Encode(file, image.NewRGBA(image.Rect(0, 0, width, height))); err != nil {
		t.Fatal(err)
	}
	file.Close()
	return "/library-assets/" + name
}

func TestContentImportPriorityActualDimensionsPublicationAndNoSecrets(t *testing.T) {
	service := New(t.TempDir(), map[string]Source{"one": &fakeSource{id: "one"}})
	first := &fakeContent{result: ContentResult{Artworks: []Artwork{{ID: "small", Kind: "background", URL: "https://example.com/small.png", Width: 1920, Height: 1080}, {ID: "large", Kind: "background", URL: "https://example.com/large.png", Width: 1920, Height: 1080}, {ID: "cover", Kind: "cover", URL: "https://example.com/cover.png", Width: 460, Height: 215}}}}
	second := &fakeContent{result: ContentResult{Description: "Enriched", Genres: []string{"Puzzle"}, Artworks: []Artwork{{ID: "second", Kind: "background", URL: "https://example.com/second.png", Width: 1280, Height: 720}}}}
	service.WithContentProviders(map[string]ContentProvider{"first": first, "second": second})
	cachedTestArt(t, service, "https://example.com/small.png", 460, 215)
	primary := cachedTestArt(t, service, "https://example.com/large.png", 1920, 1080)
	cover := cachedTestArt(t, service, "https://example.com/cover.png", 460, 215)
	cachedTestArt(t, service, "https://example.com/second.png", 1280, 720)
	draft, err := service.Prepare(context.Background(), Request{SourceID: "one", ContentProviders: []ContentRequest{{ID: "second", Priority: 2}, {ID: "first", Priority: 1, Credentials: map[string]string{"apiKey": "never-persist-secret"}}}})
	if err != nil {
		t.Fatal(err)
	}
	metadata := draft.Games[0].Metadata
	if metadata.Background != primary || metadata.Cover != cover || metadata.Description != "Enriched" || len(metadata.Artworks) != 3 {
		t.Fatalf("metadata=%+v", metadata)
	}
	if len(draft.Diagnostics) != 1 {
		t.Fatalf("expected rejected real dimensions: %v", draft.Diagnostics)
	}
	for _, art := range metadata.Artworks {
		if !strings.HasPrefix(art.URL, "/library-assets/") {
			t.Fatal("remote URL persisted")
		}
	}
	unpublished, _ := service.Load()
	if len(unpublished.Games) != 0 {
		t.Fatal("draft published before commit")
	}
	snapshot, err := service.Commit(draft.Token)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(snapshot)
	if strings.Contains(string(data), "never-persist-secret") {
		t.Fatal("credentials persisted")
	}
	reloaded, err := service.Load()
	if err != nil || len(reloaded.Games[0].Metadata.Artworks) != 3 {
		t.Fatal("gallery lost after reload")
	}
	first.err = ErrContentUnavailable
	first.result = ContentResult{}
	second.err = ErrContentUnavailable
	second.result = ContentResult{}
	offline, err := service.Prepare(context.Background(), Request{SourceID: "one", ContentProviders: []ContentRequest{{ID: "first"}, {ID: "second"}}})
	if err != nil || offline.Games[0].Metadata.Background != primary || len(offline.Games[0].Metadata.Artworks) != 3 {
		t.Fatalf("offline=%+v err=%v", offline, err)
	}
	first.err = nil
	first.result = ContentResult{Artworks: []Artwork{{ID: "updated-cover", Kind: "cover", URL: "https://example.com/cover.png", Width: 460, Height: 215}}}
	partial, err := service.Prepare(context.Background(), Request{SourceID: "one", ContentProviders: []ContentRequest{{ID: "first"}}})
	if err != nil || partial.Games[0].Metadata.Background != primary || len(partial.Games[0].Metadata.Artworks) != 3 {
		t.Fatalf("partial refresh lost cached backgrounds: %+v err=%v", partial, err)
	}
}

func TestQuotaStopsProviderAndDuplicateRequestsAreIgnored(t *testing.T) {
	provider := &fakeContent{err: ErrContentUnavailable}
	service := New(t.TempDir(), nil).WithContentProviders(map[string]ContentProvider{"one": provider})
	unavailable := map[string]bool{}
	diagnostics := []string{}
	metadata := Metadata{}
	for i := 0; i < 2; i++ {
		service.enrichMetadata(context.Background(), domain.Game{Title: "Example"}, &metadata, Metadata{}, []ContentRequest{{ID: "one"}, {ID: "one"}}, unavailable, &diagnostics)
	}
	if provider.calls != 1 || len(diagnostics) != 1 {
		t.Fatalf("calls=%d diagnostics=%v", provider.calls, diagnostics)
	}
	if err := service.CheckContentProvider(context.Background(), ContentRequest{ID: "unsupported"}); err == nil {
		t.Fatal("unknown provider accepted")
	}
}
