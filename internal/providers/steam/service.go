package steam

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kyberix.com/gemu/internal/domain"
)

// Report is a read-only snapshot, not a persisted catalog or a running session.
type Report struct {
	Root           string        `json:"root"`
	Games          []domain.Game `json:"games"`
	InstalledCount int           `json:"installedCount"`
	Diagnostics    []string      `json:"diagnostics"`
}

// Inspect accepts a Steam installation/library root. An empty path uses Windows detection.
func Inspect(root string) (Report, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		var err error
		root, err = DetectRoot()
		if err != nil {
			return Report{}, err
		}
	}
	if !filepath.IsAbs(root) {
		return Report{}, fmt.Errorf("la carpeta de Steam debe ser una ruta absoluta")
	}
	root = filepath.Clean(root)
	if info, err := os.Stat(filepath.Join(root, "steamapps")); err != nil || !info.IsDir() {
		return Report{}, fmt.Errorf("la carpeta no contiene una biblioteca Steam accesible (steamapps): %s", root)
	}
	games, diagnostics := Discover(root)
	report := Report{Root: root, Games: games, Diagnostics: []string{}}
	for _, game := range games {
		for _, installation := range game.Installations {
			if installation.Installed {
				report.InstalledCount++
				break
			}
		}
	}
	for _, diagnostic := range diagnostics {
		report.Diagnostics = append(report.Diagnostics, diagnostic.Error())
	}
	return report, nil
}

func findRoot(candidates []string) (string, error) {
	for _, candidate := range candidates {
		if !filepath.IsAbs(candidate) {
			continue
		}
		if info, err := os.Stat(filepath.Join(candidate, "steamapps")); err == nil && info.IsDir() {
			return filepath.Clean(candidate), nil
		}
	}
	return "", fmt.Errorf("no se encontró Steam; indica la carpeta de instalación o biblioteca")
}
