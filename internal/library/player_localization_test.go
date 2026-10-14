package library

import (
	"context"
	"testing"

	"kyberix.com/gemu/internal/domain"
)

type localizedPlayerSource struct {
	fakeSource
	language string
}

func (s *localizedPlayerSource) PlayerLocalized(_ context.Context, _ domain.Game, profile Profile, _, language string) PlayerState {
	s.language = language
	return PlayerState{Profile: profile, Status: "ready"}
}

func TestImportForwardsLanguageToPlayerSource(t *testing.T) {
	source := &localizedPlayerSource{fakeSource: fakeSource{id: "one"}}
	service := New(t.TempDir(), map[string]Source{"one": source})
	if _, err := service.Prepare(context.Background(), Request{SourceID: "one", ProfileID: "player", Language: "en"}); err != nil {
		t.Fatal(err)
	}
	if source.language != "en" {
		t.Fatalf("player language = %q", source.language)
	}
}
