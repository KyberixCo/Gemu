package main

import (
	"context"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"kyberix.com/gemu/internal/launcher"
	"kyberix.com/gemu/internal/library"
	"kyberix.com/gemu/internal/retroarch"
)

func (a *App) retroSession(profileID string) (*retroarch.Service, string, error) {
	dir, err := a.profiles.Directory(profileID)
	if err != nil {
		return nil, "", err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.retro == nil {
		a.retro = retroarch.New(a.dir)
	}
	return a.retro, dir, nil
}
func (a *App) operationContext() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}
func (a *App) DiscardRetroArchInstallPlan(profileID, token string) {
	if s, _, err := a.retroSession(profileID); err == nil {
		s.DiscardPlan(profileID, token)
	}
}
func (a *App) CreateRetroArchROMFolders(profileID string, systems []string) (map[string]string, error) {
	s, _, err := a.retroSession(profileID)
	if err != nil {
		return nil, err
	}
	return s.CreateFolders(profileID, systems)
}
func (a *App) GetRetroArchStatus(profileID, manual string) (retroarch.Status, error) {
	s, _, err := a.retroSession(profileID)
	if err != nil {
		return retroarch.Status{}, err
	}
	return s.Detect(profileID, manual)
}
func (a *App) PlanRetroArchInstall(profileID string, request retroarch.InstallRequest) (retroarch.Plan, error) {
	s, _, err := a.retroSession(profileID)
	if err != nil {
		return retroarch.Plan{}, err
	}
	ctx, cancel := context.WithTimeout(a.operationContext(), 2*time.Minute)
	defer cancel()
	return s.Plan(ctx, profileID, request)
}
func (a *App) StartRetroArchInstall(profileID, token string) (retroarch.Job, error) {
	s, _, err := a.retroSession(profileID)
	if err != nil {
		return retroarch.Job{}, err
	}
	return s.Start(a.operationContext(), profileID, token)
}
func (a *App) GetRetroArchInstall(profileID, id string) (retroarch.Job, error) {
	s, _, err := a.retroSession(profileID)
	if err != nil {
		return retroarch.Job{}, err
	}
	return s.Job(profileID, id)
}
func (a *App) CancelRetroArchInstall(profileID, id string) error {
	s, _, err := a.retroSession(profileID)
	if err != nil {
		return err
	}
	return s.Cancel(profileID, id)
}
func (a *App) PreviewRetroArch(profileID string, config retroarch.Config) (retroarch.Preview, error) {
	s, _, err := a.retroSession(profileID)
	if err != nil {
		return retroarch.Preview{}, err
	}
	ctx, cancel := context.WithTimeout(a.operationContext(), 2*time.Minute)
	defer cancel()
	return s.Preview(ctx, profileID, config)
}
func (a *App) DiscardRetroArchPreview(profileID, token string) {
	if s, _, err := a.retroSession(profileID); err == nil {
		s.Discard(profileID, token)
	}
}
func (a *App) CommitRetroArch(profileID, token string) (library.Snapshot, error) {
	s, dir, err := a.retroSession(profileID)
	if err != nil {
		return library.Snapshot{}, err
	}
	ctx, cancel := context.WithTimeout(a.operationContext(), 2*time.Minute)
	defer cancel()
	return s.Commit(ctx, profileID, token, library.New(dir, nil).WithPlayerStorage(a.dir, profileID))
}
func (a *App) PickRetroArchPath(profileID, kind string) (string, error) {
	if _, _, err := a.retroSession(profileID); err != nil {
		return "", err
	}
	if a.ctx == nil {
		return "", fmt.Errorf("open Gemu desktop to choose a folder")
	}
	if kind == "executable" {
		return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "RetroArch", Filters: []runtime.FileFilter{{DisplayName: "RetroArch", Pattern: "retroarch.exe"}}})
	}
	if kind != "roms" && kind != "bios" && kind != "cores" {
		return "", fmt.Errorf("invalid RetroArch folder selection")
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "RetroArch"})
}
func (a *App) OpenRetroArchMenu(profileID string, config retroarch.Config) (launcher.Session, error) {
	s, _, err := a.retroSession(profileID)
	if err != nil {
		return launcher.Session{}, err
	}
	plan, err := s.MenuLaunch(profileID, config)
	if err != nil {
		return launcher.Session{}, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.profileID != profileID || a.launch == nil {
		return launcher.Session{}, fmt.Errorf("select the same player before opening RetroArch")
	}
	result, err := a.launch.LaunchExecutable("retroarch:setup", "RetroArch", plan.Executable, plan.Args, plan.Directory)
	if err == nil && a.ctx != nil {
		runtime.WindowMinimise(a.ctx)
	}
	return result, err
}
