package steam

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/library"
)

func TestPlayerRequestsSelectedLanguageForAchievementsAndSchema(t *testing.T) {
	for _, test := range []struct{ language, want string }{{"en", "english"}, {"es", "spanish"}} {
		t.Run(test.language, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/ISteamUserStats/GetPlayerAchievements/v1/":
					requests++
					if r.URL.Query().Get("l") != test.want {
						t.Errorf("achievement language = %q", r.URL.Query().Get("l"))
					}
					w.Write([]byte(`{"playerstats":{"success":true,"achievements":[{"apiname":"FIRST","achieved":1}]}}`))
				case "/ISteamUserStats/GetSchemaForGame/v2/":
					requests++
					if r.URL.Query().Get("l") != test.want {
						t.Errorf("schema language = %q", r.URL.Query().Get("l"))
					}
					w.Write([]byte(`{"game":{"availablegamestats":{"achievements":[{"name":"FIRST","displayName":"Localized achievement","description":"Localized description"}]}}}`))
				default:
					w.Write([]byte(`{"response":{"players":[]}}`))
				}
			}))
			defer server.Close()
			p := NewProvider()
			p.Client = server.Client()
			p.apiURL = server.URL
			state := p.PlayerLocalized(context.Background(), domain.Game{ID: "steam:42"}, library.Profile{ID: "player"}, "key", test.language)
			if requests != 2 || state.Status != "ready" || len(state.Achievements) != 1 || state.Achievements[0].Name != "Localized achievement" {
				t.Fatalf("requests=%d state=%+v", requests, state)
			}
		})
	}
}
