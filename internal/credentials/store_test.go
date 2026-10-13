package credentials

import (
	"errors"
	"strings"
	"testing"
)

func TestVaultValidatesNamesFieldsAndLimitsWithoutEchoingSecrets(t *testing.T) {
	vault := New(t.TempDir())
	for _, request := range []Request{{Provider: "../other", Fields: map[string]string{"apiKey": "secret"}}, {Provider: "steam", Fields: map[string]string{"unknown": "secret"}}, {Provider: "steam", Fields: map[string]string{"apiKey": strings.Repeat("x", 4097)}}, {Provider: "steam", Fields: map[string]string{"apiKey": "secret\x00"}}, {Provider: "steam", Fields: map[string]string{"apiKey": ""}}} {
		err := vault.Save(request)
		if err == nil || strings.Contains(err.Error(), "secret") {
			t.Fatal("invalid credentials accepted or leaked")
		}
	}
	status, err := vault.Status("steam")
	if err != nil || len(status.SavedFields) != 0 {
		t.Fatal("empty vault status incorrect")
	}
	if err = vault.Delete("steam"); err != nil && !(errors.Is(err, ErrUnavailable) && !vault.Capabilities().Available) {
		t.Fatal(err)
	}
}
