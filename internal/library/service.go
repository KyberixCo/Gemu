package library

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type pending struct {
	draft   Draft
	expires time.Time
}
type Service struct {
	dir     string
	artDir  string
	sources map[string]Source
	mu      sync.Mutex
	drafts  map[string]pending
	content map[string]ContentProvider
	artJobs map[string]*artworkJob
}

func New(dir string, sources map[string]Source) *Service {
	return &Service{dir: dir, artDir: filepath.Join(dir, "art"), sources: sources, drafts: map[string]pending{}, artJobs: map[string]*artworkJob{}}
}

// Share only downloaded artwork, never player data, across local profiles.
func (s *Service) WithArtworkDirectory(dir string) *Service { s.artDir = dir; return s }
func (s *Service) WithContentProviders(providers map[string]ContentProvider) *Service {
	s.content = providers
	return s
}

type SourceInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *Service) Sources() []SourceInfo {
	result := []SourceInfo{}
	for id, source := range s.sources {
		result = append(result, SourceInfo{ID: id, Name: source.Name()})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}
func (s *Service) Preview(ctx context.Context, source, root string) (Preview, error) {
	p, ok := s.sources[source]
	if !ok {
		return Preview{}, fmt.Errorf("fuente no compatible: %s", source)
	}
	return p.Discover(ctx, root)
}
func (s *Service) Prepare(ctx context.Context, req Request) (Draft, error) {
	preview, err := s.Preview(ctx, req.SourceID, req.Root)
	if err != nil {
		return Draft{}, err
	}
	profile := Profile{}
	for _, candidate := range preview.Profiles {
		if candidate.ID == req.ProfileID {
			profile = candidate
			break
		}
	}
	if req.ProfileID != "" && profile.ID == "" {
		return Draft{}, fmt.Errorf("el perfil seleccionado ya no está disponible")
	}
	draft := Draft{Source: SourceConfig{ID: req.SourceID, Root: preview.Root, ProfileID: profile.ID}, Games: []Entry{}, Diagnostics: append([]string{}, preview.Diagnostics...)}
	provider := s.sources[req.SourceID]
	draft.Source.Name = provider.Name()
	previous, err := s.Load()
	if err != nil {
		return Draft{}, err
	}
	byID := map[string]Entry{}
	content := append([]ContentRequest(nil), req.ContentProviders...)
	sort.SliceStable(content, func(i, j int) bool { return content[i].Priority < content[j].Priority })
	unavailable := map[string]bool{}
	for _, entry := range previous.Games {
		if entry.SourceID == req.SourceID {
			byID[entry.Game.ID] = entry
		}
	}
	for _, game := range preview.Games {
		installed := false
		for _, i := range game.Installations {
			installed = installed || i.Installed
		}
		if !installed {
			continue
		}
		if err := ctx.Err(); err != nil {
			return Draft{}, err
		}
		metadata, err := provider.Metadata(ctx, game)
		old := byID[game.ID]
		if err != nil {
			metadata = Metadata{Origin: "local", Status: "pending", Genres: []string{}}
			if old.Metadata.Status == "ready" || old.Metadata.Status == "cached" {
				metadata = old.Metadata
				metadata.Status = "cached"
			}
			draft.Diagnostics = append(draft.Diagnostics, game.Title+": metadatos no disponibles")
		}
		for _, art := range []*string{&metadata.Cover, &metadata.Background} {
			if *art == "" || strings.HasPrefix(*art, "/library-assets/") {
				continue
			}
			cached, err := s.cacheArtwork(ctx, *art, provider.ArtworkHosts())
			if err != nil {
				*art = ""
				draft.Diagnostics = append(draft.Diagnostics, game.Title+": arte no disponible")
			} else {
				*art = cached
			}
		}
		s.enrichMetadata(ctx, game, &metadata, old.Metadata, content, unavailable, &draft.Diagnostics)
		player := provider.Player(ctx, game, profile, req.APIKey)
		if player.Status != "ready" && profile.ID != "" && old.Player.Profile.ID == profile.ID && (old.Player.Status == "ready" || old.Player.Status == "cached") {
			message := player.Message
			player = old.Player
			player.Status = "cached"
			player.Message = "Se muestran los últimos logros guardados. " + message
		}
		if player.Achievements == nil {
			player.Achievements = []Achievement{}
		}
		draft.Games = append(draft.Games, Entry{Game: game, SourceID: req.SourceID, Metadata: metadata, Player: player})
	}
	if err := ctx.Err(); err != nil {
		return Draft{}, err
	}
	var token [16]byte
	if _, err = rand.Read(token[:]); err != nil {
		return Draft{}, err
	}
	draft.Token = hex.EncodeToString(token[:])
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, p := range s.drafts {
		if time.Now().After(p.expires) {
			delete(s.drafts, token)
		}
	}
	if len(s.drafts) >= 8 {
		return Draft{}, fmt.Errorf("hay demasiadas importaciones pendientes; cierra los asistentes anteriores")
	}
	s.drafts[draft.Token] = pending{draft: draft, expires: time.Now().Add(15 * time.Minute)}
	return draft, nil
}
func (s *Service) Discard(token string) { s.mu.Lock(); defer s.mu.Unlock(); delete(s.drafts, token) }
func (s *Service) Load() (Snapshot, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, "library.json"))
	if os.IsNotExist(err) {
		return Snapshot{Version: 1, Sources: []SourceConfig{}, Games: []Entry{}}, nil
	}
	if err != nil {
		return Snapshot{}, err
	}
	var snapshot Snapshot
	if err = json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("biblioteca guardada inválida: %w", err)
	}
	if snapshot.Version != 1 {
		return Snapshot{}, fmt.Errorf("versión de biblioteca no compatible")
	}
	if snapshot.Games == nil {
		snapshot.Games = []Entry{}
	}
	if snapshot.Sources == nil {
		snapshot.Sources = []SourceConfig{}
	}
	// Older libraries kept logos only in the imported gallery.
	for i := range snapshot.Games {
		m := &snapshot.Games[i].Metadata
		s.includeSceneryCover(m)
		if m.Logo == "" {
			for _, art := range m.Artworks {
				if art.Kind == "logo" {
					m.Logo = art.URL
					break
				}
			}
		}
		if m.Icon == "" {
			for _, art := range m.Artworks {
				if art.Kind == "icon" {
					m.Icon = art.URL
					break
				}
			}
		}
	}
	return snapshot, nil
}
func (s *Service) Commit(token string) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.drafts[token]
	if !ok || time.Now().After(p.expires) {
		return Snapshot{}, fmt.Errorf("la importación expiró; repite el asistente")
	}
	snapshot, err := s.Load()
	if err != nil {
		return Snapshot{}, err
	}
	games := []Entry{}
	for _, game := range snapshot.Games {
		if game.SourceID != p.draft.Source.ID {
			games = append(games, game)
		}
	}
	snapshot.Games = append(games, p.draft.Games...)
	sources := []SourceConfig{}
	for _, source := range snapshot.Sources {
		if source.ID != p.draft.Source.ID {
			sources = append(sources, source)
		}
	}
	snapshot.Sources = append(sources, p.draft.Source)
	sort.Slice(snapshot.Games, func(i, j int) bool { return snapshot.Games[i].Game.Title < snapshot.Games[j].Game.Title })
	if err = s.writeSnapshot(snapshot); err != nil {
		return Snapshot{}, err
	}
	delete(s.drafts, token)
	return snapshot, nil
}

func (s *Service) writeSnapshot(snapshot Snapshot) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(s.dir, 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(s.dir, "library-*.tmp")
	if err != nil {
		return err
	}
	name := file.Name()
	defer os.Remove(name)
	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = os.Rename(name, filepath.Join(s.dir, "library.json")); err != nil {
		return err
	}
	return nil
}
