package launcher

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type executablePlatform interface {
	StartExecutable(string, []string, string) (Process, error)
}

// The caller builds an argument list from a validated connector configuration.
// A direct process is tracked by PID + creation time + path, not by its folder.
func (s *Service) LaunchExecutable(id, title, executable string, args []string, dir string) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.session.State != "idle" && s.session.State != "ended" && s.session.State != "error" {
		return s.session, fmt.Errorf("Ya hay una sesión activa")
	}
	platform, ok := s.platform.(executablePlatform)
	if !ok {
		return s.session, fmt.Errorf("direct executable launch is unavailable")
	}
	if !filepath.IsAbs(executable) || !filepath.IsAbs(dir) {
		return s.session, fmt.Errorf("executable launch requires absolute paths")
	}
	processes, err := s.platform.Processes()
	if err != nil {
		return s.session, err
	}
	for _, p := range processes {
		if strings.EqualFold(filepath.Clean(p.Path), filepath.Clean(executable)) {
			return s.session, fmt.Errorf("RetroArch is already running; close it or resume its current session")
		}
	}
	p, err := platform.StartExecutable(executable, args, dir)
	if err != nil {
		return s.session, err
	}
	s.direct = &p
	s.tracked = []Process{p}
	s.baseline = nil
	s.root = ""
	s.started = time.Now()
	s.absent = 0
	s.lastScan = time.Time{}
	s.session = Session{GameID: id, Title: title, State: "running"}
	return s.session, nil
}
