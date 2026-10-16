package launcher

import (
	"path/filepath"
	"testing"
)

type directPlatform struct {
	fakePlatform
	args       []string
	dir        string
	executable string
	started    Process
}

func (p *directPlatform) StartExecutable(path string, args []string, dir string) (Process, error) {
	p.args = append([]string{}, args...)
	p.dir = dir
	p.executable = path
	p.started = Process{PID: 42, Created: 100, Path: path}
	p.processes = append(p.processes, p.started)
	return p.started, nil
}
func TestDirectLaunchTracksOnlyItsExactProcess(t *testing.T) {
	root := t.TempDir()
	exe := filepath.Join(root, "Retro Arch", "retroarch.exe")
	dir := filepath.Dir(exe)
	p := &directPlatform{}
	s := New(p)
	args := []string{"--config", filepath.Join(root, "player config.cfg"), "-L", filepath.Join(root, "core.dll"), filepath.Join(root, "A game.nes")}
	result, err := s.LaunchExecutable("retroarch:one", "One", exe, args, dir)
	if err != nil || result.State != "running" || p.args[4] != args[4] || p.dir != dir {
		t.Fatal(result, err)
	}
	p.processes = append(p.processes, Process{PID: 43, Created: 200, Path: exe}, Process{PID: 44, Created: 201, Path: filepath.Join(dir, "helper.exe")})
	result, _ = pollSession(s)
	if result.State != "running" || len(s.tracked) != 1 || s.tracked[0] != p.started {
		t.Fatal("tracked unrelated process", s.tracked)
	}
	if err = s.Stop(); err != nil || len(p.closed) != 1 || p.closed[0] != p.started {
		t.Fatal(p.closed, err)
	}
	p.processes = []Process{{PID: 42, Created: 999, Path: exe}}
	_, restore := pollSession(s)
	if restore {
		t.Fatal("did not debounce process exit")
	}
	result, restore = pollSession(s)
	if !restore || result.State != "ended" {
		t.Fatal("reused PID kept session alive", result)
	}
}
func TestDirectLaunchRejectsAnotherOpenInstanceAndUnavailablePlatform(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "retroarch.exe")
	p := &directPlatform{fakePlatform: fakePlatform{processes: []Process{{PID: 4, Path: exe}}}}
	if _, err := New(p).LaunchExecutable("id", "title", exe, nil, filepath.Dir(exe)); err == nil || p.executable != "" {
		t.Fatal("opened second RetroArch")
	}
	if _, err := New(&fakePlatform{}).LaunchExecutable("id", "title", exe, nil, filepath.Dir(exe)); err == nil {
		t.Fatal("assumed direct launch capability")
	}
}
