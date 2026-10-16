package retroarch

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
)

func saveIdentity(id string) string          { return fmt.Sprintf("%x", sha256.Sum256([]byte(id))) }
func executableDirectory(path string) string { return filepath.Dir(path) }
func (s *Service) MenuLaunch(id string, config Config) (Launch, error) {
	config, err := s.normalize(config, id)
	if err != nil {
		return Launch{}, err
	}
	core := filepath.Join(s.managedDir(), "cores", "menu_libretro.dll")
	for _, collection := range config.Collections {
		if sys, ok := system(collection.SystemID); ok {
			if installed, e := s.corePath(config, sys.Core); e == nil {
				core = installed
				break
			}
		}
	}
	file, err := s.launchConfig(id, config, "setup", core)
	if err != nil {
		return Launch{}, err
	}
	base, err := s.baseConfig(id)
	if err != nil {
		return Launch{}, err
	}
	return Launch{Executable: config.Executable, Directory: filepath.Dir(config.Executable), Args: []string{"--config", base, "--appendconfig", file, "--menu"}}, nil
}
