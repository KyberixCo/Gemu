package main

import (
	"context"
	"runtime"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"kyberix.com/gemu/internal/launcher"
)

type gameWindow interface {
	Restore() error
	Show()
	Fullscreen()
	Overlay(bool) error
	Focus() error
}

// Restore first: fullscreen must not capture a minimised window's geometry.
// Apply overlay priority last, after all operations that change window styles.
func presentGameWindow(window gameWindow, overlay bool) error {
	if err := window.Restore(); err != nil {
		return err
	}
	window.Show()
	window.Fullscreen()
	if err := window.Overlay(overlay); err != nil {
		return err
	}
	return window.Focus()
}

type desktopGameWindow struct{ ctx context.Context }

func (w desktopGameWindow) Restore() error {
	if err := launcher.RestoreHost(); err != nil {
		return err
	}
	wailsruntime.WindowUnminimise(w.ctx)
	return nil
}
func (w desktopGameWindow) Show()       { wailsruntime.WindowShow(w.ctx) }
func (w desktopGameWindow) Fullscreen() { wailsruntime.WindowFullscreen(w.ctx) }
func (w desktopGameWindow) Overlay(enabled bool) error {
	if runtime.GOOS == "windows" {
		return launcher.SetHostOverlay(enabled)
	}
	wailsruntime.WindowSetAlwaysOnTop(w.ctx, enabled)
	return nil
}
func (w desktopGameWindow) Focus() error { return launcher.FocusHost() }
