package main

import "kyberix.com/gemu/internal/library"

func (a *App) SearchLibrary(profileID string, request library.SearchRequest) (library.SearchPage, error) {
	// A queued request must never follow a switch to a different player.
	dir, err := a.profiles.Directory(profileID)
	if err != nil {
		return library.SearchPage{}, err
	}
	return library.New(dir, nil).WithPlayerStorage(a.dir, profileID).Search(request)
}
