package retroarch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"kyberix.com/gemu/internal/providers/steam"
)

func (s *Service) Detect(id, manual string) (Status, error) {
	config, err := s.Load(id)
	if err != nil {
		return Status{}, err
	}
	status := Status{Supported: supported(), Config: config, Systems: Systems, Installations: []Installation{}, Diagnostics: []string{}}
	status.DefaultDirectories = map[string]string{}
	for _, sys := range Systems {
		status.DefaultDirectories[sys.ID] = filepath.Join(s.profileDir(id), "roms", sys.ID)
	}
	if !status.Supported {
		return status, nil
	}
	candidates := []string{config.Executable}
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		candidates = append(candidates, filepath.Join(local, "Programs", "RetroArch", "retroarch.exe"))
	}
	if manual != "" {
		if st, e := os.Stat(manual); e == nil && st.IsDir() {
			manual = filepath.Join(manual, "retroarch.exe")
		}
		candidates = append([]string{manual}, candidates...)
	}
	driveRoot := ""
	if drive := os.Getenv("SystemDrive"); drive != "" {
		driveRoot = filepath.Join(drive+string(os.PathSeparator), "RetroArch")
	}
	for _, root := range []string{s.managedDir(), os.Getenv("LOCALAPPDATA"), os.Getenv("APPDATA"), os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)"), driveRoot} {
		if root == "" {
			continue
		}
		candidates = append(candidates, filepath.Join(root, "RetroArch", "retroarch.exe"), filepath.Join(root, "retroarch.exe"))
	}
	if path, e := exec.LookPath("retroarch.exe"); e == nil {
		candidates = append(candidates, path)
	}
	candidates = append(candidates, registryCandidates()...)
	// Only bounded, known directories are inspected; never scan whole drives.
	releases, _ := os.ReadDir(filepath.Join(s.managedDir(), "releases"))
	for _, release := range releases {
		if !release.IsDir() || !versionPattern.MatchString(release.Name()) {
			continue
		}
		root := filepath.Join(s.managedDir(), "releases", release.Name())
		candidates = append(candidates, findExecutables(root, 2)...)
	}
	if report, e := steam.Inspect(""); e == nil {
		for _, g := range report.Games {
			for _, installation := range g.Installations {
				if installation.ExternalID == "1118310" && installation.Installed {
					candidates = append(candidates, filepath.Join(installation.Path, "retroarch.exe"))
				}
			}
		}
	}
	seen := map[string]bool{}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		path, e := filepath.EvalSymlinks(candidate)
		if e != nil {
			if candidate == manual {
				status.Diagnostics = append(status.Diagnostics, e.Error())
			}
			continue
		}
		key := strings.ToLower(path)
		if seen[key] {
			continue
		}
		seen[key] = true
		if e = s.validateExecutable(path); e != nil {
			if candidate == manual || candidate == config.Executable {
				status.Diagnostics = append(status.Diagnostics, e.Error())
			}
			continue
		}
		status.Installations = append(status.Installations, Installation{Executable: path, Managed: within(s.managedDir(), path)})
	}
	if manual != "" && len(status.Installations) == 0 {
		status.Diagnostics = append(status.Diagnostics, "RetroArch was not found at the selected location")
	}
	return status, nil
}

func findExecutables(root string, depth int) []string {
	result := []string{}
	entries, err := os.ReadDir(root)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		path := filepath.Join(root, entry.Name())
		if !entry.IsDir() && strings.EqualFold(entry.Name(), "retroarch.exe") {
			result = append(result, path)
		}
		if entry.IsDir() && depth > 0 {
			result = append(result, findExecutables(path, depth-1)...)
		}
	}
	return result
}
func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
func safeProfileID(id string) error {
	if len(id) != 32 || strings.Trim(id, "0123456789abcdef") != "" {
		return fmt.Errorf("invalid RetroArch player")
	}
	return nil
}
