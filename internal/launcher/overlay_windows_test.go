//go:build windows

package launcher

import "testing"

func TestExperimentalOverlayRequiresExplicitSelection(t *testing.T) {
	for _, value := range []string{"", "1", "dx9", "dx10", "vulkan"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("GEMU_EXPERIMENTAL_OVERLAY", value)
			n := &native{}
			if n.SyncOverlay([]Process{{PID: 1}}, Session{State: "running", Menu: true}) {
				t.Fatal("unvalidated renderer was enabled automatically")
			}
			if n.overlay.pending != nil || n.overlay.attempted != nil || n.overlay.client != nil {
				t.Fatal("default route must not inspect or attach to a game")
			}
		})
	}
}
func TestDisablingExperimentCancelsPendingAttachment(t *testing.T) {
	t.Setenv("GEMU_EXPERIMENTAL_OVERLAY", "")
	cancel := make(chan struct{})
	n := &native{overlay: nativeOverlay{pending: make(chan overlayResult), cancel: cancel}}
	n.SyncOverlay(nil, Session{})
	select {
	case <-cancel:
	default:
		t.Fatal("attachment was not cancelled")
	}
	if n.overlay.pending != nil {
		t.Fatal("pending attachment survived disabling the experiment")
	}
}
