package steam

import (
	"fmt"
	"kyberix.com/gemu/internal/domain"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// Discover only reads local Steam manifests. It never authenticates or runs a client.
// Malformed manifests produce diagnostics while healthy libraries remain available.
func Discover(root string) ([]domain.Game, []error) {
	roots := []string{root}
	var diagnostics []error
	data, err := os.ReadFile(filepath.Join(root, "steamapps", "libraryfolders.vdf"))
	if err == nil {
		tree, err := parse(string(data))
		if err != nil {
			diagnostics = append(diagnostics, err)
		} else {
			for key, n := range tree["libraryfolders"].children {
				if _, err := strconv.Atoi(key); err != nil {
					continue
				}
				path := n.value
				if n.children != nil {
					path = n.children["path"].value
				}
				if path != "" {
					roots = append(roots, path)
				}
			}
		}
	} else if !os.IsNotExist(err) {
		diagnostics = append(diagnostics, err)
	}
	seenRoots := map[string]bool{}
	byID := map[string]domain.Game{}
	for _, library := range roots {
		if !filepath.IsAbs(library) {
			diagnostics = append(diagnostics, fmt.Errorf("invalid relative Steam library path: %s", library))
			continue
		}
		library = filepath.Clean(library)
		key := library
		if runtime.GOOS == "windows" {
			key = strings.ToLower(key)
		}
		if seenRoots[key] {
			continue
		}
		seenRoots[key] = true
		entries, err := os.ReadDir(filepath.Join(library, "steamapps"))
		if err != nil {
			diagnostics = append(diagnostics, err)
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasPrefix(entry.Name(), "appmanifest_") || !strings.HasSuffix(entry.Name(), ".acf") {
				continue
			}
			path := filepath.Join(library, "steamapps", entry.Name())
			game, err := readManifest(path, library)
			if err != nil {
				diagnostics = append(diagnostics, fmt.Errorf("%s: %w", path, err))
				continue
			}
			// Steam's shared redistributables are a support package, not a playable title.
			if game.ID == "steam:228980" {
				continue
			}
			if existing, ok := byID[game.ID]; ok {
				existing.Installations = append(existing.Installations, game.Installations...)
				byID[game.ID] = existing
			} else {
				byID[game.ID] = game
			}
		}
	}
	games := make([]domain.Game, 0, len(byID))
	for _, game := range byID {
		games = append(games, game)
	}
	sort.Slice(games, func(i, j int) bool { return games[i].Title < games[j].Title })
	return games, diagnostics
}
func readManifest(path, root string) (domain.Game, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.Game{}, err
	}
	tree, err := parse(string(data))
	if err != nil {
		return domain.Game{}, err
	}
	state := tree["appstate"].children
	id, title, dir := state["appid"].value, state["name"].value, state["installdir"].value
	if !ValidAppID(id) || strings.TrimSpace(title) == "" || dir == "" || dir == "." || dir == ".." || strings.ContainsAny(dir, "/\\:") {
		return domain.Game{}, fmt.Errorf("invalid Steam manifest identity or install directory")
	}
	flags, err := strconv.ParseUint(state["stateflags"].value, 10, 64)
	if err != nil {
		return domain.Game{}, fmt.Errorf("invalid installation flags")
	}
	installPath := filepath.Join(root, "steamapps", "common", dir)
	info, statErr := os.Stat(installPath)
	if statErr != nil && !os.IsNotExist(statErr) {
		return domain.Game{}, fmt.Errorf("cannot inspect installation: %w", statErr)
	}
	installed := flags&4 != 0 && statErr == nil && info.IsDir()
	return domain.Game{ID: "steam:" + id, Title: title, Installations: []domain.Installation{{Provider: "steam", ExternalID: id, Path: installPath, Installed: installed}}}, nil
}
func ValidAppID(id string) bool {
	if id == "" {
		return false
	}
	for _, c := range id {
		if c < '0' || c > '9' {
			return false
		}
	}
	n, err := strconv.ParseUint(id, 10, 32)
	return err == nil && n > 0
}
