package content

import (
	"context"
	"encoding/json"
	"fmt"
	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/library"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

type ScreenScraper struct {
	Client            *http.Client
	baseURL           string
	mu                sync.Mutex
	windowsID         string
	developerID       string
	developerPassword string
	softwareName      string
}

func NewScreenScraper() *ScreenScraper {
	return &ScreenScraper{Client: client(), baseURL: "https://api.screenscraper.fr/api2", developerID: screenScraperDeveloperID, developerPassword: screenScraperDeveloperPassword, softwareName: screenScraperSoftwareName}
}
func (*ScreenScraper) Name() string { return "ScreenScraper" }
func (*ScreenScraper) ArtworkHosts() []string {
	return []string{"www.screenscraper.fr", "screenscraper.fr", "api.screenscraper.fr", "medias.screenscraper.fr"}
}
func (p *ScreenScraper) query(credentials map[string]string) (url.Values, error) {
	if p.developerID == "" || p.developerPassword == "" || p.softwareName == "" {
		return nil, fmt.Errorf("esta compilación de Gemu no incluye acceso de aplicación a ScreenScraper; el responsable debe configurarlo al compilar: %w", library.ErrContentUnavailable)
	}
	if (credentials["username"] == "") != (credentials["password"] == "") {
		return nil, fmt.Errorf("completa usuario y contraseña: %w", library.ErrContentUnavailable)
	}
	return url.Values{"devid": {p.developerID}, "devpassword": {p.developerPassword}, "softname": {p.softwareName}, "output": {"json"}, "ssid": {credentials["username"]}, "sspassword": {credentials["password"]}}, nil
}
func (p *ScreenScraper) get(ctx context.Context, endpoint string, query url.Values, out any) error {
	return getJSON(ctx, p.Client, p.baseURL+"/"+endpoint+"?"+query.Encode(), "", out, screenStatusError)
}
func screenStatusError(code int) error {
	message := ""
	switch code {
	case 400:
		return fmt.Errorf("ScreenScraper rechazó la configuración de la consulta (HTTP 400)")
	case 401:
		message = "ScreenScraper restringió el acceso (HTTP 401). Comprueba la cuenta personal; el servicio puede estar saturado"
	case 403:
		message = "ScreenScraper rechazó las credenciales de desarrollador de Gemu (HTTP 403). No uses aquí tu usuario y contraseña personales"
	case 423:
		message = "ScreenScraper está temporalmente cerrado (HTTP 423). Inténtalo más tarde"
	case 426:
		message = "ScreenScraper bloqueó el acceso de la aplicación (HTTP 426). El responsable de Gemu debe revisar su registro con ScreenScraper"
	case 429:
		message = "ScreenScraper está ocupado o alcanzó el límite de consultas (HTTP 429). Espera antes de volver a probar"
	case 430, 431:
		message = "Se alcanzó la cuota diaria de ScreenScraper. Vuelve a intentarlo cuando el servicio renueve la cuota"
	}
	if message != "" {
		return fmt.Errorf("%s: %w", message, library.ErrContentUnavailable)
	}
	return nil
}
func (p *ScreenScraper) Check(ctx context.Context, credentials map[string]string) error {
	query, err := p.query(credentials)
	if err != nil {
		return err
	}
	var response struct {
		Response struct {
			User    struct{ ID textNumber } `json:"ssuser"`
			Servers json.RawMessage         `json:"serveurs"`
		}
	}
	endpoint := "ssinfraInfos.php"
	if credentials["username"] != "" {
		endpoint = "ssuserInfos.php"
	}
	if err = p.get(ctx, endpoint, query, &response); err != nil {
		return err
	}
	if credentials["username"] != "" && response.Response.User.ID == "" || credentials["username"] == "" && (len(response.Response.Servers) == 0 || string(response.Response.Servers) == "null") {
		return fmt.Errorf("la API no confirmó el acceso")
	}
	_, err = p.system(ctx, query)
	return err
}
func (p *ScreenScraper) system(ctx context.Context, query url.Values) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.windowsID != "" {
		return p.windowsID, nil
	}
	var response struct {
		Response struct {
			Systems []struct {
				ID    textNumber
				Name  string            `json:"nom"`
				Names map[string]string `json:"noms"`
			} `json:"systemes"`
		}
	}
	if err := p.get(ctx, "systemesListe.php", query, &response); err != nil {
		return "", err
	}
	var windowsID string
	for _, system := range response.Response.Systems {
		matches := normalizedTitle(system.Name) == "pcwindows"
		for _, name := range system.Names {
			matches = matches || normalizedTitle(name) == "pcwindows"
		}
		if matches && system.ID != "" {
			if windowsID != "" && windowsID != string(system.ID) {
				return "", fmt.Errorf("varios sistemas PC Windows en el catálogo de ScreenScraper")
			}
			windowsID = string(system.ID)
		}
	}
	if windowsID != "" {
		p.windowsID = windowsID
		return windowsID, nil
	}
	return "", fmt.Errorf("sistema PC Windows no encontrado")
}

type localizedText struct {
	Language string `json:"langue"`
	Region   string `json:"region"`
	Text     string `json:"text"`
}

func localized(values []localizedText) string {
	for _, language := range []string{"es", "en", "fr"} {
		for _, value := range values {
			if value.Language == language && value.Text != "" {
				return value.Text
			}
		}
	}
	if len(values) > 0 {
		return values[0].Text
	}
	return ""
}

type screenGame struct {
	ID       textNumber
	Name     string                  `json:"nom"`
	Names    []localizedText         `json:"noms"`
	System   struct{ ID textNumber } `json:"systeme"`
	Synopsis []localizedText
	Genres   []struct {
		Names []localizedText `json:"noms"`
	}
	Medias []struct {
		Type, URL, Format, Region string
		Width, Height             textNumber
	}
}

func (p *ScreenScraper) Fetch(ctx context.Context, game domain.Game, credentials map[string]string) (library.ContentResult, error) {
	result := library.ContentResult{}
	query, err := p.query(credentials)
	if err != nil {
		return result, err
	}
	// The current import adapters are PC libraries. Future ROM adapters must supply
	// their own system mapping rather than searching across unrelated platforms.
	system, err := p.system(ctx, query)
	if err != nil {
		return result, err
	}
	query.Set("systemeid", system)
	query.Set("recherche", game.Title)
	var response struct {
		Response struct {
			Games []screenGame `json:"jeux"`
		}
	}
	if err = p.get(ctx, "jeuRecherche.php", query, &response); err != nil {
		return result, err
	}
	var match *screenGame
	for i := range response.Response.Games {
		candidate := &response.Response.Games[i]
		if string(candidate.System.ID) != system {
			continue
		}
		exact := normalizedTitle(candidate.Name) == normalizedTitle(game.Title)
		for _, name := range candidate.Names {
			exact = exact || normalizedTitle(name.Text) == normalizedTitle(game.Title)
		}
		if !exact {
			continue
		}
		if match != nil {
			return result, fmt.Errorf("varias coincidencias; no se asignaron datos")
		}
		match = candidate
	}
	if match == nil {
		return result, fmt.Errorf("sin coincidencia exacta para PC Windows")
	}
	result.Description = localized(match.Synopsis)
	for _, genre := range match.Genres {
		if name := localized(genre.Names); name != "" {
			result.Genres = append(result.Genres, name)
		}
	}
	for i, media := range match.Medias {
		kind := ""
		switch media.Type {
		case "fanart":
			kind = "background"
		case "box-2D", "steamgrid":
			kind = "cover"
		case "wheel-hd", "wheel":
			kind = "logo"
		}
		if kind == "" || media.Format != "png" && media.Format != "jpg" && media.Format != "jpeg" {
			continue
		}
		width, _ := strconv.Atoi(string(media.Width))
		height, _ := strconv.Atoi(string(media.Height))
		if kind == "background" && width > 0 && height > 0 && !library.BackgroundSize(width, height) {
			continue
		}
		address := media.URL
		// Media endpoints may require the same authentication. Never attach it to
		// arbitrary URLs supplied in the response.
		if u, err := url.Parse(address); err == nil && u.Scheme == "https" && u.Host == "api.screenscraper.fr" && u.Path == "/api2/mediaJeu.php" {
			q := u.Query()
			for _, key := range []string{"devid", "devpassword", "softname", "ssid", "sspassword"} {
				q.Set(key, query.Get(key))
			}
			u.RawQuery = q.Encode()
			address = u.String()
		}
		result.Artworks = append(result.Artworks, library.Artwork{ID: fmt.Sprintf("screenscraper-%s-%s-%d", match.ID, kind, i), Kind: kind, URL: address, Provider: "screenscraper", Page: "https://www.screenscraper.fr/gameinfos.php?gameid=" + url.QueryEscape(string(match.ID)), Width: width, Height: height})
	}
	return result, nil
}
