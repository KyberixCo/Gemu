package main

import (
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"kyberix.com/gemu/internal/launcher"
	"kyberix.com/gemu/internal/profiles"
	"kyberix.com/gemu/internal/retroarch"
)

type retroTestPlatform struct {
	args      []string
	processes []launcher.Process
}

func (*retroTestPlatform) Open(string) error                        { return nil }
func (p *retroTestPlatform) Processes() ([]launcher.Process, error) { return p.processes, nil }
func (*retroTestPlatform) Focus(launcher.Process) error             { return nil }
func (*retroTestPlatform) Close(launcher.Process) error             { return nil }
func (*retroTestPlatform) Shortcut(bool) bool                       { return false }
func (p *retroTestPlatform) StartExecutable(path string, args []string, _ string) (launcher.Process, error) {
	p.args = args
	process := launcher.Process{PID: 123, Created: 1, Path: path}
	p.processes = []launcher.Process{process}
	return process, nil
}
func TestRetroArchEndpointsPublishForExplicitOwnerAndLaunchWithoutShell(t *testing.T) {
	if runtime.GOOS != "windows" || runtime.GOARCH != "amd64" {
		t.Skip("Windows PE fixture")
	}
	dir := t.TempDir()
	app := &App{dir: dir, profiles: profiles.New(dir)}
	player, err := app.CreateProfile("Retro player")
	if err != nil {
		t.Fatal(err)
	}
	other, err := app.CreateProfile("Other")
	if err != nil {
		t.Fatal(err)
	}
	if err = app.SelectProfile(player.ID); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	image, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	// These are structural PE fixtures. They are never passed to an OS process.
	fixtureDir := t.TempDir()
	exe := filepath.Join(fixtureDir, "retroarch.exe")
	if err = os.WriteFile(exe, image, 0600); err != nil {
		t.Fatal(err)
	}
	header := int(binary.LittleEndian.Uint32(image[0x3c:]))
	flags := header + 4 + 18
	binary.LittleEndian.PutUint16(image[flags:], binary.LittleEndian.Uint16(image[flags:])|0x2000)
	if err = os.WriteFile(filepath.Join(fixtureDir, "fceumm_libretro.dll"), image, 0600); err != nil {
		t.Fatal(err)
	}
	folders, err := app.CreateRetroArchROMFolders(player.ID, []string{"nes"})
	if err != nil {
		t.Fatal(err)
	}
	rom := filepath.Join(folders["nes"], "My homebrew.nes")
	if err = os.WriteFile(rom, []byte("structural ROM fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	status, err := app.GetRetroArchStatus(player.ID, exe)
	if err != nil || len(status.Installations) == 0 {
		t.Fatal(status, err)
	}
	config := status.Config
	config.Executable = exe
	config.CoreDirectory = fixtureDir
	config.Collections = []retroarch.Collection{{SystemID: "nes", Directory: folders["nes"], Recursive: true}}
	preview, err := app.PreviewRetroArch(player.ID, config)
	if err != nil || !preview.CanPublish || len(preview.Games) != 1 {
		t.Fatal(preview, err)
	}
	if err = app.SelectProfile(other.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = app.CommitRetroArch(other.ID, preview.Token); err == nil {
		t.Fatal("another player committed draft")
	}
	snapshot, err := app.CommitRetroArch(player.ID, preview.Token)
	if err != nil || len(snapshot.Games) != 1 {
		t.Fatal(snapshot, err)
	}
	empty, err := app.GetLibrary()
	if err != nil || len(empty.Games) != 0 {
		t.Fatal("import followed player switch", empty, err)
	}
	if err = app.SelectProfile(player.ID); err != nil {
		t.Fatal(err)
	}
	p := &retroTestPlatform{}
	app.launch = launcher.New(p)
	session, err := app.LaunchGame(snapshot.Games[0].Game.ID)
	if err != nil || session.State != "running" || len(p.args) != 8 || p.args[7] != rom {
		t.Fatal(session, p.args, err)
	}
	if _, err = app.PlanRetroArchInstall(other.ID, retroarch.InstallRequest{Executable: exe, CoreDirectory: fixtureDir, Systems: []string{"nes"}}); err != nil {
		t.Fatal("local core reuse", err)
	}
	if _, err = app.GetRetroArchStatus("../../unknown", exe); err == nil {
		t.Fatal("unknown owner accepted")
	}
	// The original source interface also enumerates this configured library.
	source, err := app.session()
	if err != nil {
		t.Fatal(err)
	}
	games, err := source.Preview(context.Background(), "retroarch", "")
	if err != nil || len(games.Games) != 1 {
		t.Fatal(games, err)
	}
}
