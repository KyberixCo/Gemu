package launcher

import (
	"fmt"
	"kyberix.com/gemu/internal/domain"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Process struct {
	PID     uint32
	Created uint64
	Path    string
}
type Platform interface {
	Open(string) error
	Processes() ([]Process, error)
	Focus(Process) error
	Close(Process) error
	// Shortcut must be polled and released on the same OS thread on Windows.
	Shortcut(active bool) bool
}
type Session struct {
	GameID     string `json:"gameId"`
	Title      string `json:"title"`
	State      string `json:"state"`
	Menu       bool   `json:"menu"`
	Overlay    bool   `json:"overlay"`
	Background bool   `json:"background"`
	Message    string `json:"message"`
}
type Service struct {
	mu              sync.Mutex
	platform        Platform
	session         Session
	root            string
	baseline        map[Process]bool
	tracked         []Process
	started         time.Time
	absent          int
	lastScan        time.Time
	direct          *Process
	assistRequested bool
}

type overlayPlatform interface {
	SyncOverlay([]Process, Session) bool
	OverlayCommand() uint32
	OverlayLanguage(string)
	ReleaseOverlay()
}

func (s *Service) SetOverlayLanguage(language string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.platform.(overlayPlatform); ok {
		p.OverlayLanguage(language)
	}
}

func (s *Service) syncOverlay() {
	if s.direct != nil {
		return
	}
	if p, ok := s.platform.(overlayPlatform); ok {
		ready := p.SyncOverlay(s.tracked, s.session)
		if !s.session.Menu || s.session.Overlay {
			s.session.Overlay = ready
		}
	}
}

func New(platform Platform) *Service {
	return &Service{platform: platform, session: Session{State: "idle"}}
}
func Within(root, path string) bool {
	if root == "" || !filepath.IsAbs(root) {
		return false
	}
	rel, err := filepath.Rel(strings.ToLower(filepath.Clean(root)), strings.ToLower(filepath.Clean(path)))
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
func (s *Service) Launch(id, title string, installation domain.Installation) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.session.State != "idle" && s.session.State != "ended" && s.session.State != "error" {
		return s.session, fmt.Errorf("Ya hay una sesión activa")
	}
	spec, err := PrepareSteam(installation)
	if err != nil {
		return s.session, err
	}
	if !filepath.IsAbs(installation.Path) {
		return s.session, fmt.Errorf("La instalación no tiene una ruta absoluta válida")
	}
	processes, err := s.platform.Processes()
	if err != nil {
		return s.session, err
	}
	s.baseline = map[Process]bool{}
	for _, p := range processes {
		if Within(installation.Path, p.Path) {
			return s.session, fmt.Errorf("El juego ya tiene un proceso abierto; ciérralo antes de lanzarlo")
		}
		s.baseline[p] = true
	}
	if err = s.platform.Open(spec.URI); err != nil {
		return s.session, err
	}
	s.root = installation.Path
	s.direct = nil
	s.started = time.Now()
	s.tracked = nil
	s.absent = 0
	s.lastScan = time.Time{}
	s.session = Session{GameID: id, Title: title, State: "requested", Message: "Esperando el proceso del juego…"}
	return s.session, nil
}
func (s *Service) Poll() (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	restore := false
	active := s.session.State == "requested" || s.session.State == "running" || s.session.State == "closing"
	shortcut := s.platform.Shortcut(active)
	if s.session.State == "requested" || s.session.State == "running" || s.session.State == "closing" {
		interval := 500 * time.Millisecond
		if s.session.State == "running" {
			interval = time.Second
		}
		if time.Since(s.lastScan) >= interval {
			s.lastScan = time.Now()
			processes, err := s.platform.Processes()
			if err == nil {
				found := []Process{}
				for _, p := range processes {
					if s.direct != nil && p == *s.direct || s.direct == nil && Within(s.root, p.Path) && !s.baseline[p] {
						found = append(found, p)
					}
				}
				if len(found) > 0 {
					s.tracked = found
					s.absent = 0
					if s.session.State != "closing" {
						s.session.State = "running"
						s.session.Message = ""
					}
				} else if s.session.State != "requested" {
					s.absent++
					if s.absent >= 10 || s.direct != nil && s.absent >= 2 {
						s.session.State = "ended"
						s.session.Menu = false
						s.session.Message = ""
						s.tracked = nil
						restore = true
					}
				} else if time.Since(s.started) > 2*time.Minute {
					s.session.State = "error"
					s.session.Menu = false
					s.session.Message = "No se pudo confirmar el inicio del juego."
					restore = true
				}
			}
		}
		if s.session.State == "running" || s.session.State == "closing" {
			hadOverlay := s.session.Overlay
			s.syncOverlay()
			if hadOverlay && !s.session.Overlay && s.session.Menu {
				restore = true
			}
		}
		if p, ok := s.platform.(overlayPlatform); ok && s.session.Overlay {
			command := p.OverlayCommand()
			if command != 0 {
				shortcut = false
			}
			switch command {
			case 1:
				s.session.Menu = false
				s.session.Background = false
			case 2:
				s.session.Menu = false
				s.session.Background = true
				restore = true
			case 3:
				if err := s.stop(); err != nil {
					s.session.Message = err.Error()
					s.session.Menu = true
				} else {
					s.session.Menu = false
				}
			case 4:
				s.session.Menu = false
				s.assistRequested = true
			}
		}
		if shortcut && (s.session.State == "requested" || s.session.State == "running" || s.session.State == "closing") {
			s.session.Menu = !s.session.Menu
			s.session.Background = false
			restore = s.session.Menu && !s.session.Overlay
			if !s.session.Menu {
				if !s.session.Overlay && s.focus() != nil {
					s.session.Menu = true
					restore = true
				}
			}
		}
	}
	if s.session.State == "ended" || s.session.State == "error" {
		if p, ok := s.platform.(overlayPlatform); ok {
			p.ReleaseOverlay()
		}
		s.session.Overlay = false
	} else if s.session.Overlay {
		s.syncOverlay()
	}
	return s.session, restore
}
func (s *Service) Current() Session { s.mu.Lock(); defer s.mu.Unlock(); return s.session }

// TakeAssistanceRequest transfers a native-overlay request to the desktop host.
func (s *Service) TakeAssistanceRequest() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	requested := s.assistRequested
	s.assistRequested = false
	return requested
}

// ReleaseShortcuts is called by the polling thread before it exits.
func (s *Service) ReleaseShortcuts() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.platform.Shortcut(false)
	if p, ok := s.platform.(overlayPlatform); ok {
		p.ReleaseOverlay()
	}
}
func (s *Service) Menu() Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.session.State == "requested" || s.session.State == "running" || s.session.State == "closing" {
		s.session.Menu = true
		s.session.Background = false
		if s.session.Overlay {
			s.syncOverlay()
		}
	}
	return s.session
}
func (s *Service) Home() Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.session.Menu = false
	s.session.Background = true
	if s.session.Overlay {
		s.syncOverlay()
	}
	return s.session
}
func (s *Service) Resume() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.tracked) == 0 {
		return fmt.Errorf("Todavía no se ha detectado la ventana del juego")
	}
	if !s.session.Overlay || !s.session.Menu || s.session.Background {
		if err := s.focus(); err != nil {
			return err
		}
	}
	s.session.Menu = false
	s.session.Background = false
	if s.session.Overlay {
		s.syncOverlay()
	}
	return nil
}
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stop()
}

func (s *Service) stop() error {
	if len(s.tracked) == 0 {
		return fmt.Errorf("No hay un proceso de juego confirmado para cerrar")
	}
	closed := false
	var last error
	for _, p := range s.tracked {
		if err := s.platform.Close(p); err != nil {
			last = err
		} else {
			closed = true
		}
	}
	if !closed {
		return last
	}
	s.session.State = "closing"
	s.session.Message = "Cierre solicitado. Si el juego muestra una confirmación, vuelve a él para aceptarla."
	return nil
}
func (s *Service) focus() error {
	var last error = fmt.Errorf("Todavía no se ha detectado la ventana del juego")
	for _, p := range s.tracked {
		if err := s.platform.Focus(p); err == nil {
			return nil
		} else {
			last = err
		}
	}
	return last
}
