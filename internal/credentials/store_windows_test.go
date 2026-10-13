//go:build windows

package credentials

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestWindowsVaultEncryptionRestartMergeAndRemoval(t *testing.T) {
	dir := t.TempDir()
	vault := New(dir)
	request := Request{Provider: "screenscraper", Fields: map[string]string{"username": "dummy-user", "password": "dummy-secret-only-for-test", "developerId": "dummy-dev"}}
	if err := vault.Save(request); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "credentials.bin"))
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range request.Fields {
		if bytes.Contains(data, []byte(secret)) {
			t.Fatal("secret written in plaintext")
		}
	}
	reloaded := New(dir)
	values, err := reloaded.Get("screenscraper")
	if err != nil || values["password"] != request.Fields["password"] {
		t.Fatal("could not recover encrypted credential after restart")
	}
	if err = reloaded.Save(Request{Provider: "screenscraper", Fields: map[string]string{"username": "changed-user", "password": ""}}); err != nil {
		t.Fatal(err)
	}
	values, err = reloaded.Get("screenscraper")
	if err != nil || values["username"] != "changed-user" || values["password"] != request.Fields["password"] {
		t.Fatal("partial update erased secret")
	}
	status, err := reloaded.Status("screenscraper")
	if err != nil || len(status.SavedFields) != 3 {
		t.Fatal("incorrect saved-field status")
	}
	if err = reloaded.Delete("screenscraper"); err != nil {
		t.Fatal(err)
	}
	values, err = New(dir).Get("screenscraper")
	if err != nil || len(values) != 0 {
		t.Fatal("deleted credentials recovered")
	}
}

func TestWindowsVaultProfileIsolationAndTampering(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	vault := New(first)
	if err := vault.Save(Request{Provider: "steamgriddb", Fields: map[string]string{"apiKey": "dummy-api-key"}}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(first, "credentials.bin"))
	if err := os.WriteFile(filepath.Join(second, "credentials.bin"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := New(second).Get("steamgriddb"); err == nil {
		t.Fatal("ciphertext from another profile was accepted")
	}
	// Damage the authenticated ciphertext, retaining the file header.
	for i := len(header); i < len(data); i++ {
		data[i] = 0
	}
	os.WriteFile(filepath.Join(first, "credentials.bin"), data, 0600)
	if _, err := vault.Get("steamgriddb"); err == nil {
		t.Fatal("damaged vault was accepted")
	}
	before, _ := os.ReadFile(filepath.Join(first, "credentials.bin"))
	if err := vault.Save(Request{Provider: "steam", Fields: map[string]string{"apiKey": "replacement"}}); err == nil {
		t.Fatal("damaged vault overwritten")
	}
	after, _ := os.ReadFile(filepath.Join(first, "credentials.bin"))
	if !bytes.Equal(before, after) {
		t.Fatal("failed write changed existing vault")
	}
}

func TestWindowsVaultConcurrentProviderUpdates(t *testing.T) {
	vault := New(t.TempDir())
	var wait sync.WaitGroup
	for _, provider := range []string{"steamgriddb", "steam", "openrouter"} {
		wait.Add(1)
		go func(id string) {
			defer wait.Done()
			if err := vault.Save(Request{Provider: id, Fields: map[string]string{"apiKey": "dummy-" + id}}); err != nil {
				t.Error(err)
			}
		}(provider)
	}
	wait.Wait()
	for _, provider := range []string{"steamgriddb", "steam", "openrouter"} {
		values, err := vault.Get(provider)
		if err != nil || values["apiKey"] != "dummy-"+provider {
			t.Fatal("concurrent update lost provider credentials")
		}
	}
}
