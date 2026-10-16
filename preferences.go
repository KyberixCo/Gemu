package main

import "kyberix.com/gemu/internal/storage"

func (a *App) preferenceScope(id string) (string, error) {
	if id == "" {
		if _, err := a.profiles.List(); err != nil {
			return "", err
		}
		return storage.Global, nil
	}
	if _, err := a.profiles.Directory(id); err != nil {
		return "", err
	}
	return id, nil
}

// An explicit, validated owner prevents a delayed save from following a player
// switch. Global preferences contain only language and controller identities.
func (a *App) LoadPreferences(profileID string, legacy map[string]string) (storage.Preferences, error) {
	scope, err := a.preferenceScope(profileID)
	if err != nil {
		return storage.Preferences{}, err
	}
	return storage.LoadPreferences(a.dir, scope, legacy)
}

func (a *App) SavePreference(profileID, key, value string) error {
	scope, err := a.preferenceScope(profileID)
	if err != nil {
		return err
	}
	return storage.SavePreference(a.dir, scope, key, value)
}
