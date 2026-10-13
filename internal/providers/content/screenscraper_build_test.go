package content

import (
	"errors"
	"kyberix.com/gemu/internal/library"
	"testing"
)

func testScreenScraper() *ScreenScraper {
	provider := NewScreenScraper()
	provider.developerID = "dev"
	provider.developerPassword = "test-password"
	provider.softwareName = "Gemu"
	return provider
}

func TestScreenScraperApplicationAccessCannotBeOverriddenByProfile(t *testing.T) {
	provider := testScreenScraper()
	provider.softwareName = "Gemu-test"
	values := map[string]string{"developerId": "other-app", "developerPassword": "other-secret", "username": "player", "password": "personal-secret"}
	query, err := provider.query(values)
	if err != nil {
		t.Fatal(err)
	}
	if query.Get("devid") != "dev" || query.Get("devpassword") != "test-password" || query.Get("softname") != "Gemu-test" || query.Get("ssid") != "player" || query.Get("sspassword") != "personal-secret" {
		t.Fatal("application and personal credentials were mixed")
	}
	provider.developerID = ""
	if _, err = provider.query(values); !errors.Is(err, library.ErrContentUnavailable) {
		t.Fatal("profile credentials enabled an unconfigured build")
	}
	provider.developerID = "dev"
	if _, err = provider.query(map[string]string{"username": "player"}); err == nil {
		t.Fatal("incomplete personal account accepted")
	}
}
