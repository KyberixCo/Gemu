package content

import (
	"context"
	"errors"
	"fmt"
	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/library"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type SteamGridDB struct {
	Client  *http.Client
	baseURL string
}

func NewSteamGridDB() *SteamGridDB {
	return &SteamGridDB{Client: client(), baseURL: "https://www.steamgriddb.com/api/v2"}
}
func (*SteamGridDB) Name() string { return "SteamGridDB" }
func (*SteamGridDB) ArtworkHosts() []string {
	return []string{"cdn2.steamgriddb.com", "cdn.steamgriddb.com"}
}

type gridGame struct {
	ID   int
	Name string
}
type gridAsset struct {
	ID                   int
	URL                  string
	Width, Height, Score int
	NSFW, Humor          bool
	Mime                 string
	Author               struct{ Name string }
}
type gridResponse[T any] struct {
	Success bool
	Data    T
}

func gridKey(credentials map[string]string) (string, error) {
	key := strings.TrimSpace(credentials["apiKey"])
	if key == "" {
		return "", fmt.Errorf("falta la clave API: %w", library.ErrContentUnavailable)
	}
	return key, nil
}
func (p *SteamGridDB) Check(ctx context.Context, credentials map[string]string) error {
	key, err := gridKey(credentials)
	if err != nil {
		return err
	}
	var response gridResponse[[]gridAsset]
	if err = p.get(ctx, "/grids/steam/10?types=static", key, &response); err != nil {
		return err
	}
	if !response.Success {
		return fmt.Errorf("la API no confirmó el acceso")
	}
	return nil
}
func (p *SteamGridDB) get(ctx context.Context, path, key string, out any) error {
	return getJSON(ctx, p.Client, p.baseURL+path, key, out)
}
func (p *SteamGridDB) gameID(ctx context.Context, game domain.Game, key string) (int, error) {
	for _, installation := range game.Installations {
		if !strings.EqualFold(installation.Provider, "Steam") {
			continue
		}
		id := installation.ExternalID
		if _, err := strconv.ParseUint(id, 10, 32); err != nil || id == "0" {
			return 0, fmt.Errorf("AppID de Steam inválido")
		}
		var response gridResponse[gridGame]
		if err := p.get(ctx, "/games/steam/"+id, key, &response); err != nil {
			return 0, err
		}
		if !response.Success || response.Data.ID <= 0 {
			return 0, fmt.Errorf("juego no encontrado")
		}
		return response.Data.ID, nil
	}
	var response gridResponse[[]gridGame]
	if err := p.get(ctx, "/search/autocomplete/"+url.PathEscape(game.Title), key, &response); err != nil {
		return 0, err
	}
	match := 0
	for _, candidate := range response.Data {
		if normalizedTitle(candidate.Name) == normalizedTitle(game.Title) {
			if match != 0 {
				return 0, fmt.Errorf("varias coincidencias; no se asignó arte")
			}
			match = candidate.ID
		}
	}
	if !response.Success || match <= 0 {
		return 0, fmt.Errorf("sin coincidencia exacta")
	}
	return match, nil
}
func (p *SteamGridDB) Fetch(ctx context.Context, game domain.Game, credentials map[string]string) (library.ContentResult, error) {
	result := library.ContentResult{}
	key, err := gridKey(credentials)
	if err != nil {
		return result, err
	}
	id, err := p.gameID(ctx, game, key)
	if err != nil {
		return result, err
	}
	for _, category := range []struct{ endpoint, kind, page string }{{"heroes", "background", "hero"}, {"grids", "cover", "grid"}, {"logos", "logo", "logo"}, {"icons", "icon", "icon"}} {
		query := url.Values{"types": {"static"}, "nsfw": {"false"}, "humor": {"false"}, "mimes": {"image/png,image/jpeg"}}
		// Logos accept PNG/WebP and icons PNG/ICO, never JPEG. Request PNG
		// for these categories so the API and our local image decoder agree.
		if category.kind == "logo" || category.kind == "icon" {
			query.Set("mimes", "image/png")
		}
		var response gridResponse[[]gridAsset]
		err := p.get(ctx, "/"+category.endpoint+"/game/"+strconv.Itoa(id)+"?"+query.Encode(), key, &response)
		if err != nil {
			if errors.Is(err, library.ErrContentUnavailable) {
				return result, err
			}
			result.Diagnostics = append(result.Diagnostics, category.kind+": "+err.Error())
			continue
		}
		if !response.Success {
			result.Diagnostics = append(result.Diagnostics, category.kind+": respuesta sin confirmar")
			continue
		}
		sort.SliceStable(response.Data, func(i, j int) bool {
			a, b := response.Data[i], response.Data[j]
			if a.Score != b.Score {
				return a.Score > b.Score
			}
			return int64(a.Width)*int64(a.Height) > int64(b.Width)*int64(b.Height)
		})
		accepted := 0
		for _, asset := range response.Data {
			if asset.NSFW || asset.Humor || asset.Mime != "image/png" && asset.Mime != "image/jpeg" {
				continue
			}
			if category.kind == "background" && !library.BackgroundSize(asset.Width, asset.Height) {
				continue
			}
			if accepted >= 10 {
				break
			}
			accepted++
			result.Artworks = append(result.Artworks, library.Artwork{ID: fmt.Sprintf("steamgriddb-%s-%d", category.kind, asset.ID), Kind: category.kind, URL: asset.URL, Provider: "steamgriddb", Author: asset.Author.Name, Page: "https://www.steamgriddb.com/" + category.page + "/" + strconv.Itoa(asset.ID), Width: asset.Width, Height: asset.Height})
		}
	}
	return result, nil
}
