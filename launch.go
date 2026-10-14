package main

import (
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"kyberix.com/gemu/internal/launcher"
	"log"
	goruntime "runtime"
	"time"
)

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.launch = launcher.New(launcher.Native())
	go func() {
		// RegisterHotKey and its message queue belong to an OS thread.
		goruntime.LockOSThread()
		defer goruntime.UnlockOSThread()
		defer a.launch.ReleaseShortcuts()
		previous := launcher.Session{}
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				session, restore := a.launch.Poll()
				if session != previous {
					if !session.Overlay && !restore && session.Menu != previous.Menu {
						a.setOverlay(session.Menu)
					}
					runtime.EventsEmit(ctx, "gemu:session", session)
					previous = session
				}
				if restore {
					a.restoreWindow()
				}
			}
		}
	}()
}
func (a *App) restoreWindow() {
	if err := presentGameWindow(desktopGameWindow{ctx: a.ctx}, a.launch.Current().Menu); err != nil {
		log.Printf("No se pudo traer Gemu al frente: %v", err)
	}
}

func (a *App) setOverlay(enabled bool) {
	if err := (desktopGameWindow{ctx: a.ctx}).Overlay(enabled); err != nil {
		log.Printf("No se pudo cambiar el modo overlay de Gemu: %v", err)
	}
}
func (a *App) LaunchGame(id string) (launcher.Session, error) {
	if a.launch == nil {
		return launcher.Session{}, fmt.Errorf("El host no está listo")
	}
	library, err := a.session()
	if err != nil {
		return launcher.Session{}, err
	}
	snapshot, err := library.Load()
	if err != nil {
		return launcher.Session{}, err
	}
	for _, entry := range snapshot.Games {
		if entry.Game.ID == id {
			for _, installation := range entry.Game.Installations {
				if installation.Provider == "steam" && installation.Installed {
					result, err := a.launch.Launch(id, entry.Game.Title, installation)
					if err == nil {
						runtime.WindowMinimise(a.ctx)
					}
					return result, err
				}
			}
		}
	}
	return launcher.Session{}, fmt.Errorf("No se encontró una instalación compatible del juego")
}
func (a *App) GetGameSession() launcher.Session {
	if a.launch == nil {
		return launcher.Session{State: "idle"}
	}
	return a.launch.Current()
}
func (a *App) OpenGameMenu() launcher.Session {
	result := a.launch.Menu()
	if result.Menu && !result.Overlay {
		a.restoreWindow()
	}
	return result
}
func (a *App) ReturnToGemu() launcher.Session {
	result := a.launch.Home()
	a.restoreWindow()
	return result
}
func (a *App) ResumeGame() error {
	a.setOverlay(false)
	if err := a.launch.Resume(); err != nil {
		a.setOverlay(a.launch.Current().Menu)
		return err
	}
	return nil
}
func (a *App) CloseGame() error { return a.launch.Stop() }

func (a *App) SetGameOverlayLanguage(language string) {
	if a.launch != nil && (language == "es" || language == "en") {
		a.launch.SetOverlayLanguage(language)
	}
}
