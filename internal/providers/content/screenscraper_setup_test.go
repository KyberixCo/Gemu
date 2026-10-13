package content

import (
	"context"
	"errors"
	"fmt"
	"kyberix.com/gemu/internal/library"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestScreenScraperSetupErrorsExplainRequiredActionWithoutResponseSecrets(t *testing.T) {
	for _, code := range []int{401, 403, 423, 426, 429, 430, 431} {
		err := screenStatusError(code)
		if err == nil || !errors.Is(err, library.ErrContentUnavailable) {
			t.Fatalf("status %d not classified", code)
		}
	}
	if !strings.Contains(screenStatusError(403).Error(), "desarrollador de Gemu") {
		t.Fatal("developer credentials confused with personal account")
	}
	if !strings.Contains(screenStatusError(401).Error(), "saturado") {
		t.Fatal("restricted service misreported as invalid password")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte("personal-password-should-never-be-shown"))
	}))
	defer server.Close()
	provider := testScreenScraper()
	provider.baseURL = server.URL
	err := provider.Check(context.Background(), map[string]string{"developerId": "dev", "developerPassword": "secret"})
	if err == nil || strings.Contains(err.Error(), "personal-password") {
		t.Fatal("raw service response leaked")
	}
}

func TestScreenScraperAccessCheckAlsoVerifiesCatalogAccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ssuserInfos.php" {
			w.Write([]byte(`{"response":{"ssuser":{"id":"user"}}}`))
			return
		}
		w.WriteHeader(426)
	}))
	defer server.Close()
	provider := testScreenScraper()
	provider.baseURL = server.URL
	err := provider.Check(context.Background(), map[string]string{"developerId": "dev", "developerPassword": "secret", "username": "user", "password": "password"})
	if err == nil || !strings.Contains(err.Error(), "aplicación") {
		t.Fatal("account-only success incorrectly confirmed catalog access")
	}
}

func TestScreenScraperCheckReadsRegionalSystemNames(t *testing.T) {
	for _, account := range []bool{false, true} {
		t.Run(fmt.Sprint(account), func(t *testing.T) {
			catalogCalls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/ssinfraInfos.php":
					w.Write([]byte(`{"response":{"serveurs":{"cpu1":"0"}}}`))
				case "/ssuserInfos.php":
					w.Write([]byte(`{"response":{"ssuser":{"id":"user"}}}`))
				case "/systemesListe.php":
					catalogCalls++
					w.Write([]byte(`{"response":{"systemes":[{"id":135,"noms":{"nom_eu":"PC Dos"}},{"id":136,"noms":{"nom_eu":"PC Win3.xx","nom_launchbox":"Windows 3.X"}},{"id":137,"noms":{"nom_eu":"PC Win9X"}},{"id":138,"noms":{"nom_eu":"PC Windows","nom_launchbox":"Windows","noms_commun":"Microsoft Windows,Windows XP,Windows 10"}}]}}`))
				default:
					t.Errorf("unexpected endpoint %s", r.URL.Path)
				}
			}))
			defer server.Close()
			provider := testScreenScraper()
			provider.baseURL = server.URL
			credentials := map[string]string{}
			if account {
				credentials = map[string]string{"username": "user", "password": "password"}
			}
			for i := 0; i < 2; i++ {
				if err := provider.Check(context.Background(), credentials); err != nil {
					t.Fatal(err)
				}
			}
			if provider.windowsID != "138" || catalogCalls != 1 {
				t.Fatalf("system=%s catalog calls=%d", provider.windowsID, catalogCalls)
			}
		})
	}
}

func TestScreenScraperRejectsMissingOrAmbiguousWindowsSystem(t *testing.T) {
	for _, systems := range []string{
		`[{"id":135,"noms":{"nom_eu":"PC Dos"}},{"id":136,"noms":{"nom_eu":"PC Win3.xx"}}]`,
		`[{"noms":{"nom_eu":"PC Windows"}}]`,
		`[{"id":138,"noms":{"nom_eu":"PC Windows"}},{"id":999,"noms":{"nom_us":"PC Windows"}}]`,
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"response":{"systemes":%s}}`, systems)
		}))
		provider := testScreenScraper()
		provider.baseURL = server.URL
		query, _ := provider.query(nil)
		_, err := provider.system(context.Background(), query)
		server.Close()
		if err == nil || provider.windowsID != "" {
			t.Fatal("invalid Windows catalog was cached")
		}
	}
}
