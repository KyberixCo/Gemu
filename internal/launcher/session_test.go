package launcher

import (
	"fmt"
	"kyberix.com/gemu/internal/domain"
	"path/filepath"
	"testing"
	"time"
)

type fakePlatform struct {
	processes      []Process
	opened         string
	closed         []Process
	shortcut       bool
	shortcutActive bool
	fail           bool
}

func (p *fakePlatform) Open(uri string) error { p.opened = uri; return nil }
func (p *fakePlatform) Processes() ([]Process, error) {
	if p.fail {
		return nil, fmt.Errorf("scan failed")
	}
	return p.processes, nil
}
func (p *fakePlatform) Focus(Process) error         { return nil }
func (p *fakePlatform) Close(process Process) error { p.closed = append(p.closed, process); return nil }
func (p *fakePlatform) Shortcut(active bool) bool {
	p.shortcutActive = active
	value := p.shortcut
	p.shortcut = false
	return active && value
}
func installation(t *testing.T) domain.Installation {
	return domain.Installation{Provider: "steam", ExternalID: "10", Installed: true, Path: t.TempDir()}
}
func TestSessionLifecycle(t *testing.T) {
	p := &fakePlatform{}
	s := New(p)
	i := installation(t)
	if result, err := s.Launch("steam:10", "Game", i); err != nil || result.State != "requested" || p.opened != "steam://run/10" {
		t.Fatalf("launch: %+v %v", result, err)
	}
	if err := s.Stop(); err == nil {
		t.Fatal("must not close unconfirmed processes")
	}
	game := Process{PID: 12, Created: 123, Path: filepath.Join(i.Path, "game.exe")}
	p.processes = []Process{game, {PID: 15, Path: filepath.Join(i.Path+"-other", "other.exe")}}
	result, _ := pollSession(s)
	if result.State != "running" {
		t.Fatal(result)
	}
	if _, err := s.Launch("other", "Other", i); err == nil {
		t.Fatal("duplicate launch accepted")
	}
	p.shortcut = true
	result, restore := pollSession(s)
	if !restore || !result.Menu {
		t.Fatal("menu did not restore")
	}
	if result = s.Home(); !result.Background || result.Menu {
		t.Fatal(result)
	}
	if err := s.Resume(); err != nil || s.Current().Background {
		t.Fatal("resume failed")
	}
	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}
	if len(p.closed) != 1 || p.closed[0] != game {
		t.Fatal("unrelated process included", p.closed)
	}
	p.processes = nil
	for n := 0; n < 9; n++ {
		result, restore = pollSession(s)
		if restore {
			t.Fatal("ended too soon")
		}
	}
	result, restore = pollSession(s)
	if result.State != "ended" || !restore {
		t.Fatal(result)
	}
}
func TestSessionUnverifiedLaunch(t *testing.T) {
	p := &fakePlatform{}
	s := New(p)
	i := installation(t)
	_, err := s.Launch("id", "Game", i)
	if err != nil {
		t.Fatal(err)
	}
	s.started = time.Now().Add(-3 * time.Minute)
	p.fail = true
	result, restore := pollSession(s)
	if result.State != "requested" || restore {
		t.Fatal("scan errors must not imply exit")
	}
	p.fail = false
	result, restore = pollSession(s)
	if result.State != "error" || !restore {
		t.Fatal(result)
	}
}
func TestSessionRejectsExistingGame(t *testing.T) {
	i := installation(t)
	p := &fakePlatform{processes: []Process{{PID: 1, Path: filepath.Join(i.Path, "game.exe")}}}
	_, err := New(p).Launch("id", "Game", i)
	if err == nil || p.opened != "" {
		t.Fatal("existing game accepted")
	}
}
func TestWithin(t *testing.T) {
	root := t.TempDir()
	if !Within(root, filepath.Join(root, "bin", "game.exe")) || Within(root, filepath.Join(root+"other", "game.exe")) || Within(root, root) || Within("", root) {
		t.Fatal("unsafe containment")
	}
}

func TestShortcutOnlyActiveDuringSession(t *testing.T) {
	p := &fakePlatform{shortcut: true}
	s := New(p)
	if session, restore := s.Poll(); p.shortcutActive || session.Menu || restore {
		t.Fatal("idle session must not capture shortcuts")
	}
	if _, err := s.Launch("id", "Game", installation(t)); err != nil {
		t.Fatal(err)
	}
	p.shortcut = true
	if session, restore := s.Poll(); !p.shortcutActive || !session.Menu || !restore {
		t.Fatal("launching session must capture shortcuts")
	}
	s.ReleaseShortcuts()
	if p.shortcutActive {
		t.Fatal("shutdown must release shortcuts")
	}
	s.session.State = "ended"
	p.shortcut = true
	if _, restore := s.Poll(); p.shortcutActive || restore {
		t.Fatal("ended session must not capture shortcuts")
	}
}

func pollSession(s *Service) (Session, bool) { s.lastScan = time.Time{}; return s.Poll() }
