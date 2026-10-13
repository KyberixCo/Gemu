package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/library"
)

type Provider struct {
	Client           *http.Client
	storeURL, apiURL string
}

func (p *Provider) Name() string { return "Steam" }
func (p *Provider) ArtworkHosts() []string {
	return []string{"shared.fastly.steamstatic.com", "shared.akamai.steamstatic.com", "shared.cloudflare.steamstatic.com", "cdn.akamai.steamstatic.com", "cdn.cloudflare.steamstatic.com", "steamcdn-a.akamaihd.net"}
}
func NewProvider() *Provider {
	return &Provider{Client: &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error { return fmt.Errorf("redirección rechazada") }}, storeURL: "https://store.steampowered.com", apiURL: "https://api.steampowered.com"}
}
func (p *Provider) Discover(ctx context.Context, root string) (library.Preview, error) {
	if err := ctx.Err(); err != nil {
		return library.Preview{}, err
	}
	report, err := Inspect(root)
	if err != nil {
		return library.Preview{}, err
	}
	profiles, err := LocalProfiles(report.Root)
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, "No se pudieron leer los perfiles locales de Steam.")
	}
	return library.Preview{Root: report.Root, Games: report.Games, Profiles: profiles, Diagnostics: report.Diagnostics}, nil
}
func LocalProfiles(root string) ([]library.Profile, error) {
	data, err := os.ReadFile(filepath.Join(root, "config", "loginusers.vdf"))
	if os.IsNotExist(err) {
		if detected, e := DetectRoot(); e == nil && !strings.EqualFold(detected, root) {
			data, err = os.ReadFile(filepath.Join(detected, "config", "loginusers.vdf"))
		}
	}
	if os.IsNotExist(err) {
		return []library.Profile{}, nil
	}
	if err != nil {
		return []library.Profile{}, err
	}
	tree, err := parse(string(data))
	if err != nil {
		return []library.Profile{}, err
	}
	profiles := []library.Profile{}
	for id, user := range tree["users"].children {
		if len(id) != 17 {
			continue
		}
		if _, err := strconv.ParseUint(id, 10, 64); err != nil {
			continue
		}
		name := user.children["personaname"].value
		if name == "" {
			name = "Jugador Steam"
		}
		profiles = append(profiles, library.Profile{ID: id, Name: name, Origin: "local"})
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })
	return profiles, nil
}
func (p *Provider) get(ctx context.Context, address string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return fmt.Errorf("solicitud inválida")
	}
	resp, err := p.Client.Do(req)
	if err != nil {
		return fmt.Errorf("servicio no disponible")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("servicio HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 4*1024*1024)).Decode(out)
}

var markup = regexp.MustCompile(`<[^>]*>`)

func (p *Provider) Metadata(ctx context.Context, game domain.Game) (library.Metadata, error) {
	id := strings.TrimPrefix(game.ID, "steam:")
	if !ValidAppID(id) {
		return library.Metadata{}, fmt.Errorf("AppID inválido")
	}
	var result map[string]struct {
		Success bool
		Data    struct {
			ShortDescription string `json:"short_description"`
			HeaderImage      string `json:"header_image"`
			Background       string
			BackgroundRaw    string `json:"background_raw"`
			Screenshots      []struct {
				PathFull string `json:"path_full"`
			}
			Genres []struct{ Description string }
		}
	}
	if err := p.get(ctx, p.storeURL+"/api/appdetails?appids="+id+"&l=spanish", &result); err != nil {
		return library.Metadata{}, err
	}
	app, ok := result[id]
	if !ok || !app.Success {
		return library.Metadata{}, fmt.Errorf("metadatos no disponibles")
	}
	genres := []string{}
	for _, g := range app.Data.Genres {
		genres = append(genres, g.Description)
	}
	background := app.Data.BackgroundRaw
	if background == "" && len(app.Data.Screenshots) > 0 {
		background = app.Data.Screenshots[0].PathFull
	}
	return library.Metadata{Description: html.UnescapeString(markup.ReplaceAllString(app.Data.ShortDescription, "")), Genres: genres, Cover: app.Data.HeaderImage, Background: background, Origin: "steam-store", Status: "ready"}, nil
}
func (p *Provider) api(ctx context.Context, path string, values url.Values, out interface{}) error {
	return p.get(ctx, p.apiURL+path+"?"+values.Encode(), out)
}
func (p *Provider) Player(ctx context.Context, game domain.Game, profile library.Profile, key string) library.PlayerState {
	state := library.PlayerState{Profile: profile, Status: "pending", Achievements: []library.Achievement{}}
	if profile.ID == "" {
		state.Message = "Selecciona un perfil para sincronizar logros."
		return state
	}
	if strings.TrimSpace(key) == "" {
		state.Message = "Perfil local asociado. Falta una clave API para consultar logros."
		return state
	}
	var summary struct {
		Response struct {
			Players []struct {
				SteamID     string
				PersonaName string
			}
		}
	}
	if err := p.api(ctx, "/ISteamUser/GetPlayerSummaries/v2/", url.Values{"key": {key}, "steamids": {profile.ID}}, &summary); err == nil {
		for _, player := range summary.Response.Players {
			if player.SteamID == profile.ID {
				state.Profile.Name = player.PersonaName
				state.Profile.Origin = "steam-api"
			}
		}
	}
	id := strings.TrimPrefix(game.ID, "steam:")
	values := url.Values{"key": {key}, "appid": {id}, "steamid": {profile.ID}, "l": {"spanish"}}
	var result struct {
		PlayerStats struct {
			Success      bool
			Achievements []struct {
				APIName     string
				Name        string
				Description string
				Achieved    int
				UnlockTime  int64
			}
		} `json:"playerstats"`
	}
	if err := p.api(ctx, "/ISteamUserStats/GetPlayerAchievements/v1/", values, &result); err != nil || !result.PlayerStats.Success {
		state.Status = "unavailable"
		state.Message = "Logros no disponibles por API. Se conserva el perfil local; no hay un registro local fiable de logros."
		return state
	}
	var schema struct {
		Game struct {
			AvailableGameStats struct {
				Achievements []struct {
					Name        string
					DisplayName string
					Description string
				}
			}
		}
	}
	schemaErr := p.api(ctx, "/ISteamUserStats/GetSchemaForGame/v2/", values, &schema)
	for _, a := range result.PlayerStats.Achievements {
		name, description := a.Name, a.Description
		if schemaErr == nil {
			for _, definition := range schema.Game.AvailableGameStats.Achievements {
				if definition.Name == a.APIName {
					name = definition.DisplayName
					description = definition.Description
					break
				}
			}
		}
		if name == "" {
			name = a.APIName
		}
		state.Achievements = append(state.Achievements, library.Achievement{ID: a.APIName, Name: name, Description: description, Unlocked: a.Achieved == 1, UnlockedAt: a.UnlockTime})
	}
	state.Status = "ready"
	state.Message = "Logros sincronizados desde Steam."
	return state
}
