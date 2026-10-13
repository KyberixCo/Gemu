package steam

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// DetectRoot only reads registry values and conventional installation paths.
func DetectRoot() (string, error) {
	var candidates []string
	for _, entry := range []struct {
		hive       registry.Key
		key, value string
	}{
		{registry.CURRENT_USER, `Software\Valve\Steam`, "SteamPath"},
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Valve\Steam`, "InstallPath"},
		{registry.LOCAL_MACHINE, `SOFTWARE\Valve\Steam`, "InstallPath"},
	} {
		key, err := registry.OpenKey(entry.hive, entry.key, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		value, _, err := key.GetStringValue(entry.value)
		key.Close()
		if err == nil {
			candidates = append(candidates, value)
		}
	}
	for _, variable := range []string{"ProgramFiles(x86)", "ProgramFiles"} {
		if base := os.Getenv(variable); base != "" {
			candidates = append(candidates, filepath.Join(base, "Steam"))
		}
	}
	return findRoot(candidates)
}
