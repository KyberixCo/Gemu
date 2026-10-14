package steam

import (
	"context"
	"encoding/base64"
	"image"
	"image/png"
	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/library"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalProfileDoesNotReadAccountSecrets(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "config"), 0755)
	data := `"users" { "76561198000000001" { "PersonaName" "Player" "AccountName" "private-login" "RememberPassword" "1" } }`
	if err := os.WriteFile(filepath.Join(root, "config", "loginusers.vdf"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	profiles, err := LocalProfiles(root)
	if err != nil || len(profiles) != 1 || profiles[0].Name != "Player" || profiles[0].Origin != "local" {
		t.Fatalf("profiles=%+v err=%v", profiles, err)
	}
}
func TestLocalAccountsIncludeCachedAvatarWithoutSecrets(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "config", "avatarcache"), 0755); err != nil {
		t.Fatal(err)
	}
	data := `"users" { "76561198000000001" { "PersonaName" "Player" "AccountName" "private-login" "RememberPassword" "1" } }`
	if err := os.WriteFile(filepath.Join(root, "config", "loginusers.vdf"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	var avatar strings.Builder
	if err := png.Encode(&avatar, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config", "avatarcache", "76561198000000001.jpg"), []byte(avatar.String()), 0600); err != nil {
		t.Fatal(err)
	}
	accounts, err := LocalAccounts(root)
	if err != nil || len(accounts) != 1 || accounts[0].Name != "Player" {
		t.Fatalf("accounts=%+v err=%v", accounts, err)
	}
	if !strings.HasPrefix(accounts[0].Avatar, "data:image/png;base64,") ||
		!strings.Contains(accounts[0].Avatar, base64.StdEncoding.EncodeToString([]byte(avatar.String()))) ||
		strings.Contains(accounts[0].Name, "private-login") {
		t.Fatalf("unexpected account data: %+v", accounts[0])
	}
}
func TestAPIFirstAndLocalFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") == "bad" {
			w.WriteHeader(403)
			return
		}

		switch r.URL.Path {
		case "/ISteamUser/GetPlayerSummaries/v2/":
			w.Write([]byte(`{"response":{"players":[{"steamid":"76561198000000001","personaname":"Remote"}]}}`))
		case "/ISteamUserStats/GetPlayerAchievements/v1/":
			w.Write([]byte(`{"playerstats":{"success":true,"achievements":[{"apiname":"FIRST","achieved":1,"unlocktime":123}]}}`))
		case "/ISteamUserStats/GetSchemaForGame/v2/":
			w.Write([]byte(`{"game":{"availableGameStats":{"achievements":[{"name":"FIRST","displayName":"First step","description":"Finish tutorial"}]}}}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	p := NewProvider()
	p.Client = server.Client()
	p.apiURL = server.URL
	profile := library.Profile{ID: "76561198000000001", Name: "Local", Origin: "local"}
	game := domain.Game{ID: "steam:42"}
	result := p.Player(context.Background(), game, profile, "key")
	if result.Status != "ready" || result.Profile.Name != "Remote" || len(result.Achievements) != 1 || !result.Achievements[0].Unlocked || result.Achievements[0].Name != "First step" {
		t.Fatalf("result=%+v", result)
	}
	for _, key := range []string{"", "bad"} {
		result = p.Player(context.Background(), game, profile, key)
		if result.Profile.Name != "Local" || result.Status == "ready" || len(result.Achievements) != 0 {
			t.Fatalf("fallback=%+v", result)
		}
	}
}

func TestMetadataUsesSelectedLanguageAndLocalizedStoreTitle(t *testing.T) {
	for _, test := range []struct {
		language string
		expected string
		title    string
	}{
		{language: "en", expected: "english", title: "Mouse: P.I. For Hire"},
		{language: "es", expected: "spanish", title: "Mouse: P.I. For Hire"},
	} {
		t.Run(test.language, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("l") != test.expected {
					t.Errorf("language=%q, want %q", r.URL.Query().Get("l"), test.expected)
				}
				if r.URL.Query().Get("appids") != "42" {
					t.Errorf("appids=%q", r.URL.Query().Get("appids"))
				}
				_, _ = w.Write([]byte(`{"42":{"success":true,"data":{"name":"Mouse: P.I. For Hire","short_description":"A game","header_image":"https://example.com/cover.jpg"}}}`))
			}))
			defer server.Close()

			provider := NewProvider()
			provider.Client = server.Client()
			provider.storeURL = server.URL
			metadata, err := provider.Metadata(context.Background(), domain.Game{ID: "steam:42", Title: "عنوان عربي"}, test.language)
			if err != nil {
				t.Fatal(err)
			}
			if metadata.Title != test.title {
				t.Fatalf("title=%q, want %q", metadata.Title, test.title)
			}
		})
	}
}
