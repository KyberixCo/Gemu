package retroarch

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/library"
)

type Source struct {
	service *Service
	owner   string
}

func NewSource(service *Service, owner string) *Source {
	return &Source{service: service, owner: owner}
}
func (*Source) Name() string           { return "RetroArch" }
func (*Source) ArtworkHosts() []string { return nil }
func (p *Source) Discover(ctx context.Context, _ string) (library.Preview, error) {
	if err := safeProfileID(p.owner); err != nil {
		return library.Preview{}, err
	}
	config, err := p.service.Load(p.owner)
	if err != nil {
		return library.Preview{}, err
	}
	config, err = p.service.normalize(config, p.owner)
	if err != nil {
		return library.Preview{}, err
	}
	preview, err := p.service.scan(ctx, config)
	if err != nil {
		return library.Preview{}, err
	}
	if !preview.CanPublish {
		return library.Preview{}, fmt.Errorf("resolve missing cores or invalid BIOS before publishing")
	}
	games := []domain.Game{}
	for _, entry := range preview.Games {
		games = append(games, entry.Game)
	}
	return library.Preview{Root: config.Executable, Games: games, Profiles: []library.Profile{}, Diagnostics: preview.Diagnostics}, nil
}
func (*Source) Metadata(_ context.Context, game domain.Game, _ string) (library.Metadata, error) {
	platform := ""
	if len(game.Installations) > 0 {
		if sys, ok := system(game.Installations[0].ExternalID); ok {
			platform = sys.Name
		}
	}
	return library.Metadata{Platform: platform, Genres: []string{}, Origin: "local", Status: "pending"}, nil
}
func (*Source) Player(context.Context, domain.Game, library.Profile, string) library.PlayerState {
	return library.PlayerState{Profile: library.Profile{}, Status: "pending", Achievements: []library.Achievement{}}
}

type Launch struct {
	Executable string
	Args       []string
	Directory  string
}

func (s *Service) BuildLaunch(id, gameID string, installation domain.Installation) (Launch, error) {
	if installation.Provider != "retroarch" || !installation.Installed {
		return Launch{}, fmt.Errorf("invalid RetroArch installation")
	}
	config, err := s.Load(id)
	if err != nil {
		return Launch{}, err
	}
	config, err = s.normalize(config, id)
	if err != nil {
		return Launch{}, err
	}
	sys, ok := system(installation.ExternalID)
	if !ok {
		return Launch{}, fmt.Errorf("unknown RetroArch system")
	}
	if _, diagnostics := s.inspectFirmware(config, sys); len(diagnostics) > 0 {
		return Launch{}, fmt.Errorf("invalid RetroArch BIOS checksum")
	}
	var collection *Collection
	for i := range config.Collections {
		if config.Collections[i].SystemID == sys.ID {
			collection = &config.Collections[i]
			break
		}
	}
	if collection == nil {
		return Launch{}, fmt.Errorf("RetroArch collection is unavailable")
	}
	path, err := contentPath(collection.Directory, installation.Path)
	if err != nil {
		return Launch{}, err
	}
	if !collection.Recursive && !strings.EqualFold(filepath.Dir(path), collection.Directory) {
		return Launch{}, fmt.Errorf("game is outside the configured non-recursive collection")
	}
	if !hasExtension(path, sys.Extensions) && !(sys.ID != "psx" && hasExtension(path, []string{".zip"})) {
		return Launch{}, fmt.Errorf("unsupported RetroArch game format")
	}
	if _, err = validateContent(collection.Directory, path, sys); err != nil {
		return Launch{}, err
	}
	core, err := s.corePath(config, sys.Core)
	if err != nil {
		return Launch{}, err
	}
	// Avoid OS-invalid game IDs and collisions between same-named ROMs.
	saveID := saveIdentity(gameID)
	file, err := s.launchConfig(id, config, saveID, core)
	if err != nil {
		return Launch{}, err
	}
	base, err := s.baseConfig(id)
	if err != nil {
		return Launch{}, err
	}
	return Launch{Executable: config.Executable, Directory: executableDirectory(config.Executable), Args: []string{"--config", base, "--appendconfig", file, "--verbose", "-L", core, path}}, nil
}
