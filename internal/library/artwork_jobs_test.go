package library

import (
	"context"
	"kyberix.com/gemu/internal/domain"
	"reflect"
	"testing"
	"time"
)

func waitArtwork(t *testing.T, s *Service, id string) ArtworkSearch {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		v, err := s.GetArtworkSearch(id)
		if err != nil {
			t.Fatal(err)
		}
		if v.State != "searching" {
			return v
		}
		time.Sleep(time.Millisecond * 5)
	}
	t.Fatal("search did not finish")
	return ArtworkSearch{}
}
func artworkFixture(t *testing.T) (*Service, Snapshot) {
	s := New(t.TempDir(), nil)
	cachedTestArt(t, s, "https://example.com/new.png", 460, 215)
	s.WithContentProviders(map[string]ContentProvider{"test": &fakeContent{result: ContentResult{Description: "Must not replace description", Artworks: []Artwork{{ID: "candidate", Kind: "cover", URL: "https://example.com/new.png"}}}}})
	snapshot := Snapshot{Version: 1, Sources: []SourceConfig{{ID: "steam", Root: "keep"}}, Games: []Entry{{Game: domain.Game{ID: "one", Title: "One"}, SourceID: "steam", Metadata: Metadata{Description: "Keep description", Genres: []string{"Keep"}}, Player: PlayerState{Status: "ready", Achievements: []Achievement{{ID: "keep", Unlocked: true}}}}, {Game: domain.Game{ID: "two", Title: "Two"}, Metadata: Metadata{Cover: "keep-second"}}}}
	if err := s.writeSnapshot(snapshot); err != nil {
		t.Fatal(err)
	}
	return s, snapshot
}
func TestArtworkCommitOnlyChangesSelectedArtAndUndoSurvivesRestart(t *testing.T) {
	s, before := artworkFixture(t)
	job, err := s.StartArtworkSearch(ArtworkSearchRequest{GameIDs: []string{"one"}, Providers: []ContentRequest{{ID: "test"}}})
	if err != nil {
		t.Fatal(err)
	}
	view := waitArtwork(t, s, job.ID)
	if len(view.Proposals) != 1 || len(view.Proposals[0].Candidates) != 1 {
		t.Fatalf("proposal=%+v", view)
	}
	unpublished, _ := s.Load()
	if !reflect.DeepEqual(unpublished, before) {
		t.Fatal("search published changes")
	}
	if _, err = s.CommitArtwork(ArtworkCommit{ID: job.ID, Choices: []ArtworkChoice{{GameID: "one", IDs: []string{"forged"}}}}); err == nil {
		t.Fatal("forged selection accepted")
	}
	result, err := s.CommitArtwork(ArtworkCommit{ID: job.ID, Choices: []ArtworkChoice{{GameID: "one", IDs: []string{"candidate"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Games[0].Metadata.Cover == "" || result.Games[0].Metadata.Description != "Keep description" || !reflect.DeepEqual(result.Games[0].Player, before.Games[0].Player) || !reflect.DeepEqual(result.Games[1], before.Games[1]) || !reflect.DeepEqual(result.Sources, before.Sources) {
		t.Fatal("commit changed unrelated data")
	}
	restored, err := New(s.dir, nil).UndoArtwork()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored, before) {
		t.Fatal("undo did not restore previous state")
	}
}
func TestArtworkConflictAndMissingOnlyPreserveExistingArt(t *testing.T) {
	s, _ := artworkFixture(t)
	job, _ := s.StartArtworkSearch(ArtworkSearchRequest{GameIDs: []string{"two"}, Providers: []ContentRequest{{ID: "test"}}, MissingOnly: true})
	v := waitArtwork(t, s, job.ID)
	if len(v.Proposals[0].Candidates) != 0 {
		t.Fatal("existing cover offered for replacement")
	}
	job, _ = s.StartArtworkSearch(ArtworkSearchRequest{GameIDs: []string{"one"}, Providers: []ContentRequest{{ID: "test"}}})
	waitArtwork(t, s, job.ID)
	snapshot, _ := s.Load()
	snapshot.Games[0].Metadata.Cover = "newer-choice"
	s.writeSnapshot(snapshot)
	if _, err := s.CommitArtwork(ArtworkCommit{ID: job.ID, Choices: []ArtworkChoice{{GameID: "one", IDs: []string{"candidate"}}}}); err == nil {
		t.Fatal("newer choice overwritten")
	}
}

type cancellableContent struct{ started chan struct{} }

func (*cancellableContent) Name() string                                   { return "Cancel" }
func (*cancellableContent) ArtworkHosts() []string                         { return nil }
func (*cancellableContent) Check(context.Context, map[string]string) error { return nil }
func (p *cancellableContent) Fetch(ctx context.Context, _ domain.Game, _ map[string]string) (ContentResult, error) {
	close(p.started)
	<-ctx.Done()
	return ContentResult{}, ctx.Err()
}
func TestArtworkCancellationStopsProviderAndDoesNotPublish(t *testing.T) {
	s, before := artworkFixture(t)
	p := &cancellableContent{started: make(chan struct{})}
	s.content["cancel"] = p
	job, err := s.StartArtworkSearch(ArtworkSearchRequest{GameIDs: []string{"one"}, Providers: []ContentRequest{{ID: "cancel"}}})
	if err != nil {
		t.Fatal(err)
	}
	<-p.started
	s.CancelArtworkSearch(job.ID)
	if v := waitArtwork(t, s, job.ID); v.State != "cancelled" {
		t.Fatal("not cancelled")
	}
	snapshot, _ := s.Load()
	if !reflect.DeepEqual(snapshot, before) {
		t.Fatal("cancel published results")
	}
}
func TestArtworkManualChoicesSurviveSourceEnrichment(t *testing.T) {
	s, _ := artworkFixture(t)
	old := Metadata{Cover: "/library-assets/manual.img", ManualArt: []string{"cover"}}
	metadata := Metadata{}
	s.enrichMetadata(context.Background(), domain.Game{Title: "One"}, &metadata, old, []ContentRequest{{ID: "test"}}, map[string]bool{}, &[]string{})
	if metadata.Cover != old.Cover || len(metadata.ManualArt) != 1 {
		t.Fatal("manual cover overwritten during import")
	}
}

func TestArtworkRotationPreservesPrimaryAndUndoRejectsLaterChanges(t *testing.T) {
	s, _ := artworkFixture(t)
	bg := cachedTestArt(t, s, "https://example.com/bg.png", 1280, 720)
	s.content["test"] = &fakeContent{result: ContentResult{Artworks: []Artwork{{ID: "extra-bg", Kind: "background", URL: "https://example.com/bg.png"}}}}
	snapshot, _ := s.Load()
	snapshot.Games[0].Metadata.Background = "/original"
	s.writeSnapshot(snapshot)
	job, _ := s.StartArtworkSearch(ArtworkSearchRequest{GameIDs: []string{"one"}, Providers: []ContentRequest{{ID: "test"}}})
	waitArtwork(t, s, job.ID)
	result, err := s.CommitArtwork(ArtworkCommit{ID: job.ID, Choices: []ArtworkChoice{{GameID: "one", RotateIDs: []string{"extra-bg"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Games[0].Metadata.Background != "/original" || !reflect.DeepEqual(result.Games[0].Metadata.Backgrounds, []string{"/original", bg}) {
		t.Fatal("rotation replaced the primary background")
	}
	result.Games[0].Metadata.Background = "/later-choice"
	s.writeSnapshot(result)
	if _, err = s.UndoArtwork(); err == nil {
		t.Fatal("undo overwrote a later choice")
	}
}

func TestHighResolutionCoverCanAlsoRotateAndLegacyCoverIsMeasured(t *testing.T) {
	s, _ := artworkFixture(t)
	large := cachedTestArt(t, s, "https://example.com/large-cover.png", 1920, 1080)
	snapshot, _ := s.Load()
	snapshot.Games[0].Metadata.Cover = large
	s.writeSnapshot(snapshot)
	loaded, _ := s.Load()
	if len(loaded.Games[0].Metadata.Artworks) != 1 || loaded.Games[0].Metadata.Artworks[0].Width != 1920 {
		t.Fatal("legacy cover dimensions were not recovered")
	}
	s.content["test"] = &fakeContent{result: ContentResult{Artworks: []Artwork{{ID: "large-cover", Kind: "cover", URL: "https://example.com/large-cover.png"}, {ID: "small-cover", Kind: "cover", URL: "https://example.com/new.png"}}}}
	job, _ := s.StartArtworkSearch(ArtworkSearchRequest{GameIDs: []string{"one"}, Providers: []ContentRequest{{ID: "test"}}})
	waitArtwork(t, s, job.ID)
	if _, err := s.CommitArtwork(ArtworkCommit{ID: job.ID, Choices: []ArtworkChoice{{GameID: "one", RotateIDs: []string{"small-cover"}}}}); err == nil {
		t.Fatal("small cover accepted as background")
	}
	result, err := s.CommitArtwork(ArtworkCommit{ID: job.ID, Choices: []ArtworkChoice{{GameID: "one", RotateIDs: []string{"large-cover"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Games[0].Metadata.Cover != large || result.Games[0].Metadata.Background != large {
		t.Fatal("cover was not retained while also used as scenery")
	}
}

func TestArtworkIconsAreSelectablePersistentAndSurviveImport(t *testing.T) {
	s, before := artworkFixture(t)
	icon := cachedTestArt(t, s, "https://example.com/icon.png", 256, 256)
	s.content["test"] = &fakeContent{result: ContentResult{Artworks: []Artwork{{ID: "chosen-icon", Kind: "icon", URL: "https://example.com/icon.png"}}}}
	job, err := s.StartArtworkSearch(ArtworkSearchRequest{GameIDs: []string{"one"}, Providers: []ContentRequest{{ID: "test"}}, MissingOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	proposal := waitArtwork(t, s, job.ID)
	if len(proposal.Proposals[0].Candidates) != 1 {
		t.Fatal("icon filtered out")
	}
	result, err := s.CommitArtwork(ArtworkCommit{ID: job.ID, Choices: []ArtworkChoice{{GameID: "one", IDs: []string{"chosen-icon"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Games[0].Metadata.Icon != icon {
		t.Fatal("icon selection not persisted")
	}
	restarted := New(s.dir, nil)
	loaded, _ := restarted.Load()
	if loaded.Games[0].Metadata.Icon != icon {
		t.Fatal("icon lost on restart")
	}
	newer := cachedTestArt(t, s, "https://example.com/other-icon.png", 256, 256)
	s.content["test"] = &fakeContent{result: ContentResult{Artworks: []Artwork{{ID: "other-icon", Kind: "icon", URL: "https://example.com/other-icon.png"}}}}
	m := Metadata{}
	s.enrichMetadata(context.Background(), domain.Game{Title: "One"}, &m, result.Games[0].Metadata, []ContentRequest{{ID: "test"}}, map[string]bool{}, &[]string{})
	if m.Icon != icon || m.Icon == newer {
		t.Fatal("manual icon overwritten by import")
	}
	restored, err := restarted.UndoArtwork()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored, before) {
		t.Fatal("icon undo failed")
	}
}
