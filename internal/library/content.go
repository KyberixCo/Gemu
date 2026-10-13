package library

import (
	"context"
	"errors"
	"fmt"
	"kyberix.com/gemu/internal/domain"
	"os"
	"path/filepath"
	"strings"
)

// Authentication and quota errors stop this provider for the remaining batch.
var ErrContentUnavailable = errors.New("proveedor no disponible")

func BackgroundSize(width, height int) bool { return width >= 1280 && height >= 720 && width > height }

func (s *Service) enrichMetadata(ctx context.Context, game domain.Game, metadata *Metadata, old Metadata, requests []ContentRequest, unavailable map[string]bool, diagnostics *[]string) {
	selectedCover, selectedBackground, selectedDescription, selectedGenres := false, false, false, false
	// Fresh store metadata does not include third-party artwork. Retain cached art
	// when enrichment is disabled or unavailable, without keeping remote URLs.
	artworks := []Artwork{}
	seen := map[string]bool{}
	queried := map[string]bool{}
	for _, request := range requests {
		provider := s.content[request.ID]
		if provider == nil || unavailable[request.ID] || queried[request.ID] {
			continue
		}
		queried[request.ID] = true
		result, err := provider.Fetch(ctx, game, request.Credentials)
		if err != nil {
			*diagnostics = append(*diagnostics, game.Title+": "+provider.Name()+" · "+err.Error())
			if errors.Is(err, ErrContentUnavailable) {
				unavailable[request.ID] = true
			}
		}
		for _, message := range result.Diagnostics {
			*diagnostics = append(*diagnostics, game.Title+": "+provider.Name()+" · "+message)
		}
		if !selectedDescription && result.Description != "" {
			metadata.Description = result.Description
			metadata.Origin = request.ID
			metadata.Status = "ready"
			selectedDescription = true
		}
		if !selectedGenres && len(result.Genres) > 0 {
			metadata.Genres = result.Genres
			selectedGenres = true
		}
		backgrounds, covers, logos, icons := 0, 0, 0, 0
		attempts := map[string]int{}
		for _, art := range result.Artworks {
			if ctx.Err() != nil {
				return
			}
			if art.Kind == "background" && (art.Width > 0 && art.Height > 0 && !BackgroundSize(art.Width, art.Height) || backgrounds >= 3) {
				continue
			}
			if art.Kind == "cover" && covers >= 1 || art.Kind == "logo" && logos >= 1 || art.Kind == "icon" && icons >= 1 {
				continue
			}
			if art.Kind != "background" && art.Kind != "cover" && art.Kind != "logo" && art.Kind != "icon" {
				continue
			}
			if attempts[art.Kind] >= 10 {
				continue
			}
			attempts[art.Kind]++
			local, err := s.cacheArtwork(ctx, art.URL, provider.ArtworkHosts())
			if err != nil {
				*diagnostics = append(*diagnostics, game.Title+": "+provider.Name()+" · arte no disponible")
				continue
			}
			// Validate the file, not just dimensions claimed by the remote API.
			data, err := os.ReadFile(filepath.Join(s.artDir, strings.TrimPrefix(local, "/library-assets/")))
			width, height, errSize := artworkDimensions(data)
			if err != nil || errSize != nil || art.Kind == "background" && !BackgroundSize(width, height) {
				*diagnostics = append(*diagnostics, game.Title+": "+provider.Name()+" · imagen inválida o resolución insuficiente")
				continue
			}
			if seen[local] {
				continue
			}
			seen[local] = true
			art.URL = local
			art.Provider = request.ID
			art.Width = width
			art.Height = height
			artworks = append(artworks, art)
			switch art.Kind {
			case "cover":
				covers++
				if !selectedCover {
					metadata.Cover = local
					selectedCover = true
				}
			case "background":
				backgrounds++
				if !selectedBackground {
					metadata.Background = local
					selectedBackground = true
				}
			case "logo":
				logos++
			case "icon":
				icons++
			}
		}
	}
	// A partial refresh (for example, only an icon) must not discard cached backgrounds.
	refreshed := map[string]bool{}
	for _, art := range artworks {
		refreshed[art.Kind] = true
	}
	for _, art := range old.Artworks {
		if !refreshed[art.Kind] {
			artworks = append(artworks, art)
		}
	}
	metadata.Artworks = nil
	for _, art := range artworks {
		if !strings.HasPrefix(art.URL, "/library-assets/") || !assetName.MatchString(strings.TrimPrefix(art.URL, "/library-assets/")) {
			continue
		}
		metadata.Artworks = append(metadata.Artworks, art)
		if art.Kind == "background" && !selectedBackground && BackgroundSize(art.Width, art.Height) {
			metadata.Background = art.URL
			selectedBackground = true
		}
		if art.Kind == "cover" && !selectedCover {
			metadata.Cover = art.URL
			selectedCover = true
		}
	}
	// Choices made in the artwork editor survive later source imports.
	metadata.ManualArt = old.ManualArt
	metadata.Backgrounds = old.Backgrounds
	metadata.Logo = old.Logo
	metadata.Icon = old.Icon
	selectedLogo, selectedIcon := false, false
	for _, art := range metadata.Artworks {
		if art.Kind == "logo" && !selectedLogo {
			metadata.Logo = art.URL
			selectedLogo = true
		}
		if art.Kind == "icon" && !selectedIcon {
			metadata.Icon = art.URL
			selectedIcon = true
		}
	}
	for _, kind := range old.ManualArt {
		switch kind {
		case "cover":
			metadata.Cover = old.Cover
		case "background":
			metadata.Background = old.Background
		case "logo":
			metadata.Logo = old.Logo
		case "icon":
			metadata.Icon = old.Icon
		}
		for _, art := range old.Artworks {
			if art.Kind != kind {
				continue
			}
			found := false
			for _, current := range metadata.Artworks {
				found = found || current.URL == art.URL
			}
			if !found {
				metadata.Artworks = append(metadata.Artworks, art)
			}
		}
	}
	s.includeSceneryCover(metadata)
}

func (s *Service) CheckContentProvider(ctx context.Context, request ContentRequest) error {
	provider := s.content[request.ID]
	if provider == nil {
		return fmt.Errorf("proveedor no compatible")
	}
	return provider.Check(ctx, request.Credentials)
}
