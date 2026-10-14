package content

import "testing"

func TestLocalizedTextPrefersRequestedLanguage(t *testing.T) {
	values := []localizedText{{Language: "es", Text: "Español"}, {Language: "en", Text: "English"}}
	for _, test := range []struct{ language, want string }{{"es", "Español"}, {"en", "English"}, {"en-US", "English"}, {"fr", "English"}} {
		if got := localized(values, test.language); got != test.want {
			t.Errorf("localized(%q) = %q, want %q", test.language, got, test.want)
		}
	}
}

func TestLocalizedTextSkipsEmptyFallbacks(t *testing.T) {
	values := []localizedText{{Language: "en"}, {Language: "ja", Text: "日本語"}}
	if got := localized(values, "en"); got != "日本語" {
		t.Fatalf("got %q, want first nonempty fallback", got)
	}
}
