package main

import (
	"errors"
	"testing"
)

// Reproduce the host behaviour: Show is queued, fullscreen does not restore a
// minimised window, and changing fullscreen styles can reset overlay priority.
type simulatedGameWindow struct {
	minimised, visible, fullscreen, overlay, focused bool
	restoreErr, overlayErr, focusErr                 error
}

func (w *simulatedGameWindow) Restore() error {
	if w.restoreErr != nil {
		return w.restoreErr
	}
	w.minimised = false
	w.visible = true
	return nil
}
func (w *simulatedGameWindow) Show() {}
func (w *simulatedGameWindow) Fullscreen() {
	if !w.minimised {
		w.fullscreen = true
	}
	w.overlay = false
}
func (w *simulatedGameWindow) Overlay(enabled bool) error {
	if w.overlayErr != nil {
		return w.overlayErr
	}
	w.overlay = enabled
	return nil
}
func (w *simulatedGameWindow) Focus() error {
	if w.focusErr != nil {
		return w.focusErr
	}
	w.focused = w.visible && !w.minimised
	return nil
}

func TestPresentMinimisedFullscreenOverlay(t *testing.T) {
	w := &simulatedGameWindow{minimised: true, fullscreen: true}
	if err := presentGameWindow(w, true); err != nil {
		t.Fatal(err)
	}
	if w.minimised || !w.visible || !w.fullscreen || !w.overlay || !w.focused {
		t.Fatalf("F10 must restore fullscreen above the game: %+v", w)
	}
}

func TestReturnToLibraryReleasesOverlay(t *testing.T) {
	w := &simulatedGameWindow{overlay: true}
	if err := presentGameWindow(w, false); err != nil {
		t.Fatal(err)
	}
	if !w.visible || !w.fullscreen || !w.focused || w.overlay {
		t.Fatalf("library must be visible without overlay priority: %+v", w)
	}
}

func TestPresentGameWindowReportsFailures(t *testing.T) {
	failure := errors.New("window operation failed")
	for _, stage := range []string{"restore", "overlay", "focus"} {
		t.Run(stage, func(t *testing.T) {
			w := &simulatedGameWindow{minimised: true}
			switch stage {
			case "restore":
				w.restoreErr = failure
			case "overlay":
				w.overlayErr = failure
			case "focus":
				w.focusErr = failure
			}
			if err := presentGameWindow(w, true); !errors.Is(err, failure) {
				t.Fatalf("window failure was discarded: %v", err)
			}
			if w.focused {
				t.Fatal("failed presentation must not report a focused window")
			}
		})
	}
}
