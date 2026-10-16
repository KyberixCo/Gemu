package library

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

type ArtworkSearchRequest struct {
	GameIDs     []string         `json:"gameIds"`
	Providers   []ContentRequest `json:"providers"`
	MissingOnly bool             `json:"missingOnly"`
}
type ArtworkProposal struct {
	GameID      string    `json:"gameId"`
	Title       string    `json:"title"`
	Candidates  []Artwork `json:"candidates"`
	Diagnostics []string  `json:"diagnostics"`
}
type ArtworkSearch struct {
	ID        string            `json:"id"`
	State     string            `json:"state"`
	Total     int               `json:"total"`
	Completed int               `json:"completed"`
	Current   string            `json:"current"`
	Proposals []ArtworkProposal `json:"proposals"`
}
type ArtworkChoice struct {
	GameID    string   `json:"gameId"`
	IDs       []string `json:"ids"`
	RotateIDs []string `json:"rotateIds,omitempty"`
}
type ArtworkCommit struct {
	ID      string          `json:"id"`
	Choices []ArtworkChoice `json:"choices"`
}
type ArtworkState struct {
	Cover       string    `json:"cover"`
	Background  string    `json:"background"`
	Logo        string    `json:"logo"`
	Icon        string    `json:"icon"`
	Backgrounds []string  `json:"backgrounds"`
	Manual      []string  `json:"manual"`
	Artworks    []Artwork `json:"artworks"`
}
type ArtworkUndo struct {
	Before map[string]ArtworkState `json:"before"`
	After  map[string]ArtworkState `json:"after"`
}
type artworkJob struct {
	view    ArtworkSearch
	before  map[string]ArtworkState
	cancel  context.CancelFunc
	expires time.Time
	missing bool
}

func artState(m Metadata) ArtworkState {
	return ArtworkState{Cover: m.Cover, Background: m.Background, Logo: m.Logo, Icon: m.Icon, Backgrounds: m.Backgrounds, Manual: m.ManualArt, Artworks: m.Artworks}
}
func applyArt(m *Metadata, a ArtworkState) {
	m.Cover = a.Cover
	m.Background = a.Background
	m.Logo = a.Logo
	m.Icon = a.Icon
	m.Backgrounds = a.Backgrounds
	m.ManualArt = a.Manual
	m.Artworks = a.Artworks
}
func missingKind(m Metadata, kind string) bool {
	switch kind {
	case "cover":
		return m.Cover == ""
	case "background":
		return m.Background == ""
	case "logo":
		return m.Logo == ""
	case "icon":
		return m.Icon == ""
	}
	return false
}

func (s *Service) StartArtworkSearch(req ArtworkSearchRequest) (ArtworkSearch, error) {
	if len(req.GameIDs) == 0 || len(req.GameIDs) > 200 || len(req.Providers) == 0 || len(req.Providers) > 2 {
		return ArtworkSearch{}, fmt.Errorf("elige juegos y al menos un proveedor de arte")
	}
	snapshot, err := s.Load()
	if err != nil {
		return ArtworkSearch{}, err
	}
	selected := map[string]bool{}
	for _, id := range req.GameIDs {
		selected[id] = true
	}
	entries := []Entry{}
	before := map[string]ArtworkState{}
	for _, entry := range snapshot.Games {
		if selected[entry.Game.ID] {
			entries = append(entries, entry)
			before[entry.Game.ID] = artState(entry.Metadata)
		}
	}
	if len(entries) != len(selected) {
		return ArtworkSearch{}, fmt.Errorf("algún juego ya no está disponible")
	}
	for _, p := range req.Providers {
		if s.content[p.ID] == nil {
			return ArtworkSearch{}, fmt.Errorf("proveedor no disponible")
		}
	}
	sort.SliceStable(req.Providers, func(i, j int) bool { return req.Providers[i].Priority < req.Providers[j].Priority })
	var token [16]byte
	if _, err = rand.Read(token[:]); err != nil {
		return ArtworkSearch{}, err
	}
	id := hex.EncodeToString(token[:])
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	job := &artworkJob{view: ArtworkSearch{ID: id, State: "searching", Total: len(entries), Proposals: []ArtworkProposal{}}, before: before, cancel: cancel, expires: time.Now().Add(30 * time.Minute), missing: req.MissingOnly}
	s.mu.Lock()
	for key, old := range s.artJobs {
		if time.Now().After(old.expires) {
			old.cancel()
			delete(s.artJobs, key)
		}
	}
	if len(s.artJobs) >= 8 {
		s.mu.Unlock()
		cancel()
		return ArtworkSearch{}, fmt.Errorf("cierra las búsquedas anteriores antes de continuar")
	}
	s.artJobs[id] = job
	view := job.view
	s.mu.Unlock()
	go s.searchArtwork(ctx, job, entries, req.Providers)
	return view, nil
}
func (s *Service) searchArtwork(ctx context.Context, job *artworkJob, entries []Entry, providers []ContentRequest) {
	defer job.cancel()
	unavailable := map[string]bool{}
	for _, entry := range entries {
		if ctx.Err() != nil {
			break
		}
		s.mu.Lock()
		job.view.Current = entry.Game.Title
		s.mu.Unlock()
		proposal := ArtworkProposal{GameID: entry.Game.ID, Title: entry.Game.Title, Candidates: []Artwork{}, Diagnostics: []string{}}
		seen := map[string]bool{}
		for _, request := range providers {
			if ctx.Err() != nil {
				break
			}
			if unavailable[request.ID] {
				proposal.Diagnostics = append(proposal.Diagnostics, "Proveedor pendiente por acceso o cuota: "+request.ID)
				continue
			}
			provider := s.content[request.ID]
			result, err := s.fetchContent(ctx, provider, entry.Game, request)
			if ctx.Err() != nil {
				break
			}
			if err != nil {
				proposal.Diagnostics = append(proposal.Diagnostics, provider.Name()+": "+err.Error())
				if errors.Is(err, ErrContentUnavailable) {
					unavailable[request.ID] = true
				}
			}
			proposal.Diagnostics = append(proposal.Diagnostics, result.Diagnostics...)
			counts := map[string]int{}
			for _, art := range result.Artworks {
				if ctx.Err() != nil {
					break
				}
				if art.Kind != "cover" && art.Kind != "background" && art.Kind != "logo" && art.Kind != "icon" {
					continue
				}
				if job.missing && !missingKind(entry.Metadata, art.Kind) {
					continue
				}
				limit := 6
				if job.missing {
					limit = 1
				}
				if counts[art.Kind] >= limit {
					continue
				}
				counts[art.Kind]++
				local, err := s.cacheArtwork(ctx, art.URL, provider.ArtworkHosts())
				if err != nil {
					proposal.Diagnostics = append(proposal.Diagnostics, "No se pudo descargar una imagen de "+provider.Name())
					continue
				}
				data, err := os.ReadFile(filepath.Join(s.artDir, strings.TrimPrefix(local, "/library-assets/")))
				if err != nil {
					continue
				}
				w, h, err := artworkDimensions(data)
				if err != nil || art.Kind == "background" && !BackgroundSize(w, h) {
					continue
				}
				if seen[local] {
					continue
				}
				seen[local] = true
				art.URL = local
				art.Provider = request.ID
				art.Width = w
				art.Height = h
				proposal.Candidates = append(proposal.Candidates, art)
			}
		}
		s.mu.Lock()
		job.view.Proposals = append(job.view.Proposals, proposal)
		if ctx.Err() == nil {
			job.view.Completed++
		}
		s.mu.Unlock()
	}
	s.mu.Lock()
	job.view.Current = ""
	job.view.State = "ready"
	if ctx.Err() != nil {
		job.view.State = "cancelled"
	}
	s.mu.Unlock()
}
func (s *Service) GetArtworkSearch(id string) (ArtworkSearch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := s.artJobs[id]
	if j == nil || time.Now().After(j.expires) {
		return ArtworkSearch{}, fmt.Errorf("la búsqueda expiró; repítela")
	}
	v := j.view
	v.Proposals = append([]ArtworkProposal{}, v.Proposals...)
	return v, nil
}
func (s *Service) CancelArtworkSearch(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j := s.artJobs[id]; j != nil {
		j.cancel()
	}
}
func (s *Service) DiscardArtworkSearch(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j := s.artJobs[id]; j != nil {
		j.cancel()
		delete(s.artJobs, id)
	}
}
func (s *Service) CancelArtworkSearches() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, j := range s.artJobs {
		j.cancel()
		delete(s.artJobs, id)
	}
}

func (s *Service) CommitArtwork(req ArtworkCommit) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job := s.artJobs[req.ID]
	if job == nil || time.Now().After(job.expires) || job.view.State == "searching" {
		return Snapshot{}, fmt.Errorf("termina la búsqueda antes de guardar")
	}
	snapshot, err := s.Load()
	if err != nil {
		return Snapshot{}, err
	}
	undo := &ArtworkUndo{Before: map[string]ArtworkState{}, After: map[string]ArtworkState{}}
	seenGames := map[string]bool{}
	for _, choice := range req.Choices {
		if len(choice.IDs) == 0 && len(choice.RotateIDs) == 0 {
			continue
		}
		if seenGames[choice.GameID] {
			return Snapshot{}, fmt.Errorf("selección de juego duplicada")
		}
		seenGames[choice.GameID] = true
		var entry *Entry
		for i := range snapshot.Games {
			if snapshot.Games[i].Game.ID == choice.GameID {
				entry = &snapshot.Games[i]
				break
			}
		}
		before, ok := job.before[choice.GameID]
		if !ok || entry == nil || !reflect.DeepEqual(artState(entry.Metadata), before) {
			return Snapshot{}, fmt.Errorf("el arte cambió durante la revisión; repite la búsqueda")
		}
		allowed := map[string]Artwork{}
		for _, p := range job.view.Proposals {
			if p.GameID == choice.GameID {
				for _, a := range p.Candidates {
					allowed[a.ID] = a
				}
			}
		}
		kinds := map[string]bool{}
		chosen := []Artwork{}
		for _, id := range choice.IDs {
			a, ok := allowed[id]
			if !ok || kinds[a.Kind] {
				return Snapshot{}, fmt.Errorf("selección de arte inválida")
			}
			if job.missing && !missingKind(entry.Metadata, a.Kind) {
				return Snapshot{}, fmt.Errorf("la imagen ya no falta")
			}
			kinds[a.Kind] = true
			chosen = append(chosen, a)
		}
		rotations := []Artwork{}
		seenRotation := map[string]bool{}
		if len(choice.RotateIDs) > 3 || job.missing && len(choice.RotateIDs) > 0 {
			return Snapshot{}, fmt.Errorf("elige hasta tres fondos adicionales")
		}
		for _, id := range choice.RotateIDs {
			a, ok := allowed[id]
			if !ok || (a.Kind != "background" && a.Kind != "cover") || !BackgroundSize(a.Width, a.Height) || seenRotation[id] {
				return Snapshot{}, fmt.Errorf("fondo de rotación inválido")
			}
			seenRotation[id] = true
			rotations = append(rotations, a)
		}
		undo.Before[choice.GameID] = before
		// Copy slices before modifying them: job baselines must stay immutable.
		entry.Metadata.Artworks = append([]Artwork{}, entry.Metadata.Artworks...)
		entry.Metadata.ManualArt = append([]string{}, entry.Metadata.ManualArt...)
		for _, a := range chosen {
			switch a.Kind {
			case "cover":
				entry.Metadata.Cover = a.URL
			case "background":
				entry.Metadata.Background = a.URL
				entry.Metadata.Backgrounds = []string{a.URL}
			case "logo":
				entry.Metadata.Logo = a.URL
			case "icon":
				entry.Metadata.Icon = a.URL
			}
			found := false
			for _, old := range entry.Metadata.Artworks {
				found = found || old.URL == a.URL
			}
			if !found {
				entry.Metadata.Artworks = append(entry.Metadata.Artworks, a)
			}
			marked := false
			for _, kind := range entry.Metadata.ManualArt {
				marked = marked || kind == a.Kind
			}
			if !marked {
				entry.Metadata.ManualArt = append(entry.Metadata.ManualArt, a.Kind)
			}
		}
		if len(rotations) > 0 {
			urls := append([]string{}, entry.Metadata.Backgrounds...)
			if entry.Metadata.Background != "" {
				found := false
				for _, u := range urls {
					found = found || u == entry.Metadata.Background
				}
				if !found {
					urls = append([]string{entry.Metadata.Background}, urls...)
				}
			}
			for _, a := range rotations {
				found := false
				for _, u := range urls {
					found = found || u == a.URL
				}
				if !found {
					urls = append(urls, a.URL)
				}
				if entry.Metadata.Background == "" {
					entry.Metadata.Background = a.URL
				}
				found = false
				for _, old := range entry.Metadata.Artworks {
					found = found || old.URL == a.URL
				}
				if !found {
					entry.Metadata.Artworks = append(entry.Metadata.Artworks, a)
				}
			}
			entry.Metadata.Backgrounds = urls
			marked := false
			for _, kind := range entry.Metadata.ManualArt {
				marked = marked || kind == "background"
			}
			if !marked {
				entry.Metadata.ManualArt = append(entry.Metadata.ManualArt, "background")
			}
		}
		undo.After[choice.GameID] = artState(entry.Metadata)
	}
	if len(undo.Before) == 0 {
		return Snapshot{}, fmt.Errorf("no hay imágenes seleccionadas")
	}
	snapshot.ArtworkUndo = undo
	if err = s.writeSnapshot(snapshot); err != nil {
		return Snapshot{}, err
	}
	delete(s.artJobs, req.ID)
	snapshot.revision++
	return snapshot, nil
}
func (s *Service) UndoArtwork() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot, err := s.Load()
	if err != nil {
		return Snapshot{}, err
	}
	undo := snapshot.ArtworkUndo
	if undo == nil {
		return Snapshot{}, fmt.Errorf("no hay un cambio de arte para deshacer")
	}
	for id, after := range undo.After {
		found := false
		for _, entry := range snapshot.Games {
			if entry.Game.ID == id {
				found = true
				if !reflect.DeepEqual(artState(entry.Metadata), after) {
					return Snapshot{}, fmt.Errorf("el arte cambió después; no se puede deshacer sin revisar")
				}
			}
		}
		if !found {
			return Snapshot{}, fmt.Errorf("el juego ya no está disponible")
		}
	}
	for i := range snapshot.Games {
		if before, ok := undo.Before[snapshot.Games[i].Game.ID]; ok {
			applyArt(&snapshot.Games[i].Metadata, before)
		}
	}
	snapshot.ArtworkUndo = nil
	if err := s.writeSnapshot(snapshot); err != nil {
		return Snapshot{}, err
	}
	snapshot.revision++
	return snapshot, nil
}
