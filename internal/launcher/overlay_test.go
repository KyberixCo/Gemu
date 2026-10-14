package launcher

import (
	"path/filepath"
	"testing"
)

type fakeOverlayPlatform struct {
	fakePlatform
	ready, visible           bool
	command                  uint32
	focusCalls, releaseCalls int
}

func (p *fakeOverlayPlatform) SyncOverlay(_ []Process, s Session) bool {
	p.visible = p.ready && s.Menu && s.Overlay
	return p.ready
}
func (p *fakeOverlayPlatform) OverlayCommand() uint32 {
	command := p.command
	p.command = 0
	return command
}
func (p *fakeOverlayPlatform) OverlayLanguage(string) {}
func (p *fakeOverlayPlatform) ReleaseOverlay()        { p.releaseCalls++; p.visible = false }
func (p *fakeOverlayPlatform) Focus(Process) error    { p.focusCalls++; return nil }

func overlaySession(t *testing.T) (*Service, *fakeOverlayPlatform) {
	p := &fakeOverlayPlatform{}
	s := New(p)
	i := installation(t)
	if _, err := s.Launch("id", "Game", i); err != nil {
		t.Fatal(err)
	}
	p.processes = []Process{{PID: 12, Created: 42, Path: filepath.Join(i.Path, "game.exe")}}
	p.ready = true
	pollSession(s)
	return s, p
}
func TestInjectedMenuDoesNotRestoreOrFocusHost(t *testing.T) {
	s, p := overlaySession(t)
	p.shortcut = true
	session, restore := s.Poll()
	if !session.Overlay || !session.Menu || restore || !p.visible {
		t.Fatalf("injected menu opened a window: %+v restore=%v", session, restore)
	}
	p.shortcut = true
	session, restore = s.Poll()
	if session.Menu || restore || p.visible || p.focusCalls != 0 {
		t.Fatal("closing overlay changed foreground focus")
	}
}
func TestInjectedActions(t *testing.T) {
	for _, command := range []uint32{1, 2, 3} {
		t.Run(string(rune('0'+command)), func(t *testing.T) {
			s, p := overlaySession(t)
			s.Menu()
			p.command = command
			session, restore := s.Poll()
			if session.Menu || p.visible || p.focusCalls != 0 {
				t.Fatal("native command left menu open or changed focus")
			}
			if restore != (command == 2) || session.Background != (command == 2) {
				t.Fatal("only returning to the library may restore Gemu")
			}
			if command == 3 && (session.State != "closing" || len(p.closed) != 1) {
				t.Fatal("close command did not use confirmed game processes")
			}
		})
	}
}
func TestOverlayLossRestoresWindowFallback(t *testing.T) {
	s, p := overlaySession(t)
	s.Menu()
	p.ready = false
	session, restore := s.Poll()
	if session.Overlay || !session.Menu || !restore {
		t.Fatal("lost renderer left an invisible menu")
	}
}
func TestShutdownReleasesInjectedOverlay(t *testing.T) {
	s, p := overlaySession(t)
	s.Menu()
	s.ReleaseShortcuts()
	if p.releaseCalls == 0 || p.visible {
		t.Fatal("overlay was not released")
	}
}
