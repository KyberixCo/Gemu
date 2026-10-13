package credentials

import (
	"errors"
	"testing"
)

func TestUnsupportedBackendAllowsTemporaryUseAndNeverPretendsToPersist(t *testing.T) {
	var vault Vault = unsupportedVault{}
	if vault.Capabilities().Available {
		t.Fatal("unsupported backend claims secure storage")
	}
	status, err := vault.Status("steam")
	if err != nil || status.Storage.Available || len(status.SavedFields) != 0 {
		t.Fatal("unsupported status incorrect")
	}
	values, err := vault.Get("steam")
	if err != nil || len(values) != 0 {
		t.Fatal("session-only use should not require a secure backend")
	}
	if err = vault.Save(Request{Provider: "steam", Fields: map[string]string{"apiKey": "dummy"}}); !errors.Is(err, ErrUnavailable) {
		t.Fatal("unsupported backend accepted persistence")
	}
	if err = vault.Delete("steam"); !errors.Is(err, ErrUnavailable) {
		t.Fatal("unsupported backend pretended to delete saved secrets")
	}
	if _, err = vault.Get("../invalid"); err == nil {
		t.Fatal("invalid provider accepted")
	}
}
