package main

import (
	"context"
	"fmt"
	"kyberix.com/gemu/internal/credentials"
	"kyberix.com/gemu/internal/library"
	"kyberix.com/gemu/internal/profiles"
	"kyberix.com/gemu/internal/providers/content"
	"kyberix.com/gemu/internal/providers/steam"
	"path/filepath"
	"sync"
	"time"
)

// Each selected Gemu profile owns its library and Steam association.
type App struct {
	library  *library.Service
	profiles *profiles.Store
	sources  map[string]library.Source
	dir      string
	mu       sync.Mutex
	vault    credentials.Vault
}

func (a *App) GetProfiles() ([]profiles.Profile, error)            { return a.profiles.List() }
func (a *App) CreateProfile(name string) (profiles.Profile, error) { return a.profiles.Create(name) }
func (a *App) SelectProfile(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	dir, err := a.profiles.Directory(id)
	if err != nil {
		return err
	}
	if a.library != nil {
		a.library.CancelArtworkSearches()
	}
	a.library = library.New(dir, a.sources).WithArtworkDirectory(filepath.Join(a.dir, "art")).WithContentProviders(map[string]library.ContentProvider{
		"steamgriddb": content.NewSteamGridDB(), "screenscraper": content.NewScreenScraper(),
	})
	a.vault = credentials.New(dir)
	return nil
}
func (a *App) DiscoverSteam(root string) (steam.Report, error) { return steam.Inspect(root) }

// Capture the session before slow I/O so switching players does not wait for
// a metadata request. Work already in flight remains owned by its old session.
func (a *App) session() (*library.Service, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.library == nil {
		return nil, fmt.Errorf("selecciona un perfil de Gemu")
	}
	return a.library, nil
}
func (a *App) PreviewSource(sourceID, root string) (library.Preview, error) {
	session, err := a.session()
	if err != nil {
		return library.Preview{}, err
	}
	return session.Preview(context.Background(), sourceID, root)
}
func (a *App) PrepareLibraryImport(request library.Request) (library.Draft, error) {
	session, vault, err := a.credentialSession()
	if err != nil {
		return library.Draft{}, err
	}
	for i := range request.ContentProviders {
		values, err := resolvedCredentials(vault, request.ContentProviders[i].ID, request.ContentProviders[i].Credentials)
		if err != nil {
			return library.Draft{}, err
		}
		request.ContentProviders[i].Credentials = values
	}
	if request.SourceID == "steam" && request.APIKey == "" && vault != nil {
		values, err := vault.Get("steam")
		if err != nil {
			return library.Draft{}, err
		}
		request.APIKey = values["apiKey"]
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return session.Prepare(ctx, request)
}
func (a *App) CommitLibraryImport(token string) (library.Snapshot, error) {
	session, err := a.session()
	if err != nil {
		return library.Snapshot{}, err
	}
	return session.Commit(token)
}
func (a *App) DiscardLibraryImport(token string) {
	if session, err := a.session(); err == nil {
		session.Discard(token)
	}
}
func (a *App) GetLibrary() (library.Snapshot, error) {
	session, err := a.session()
	if err != nil {
		return library.Snapshot{}, err
	}
	return session.Load()
}
func (a *App) GetSources() []library.SourceInfo { return library.New(a.dir, a.sources).Sources() }

func (a *App) CheckContentProvider(request library.ContentRequest) error {
	session, vault, err := a.credentialSession()
	if err != nil {
		return err
	}
	request.Credentials, err = resolvedCredentials(vault, request.ID, request.Credentials)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return session.CheckContentProvider(ctx, request)
}

func (a *App) StartArtworkSearch(request library.ArtworkSearchRequest) (library.ArtworkSearch, error) {
	session, vault, err := a.credentialSession()
	if err != nil {
		return library.ArtworkSearch{}, err
	}
	for i := range request.Providers {
		values, err := resolvedCredentials(vault, request.Providers[i].ID, request.Providers[i].Credentials)
		if err != nil {
			return library.ArtworkSearch{}, err
		}
		request.Providers[i].Credentials = values
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.library != session {
		return library.ArtworkSearch{}, fmt.Errorf("el perfil cambió; repite la búsqueda de arte")
	}
	return session.StartArtworkSearch(request)
}
func (a *App) GetArtworkSearch(id string) (library.ArtworkSearch, error) {
	s, e := a.session()
	if e != nil {
		return library.ArtworkSearch{}, e
	}
	return s.GetArtworkSearch(id)
}
func (a *App) CancelArtworkSearch(id string) {
	if s, e := a.session(); e == nil {
		s.CancelArtworkSearch(id)
	}
}
func (a *App) DiscardArtworkSearch(id string) {
	if s, e := a.session(); e == nil {
		s.DiscardArtworkSearch(id)
	}
}
func (a *App) CommitArtwork(request library.ArtworkCommit) (library.Snapshot, error) {
	s, e := a.session()
	if e != nil {
		return library.Snapshot{}, e
	}
	return s.CommitArtwork(request)
}
func (a *App) UndoArtwork() (library.Snapshot, error) {
	s, e := a.session()
	if e != nil {
		return library.Snapshot{}, e
	}
	return s.UndoArtwork()
}

// Capture library and vault together so a concurrent player switch never mixes them.
func (a *App) credentialSession() (*library.Service, credentials.Vault, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.library == nil {
		return nil, nil, fmt.Errorf("selecciona un perfil de Gemu")
	}
	return a.library, a.vault, nil
}
func resolvedCredentials(vault credentials.Vault, provider string, draft map[string]string) (map[string]string, error) {
	result := map[string]string{}
	if vault != nil {
		saved, err := vault.Get(provider)
		if err != nil {
			return nil, err
		}
		result = saved
	}
	for field, value := range draft {
		if value != "" {
			result[field] = value
		}
	}
	return result, nil
}
func (a *App) GetCredentialStatus(provider string) (credentials.Status, error) {
	_, vault, err := a.credentialSession()
	if err != nil {
		return credentials.Status{}, err
	}
	if vault == nil {
		return credentials.Status{}, fmt.Errorf("gestor de credenciales no disponible")
	}
	status, err := vault.Status(provider)
	if provider == "screenscraper" {
		personal := []string{}
		for _, field := range status.SavedFields {
			if field == "username" || field == "password" {
				personal = append(personal, field)
			}
		}
		status.SavedFields = personal
	}
	return status, err
}
func (a *App) SaveProviderCredentials(request credentials.Request) error {
	_, vault, err := a.credentialSession()
	if err != nil {
		return err
	}
	if vault == nil {
		return fmt.Errorf("gestor de credenciales no disponible")
	}
	return vault.Save(request)
}
func (a *App) DeleteProviderCredentials(provider string) error {
	_, vault, err := a.credentialSession()
	if err != nil {
		return err
	}
	if vault == nil {
		return fmt.Errorf("gestor de credenciales no disponible")
	}
	return vault.Delete(provider)
}
