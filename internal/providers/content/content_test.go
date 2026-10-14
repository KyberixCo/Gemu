package content

import (
	"context"
	"encoding/json"
	"errors"
	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/library"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSteamGridDBSteamIdentityAndSeparateArtworkRoles(t *testing.T) {
	calls := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		if r.Header.Get("Authorization") != "Bearer test-key" || r.URL.Query().Get("apiKey") != "" {
			t.Error("invalid authentication")
		}
		if r.URL.Path == "/games/steam/42" {
			w.Write([]byte(`{"success":true,"data":{"id":7,"name":"Example"}}`))
			return
		}
		expectedMimes := "image/png,image/jpeg"
		if r.URL.Path == "/logos/game/7" || r.URL.Path == "/icons/game/7" {
			expectedMimes = "image/png"
			if r.URL.Query().Get("mimes") != expectedMimes {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"success":false,"errors":["invalid mime filter"]}`))
				return
			}
		}
		if r.URL.Query().Get("types") != "static" || r.URL.Query().Get("mimes") != expectedMimes {
			t.Error("missing static image filters")
		}
		switch r.URL.Path {
		case "/heroes/game/7":
			w.Write([]byte(`{"success":true,"data":[{"id":1,"url":"https://cdn2.steamgriddb.com/low.jpg","width":460,"height":215,"mime":"image/jpeg","score":100},{"id":2,"url":"https://cdn2.steamgriddb.com/wide.jpg","width":1920,"height":620,"mime":"image/jpeg","score":99},{"id":3,"url":"https://cdn2.steamgriddb.com/good.jpg","width":3840,"height":1240,"mime":"image/jpeg","score":20,"author":{"name":"Artist"}},{"id":4,"url":"https://cdn2.steamgriddb.com/portrait.jpg","width":1280,"height":1920,"mime":"image/jpeg"}]}`))
		case "/grids/game/7":
			if r.URL.Query().Get("dimensions") != "" {
				t.Error("grid dimensions were unnecessarily restricted")
			}
			w.Write([]byte(`{"success":true,"data":[{"id":5,"url":"https://cdn2.steamgriddb.com/cover.png","width":600,"height":900,"mime":"image/png"},{"id":6,"url":"https://cdn2.steamgriddb.com/wide-grid.png","width":1920,"height":1080,"mime":"image/png"}]}`))
		case "/logos/game/7":
			w.Write([]byte(`{"success":true,"data":[{"id":6,"url":"https://cdn2.steamgriddb.com/logo-ar.png","width":1000,"height":300,"mime":"image/png","language":"ar"},{"id":7,"url":"https://cdn2.steamgriddb.com/logo.png","width":1000,"height":300,"mime":"image/png","language":"en"}]}`))
		case "/icons/game/7":
			w.Write([]byte(`{"success":true,"data":[{"id":8,"url":"https://cdn2.steamgriddb.com/icon.png","width":256,"height":256,"mime":"image/png"}]}`))
		default:
			w.Write([]byte(`{"success":true,"data":[]}`))
		}
	}))
	defer server.Close()
	provider := NewSteamGridDB()
	provider.baseURL = server.URL
	result, err := provider.FetchLocalized(context.Background(), domain.Game{Title: "Example", Installations: []domain.Installation{{Provider: "Steam", ExternalID: "42"}}}, map[string]string{"apiKey": "test-key"}, "en")
	if err != nil || len(result.Artworks) != 5 || len(result.Diagnostics) != 0 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if result.Artworks[0].Kind != "background" || result.Artworks[0].Author != "Artist" || result.Artworks[1].Kind != "cover" {
		t.Fatalf("incorrect roles: %+v", result.Artworks)
	}
	if result.Artworks[3].Kind != "logo" || result.Artworks[4].Kind != "icon" {
		t.Fatal("logo or icon omitted")
	}
	if len(calls) != 5 || calls[0] != "/games/steam/42" {
		t.Fatalf("requests=%v", calls)
	}
}

func TestSteamGridDBLanguageMatchingKeepsUnlabelledAssetsAndFiltersKnownMismatches(t *testing.T) {
	for _, test := range []struct {
		asset, preferred string
		want             bool
	}{
		{"Arabic", "en", false},
		{"ar", "es", false},
		{"English", "en", true},
		{"en-US", "en", true},
		{"Spanish (Spain)", "es", true},
		{"es", "es", true},
		{"", "en", true},
		{"unrecognized", "en", true},
		{"Arabic", "", true},
	} {
		if got := languageMatches(test.asset, test.preferred); got != test.want {
			t.Errorf("languageMatches(%q, %q) = %v, want %v", test.asset, test.preferred, got, test.want)
		}
	}
}

func TestSteamGridDBRejectsAmbiguousSearchAndStopsOnQuota(t *testing.T) {
	status := 200
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if status != 200 {
			w.WriteHeader(status)
			w.Write([]byte("secret from server"))
			return
		}
		w.Write([]byte(`{"success":true,"data":[{"id":1,"name":"Example"},{"id":2,"name":"Example"}]}`))
	}))
	defer server.Close()
	provider := NewSteamGridDB()
	provider.baseURL = server.URL
	_, err := provider.Fetch(context.Background(), domain.Game{Title: "Example"}, map[string]string{"apiKey": "test-key"})
	if err == nil || calls != 1 {
		t.Fatal("ambiguous game was assigned artwork")
	}
	status = 429
	err = provider.Check(context.Background(), map[string]string{"apiKey": "test-key"})
	if !errors.Is(err, library.ErrContentUnavailable) || strings.Contains(err.Error(), "secret") {
		t.Fatalf("quota error=%v", err)
	}
}

func TestScreenScraperMatchesPCOnlyAndSelectsSpanishMetadata(t *testing.T) {
	credentials := map[string]string{"developerId": "dev", "developerPassword": "test-password", "username": "player", "password": "player-password"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("devid") != "dev" || q.Get("devpassword") != "test-password" || q.Get("output") != "json" || q.Get("softname") != "Gemu" {
			t.Error("missing ScreenScraper parameters")
		}
		switch r.URL.Path {
		case "/systemesListe.php":
			w.Write([]byte(`{"response":{"systemes":[{"id":"138","noms":{"nom_eu":"PC Windows"}}]}}`))
		case "/jeuRecherche.php":
			if q.Get("systemeid") != "138" || q.Get("recherche") != "Example" {
				t.Error("incorrect search scope")
			}
			w.Write([]byte(`{"response":{"jeux":[{"id":"8","noms":[{"region":"wor","text":"Example"}],"systeme":{"id":"1"}},{"id":"9","noms":[{"region":"wor","text":"Example"}],"systeme":{"id":"138"},"synopsis":[{"langue":"en","text":"English"},{"langue":"es","text":"Spanish"}],"genres":[{"noms":[{"langue":"es","text":"Adventure"}]}],"medias":[{"type":"fanart","url":"https://www.screenscraper.fr/good.jpg","format":"jpg","width":"1920","height":"1080"},{"type":"ss","url":"https://www.screenscraper.fr/screenshot.jpg","format":"jpg","width":320,"height":240},{"type":"box-2D","url":"https://www.screenscraper.fr/cover.png","format":"png","width":600,"height":900},{"type":"wheel-hd","url":"https://api.screenscraper.fr/api2/mediaJeu.php?gameid=9&media=wheel-hd","format":"png","width":1000,"height":400}]}]}}`))
		case "/ssuserInfos.php":
			w.Write([]byte(`{"response":{"ssuser":{"id":"player"}}}`))
		default:
			t.Errorf("unexpected endpoint %s", r.URL.Path)
		}
	}))
	defer server.Close()
	provider := testScreenScraper()
	provider.baseURL = server.URL
	if err := provider.Check(context.Background(), credentials); err != nil {
		t.Fatal(err)
	}
	result, err := provider.Fetch(context.Background(), domain.Game{Title: "Example"}, credentials)
	if err != nil || result.Description != "Spanish" || len(result.Genres) != 1 || len(result.Artworks) != 3 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if result.Artworks[0].Kind != "background" || result.Artworks[1].Kind != "cover" || result.Artworks[2].Kind != "logo" {
		t.Fatalf("roles=%+v", result.Artworks)
	}
	if !strings.Contains(result.Artworks[2].URL, "devpassword=test-password") {
		t.Fatal("media endpoint missing authentication")
	}
	if strings.Contains(result.Artworks[2].Page, "password") {
		t.Fatal("credentials leaked into provenance")
	}
}

func TestScreenScraperAmbiguityInvalidCredentialsAndRedactedErrors(t *testing.T) {
	provider := testScreenScraper()
	provider.developerID = ""
	if err := provider.Check(context.Background(), nil); !errors.Is(err, library.ErrContentUnavailable) {
		t.Fatal("missing developer credentials accepted")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/systemesListe.php" {
			w.Write([]byte(`{"response":{"systemes":[{"id":138,"nom":"PC Windows"}]}}`))
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"jeux": []any{map[string]any{"id": 1, "nom": "Example", "systeme": map[string]any{"id": 138}}, map[string]any{"id": 2, "nom": "Example", "systeme": map[string]any{"id": 138}}}}})
	}))
	defer server.Close()
	provider.baseURL = server.URL
	provider.developerID = "dev"
	_, err := provider.Fetch(context.Background(), domain.Game{Title: "Example"}, map[string]string{"developerId": "dev", "developerPassword": "secret"})
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("err=%v", err)
	}
}

func TestTransportErrorsNeverExposeCredentialQuery(t *testing.T) {
	provider := testScreenScraper()
	provider.baseURL = ":invalid"
	err := provider.Check(context.Background(), map[string]string{"developerId": "dev", "developerPassword": "secret"})
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("error=%v", err)
	}
}
