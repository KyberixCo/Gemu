//go:build windows

package main

import (
	"context"
	"encoding/json"
	"kyberix.com/gemu/internal/credentials"
	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/library"
	"strings"
	"testing"
)

type credentialContent struct{ received map[string]string }

func (*credentialContent) Name() string           { return "Test" }
func (*credentialContent) ArtworkHosts() []string { return nil }
func (c *credentialContent) Check(_ context.Context, values map[string]string) error {
	c.received = values
	return nil
}
func (c *credentialContent) Fetch(_ context.Context, _ domain.Game, values map[string]string) (library.ContentResult, error) {
	c.received = values
	return library.ContentResult{}, nil
}

type credentialSource struct {
	profileSource
	key string
}

func (s *credentialSource) Player(_ context.Context, _ domain.Game, _ library.Profile, key string) library.PlayerState {
	s.key = key
	return library.PlayerState{Status: "pending"}
}

func TestSavedCredentialsAreResolvedOnlyInBackendAndSessionOverridesAreNotSaved(t *testing.T) {
	dir := t.TempDir()
	vault := credentials.New(dir)
	provider := &credentialContent{}
	source := &credentialSource{}
	app := &App{vault: vault, library: library.New(dir, map[string]library.Source{"steam": source}).WithContentProviders(map[string]library.ContentProvider{"screenscraper": provider})}
	if err := app.SaveProviderCredentials(credentials.Request{Provider: "screenscraper", Fields: map[string]string{"username": "dummy-user", "password": "dummy-saved-password"}}); err != nil {
		t.Fatal(err)
	}
	if err := app.SaveProviderCredentials(credentials.Request{Provider: "steam", Fields: map[string]string{"apiKey": "dummy-steam-key"}}); err != nil {
		t.Fatal(err)
	}
	if err := app.CheckContentProvider(library.ContentRequest{ID: "screenscraper", Credentials: map[string]string{"username": "", "password": "dummy-session-password"}}); err != nil {
		t.Fatal(err)
	}
	if provider.received["username"] != "dummy-user" || provider.received["password"] != "dummy-session-password" {
		t.Fatal("saved credentials were not merged with session overrides")
	}
	stored, _ := vault.Get("screenscraper")
	if stored["password"] != "dummy-saved-password" {
		t.Fatal("temporary credential persisted automatically")
	}
	draft, err := app.PrepareLibraryImport(library.Request{SourceID: "steam", ContentProviders: []library.ContentRequest{{ID: "screenscraper"}}})
	if err != nil || source.key != "dummy-steam-key" || provider.received["password"] != "dummy-saved-password" {
		t.Fatal("import did not resolve stored secrets")
	}
	status, err := app.GetCredentialStatus("screenscraper")
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []any{draft, status} {
		data, _ := json.Marshal(value)
		if strings.Contains(string(data), "dummy-saved-password") || strings.Contains(string(data), "dummy-steam-key") {
			t.Fatal("secret exposed to frontend")
		}
	}
}
