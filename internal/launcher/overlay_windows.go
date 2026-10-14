//go:build windows

package launcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"kyberix.com/gemu/internal/overlay"
)

type overlayResult struct {
	client  *overlay.Client
	process Process
	err     error
}
type nativeOverlay struct {
	client    *overlay.Client
	process   Process
	pending   chan overlayResult
	cancel    chan struct{}
	attempted map[Process]bool
	nextCheck time.Time
	language  string
}

func overlayDLL() string {
	executable, _ := os.Executable()
	for _, candidate := range []string{filepath.Join(filepath.Dir(executable), "gemu-overlay.dll"), filepath.Join("artifacts", "overlay", "gemu-overlay.dll")} {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			path, _ := filepath.Abs(candidate)
			return path
		}
	}
	return ""
}

func (n *native) SyncOverlay(processes []Process, session Session) bool {
	// Finding the experimental DLL must never enable attachment automatically.
	if os.Getenv("GEMU_EXPERIMENTAL_OVERLAY") != "dx11" {
		if n.overlay.client != nil || n.overlay.pending != nil {
			n.ReleaseOverlay()
		}
		return false
	}
	o := &n.overlay
	if o.pending != nil {
		select {
		case result := <-o.pending:
			o.pending = nil
			o.cancel = nil
			if result.err != nil {
				log.Printf("Overlay DX11 no disponible para PID %d: %v", result.process.PID, result.err)
			} else {
				o.client = result.client
				o.process = result.process
			}
		default:
		}
	}
	if o.client != nil {
		o.client.Update(o.language, session.State, session.Menu && session.Overlay)
		return o.client.Ready()
	}
	if o.pending != nil || time.Now().Before(o.nextCheck) {
		return false
	}
	o.nextCheck = time.Now().Add(time.Second)
	dll := overlayDLL()
	if dll == "" {
		return false
	}
	if o.attempted == nil {
		o.attempted = map[Process]bool{}
	}
	for _, p := range processes {
		if o.attempted[p] || !overlay.HasD3D11(p.PID) {
			continue
		}
		o.attempted[p] = true
		o.pending = make(chan overlayResult)
		o.cancel = make(chan struct{})
		pending, cancel, language := o.pending, o.cancel, o.language
		go func() {
			result := overlayResult{process: p}
			if current, err := process(p.PID); err != nil || current != p {
				result.err = fmt.Errorf("game identity changed before overlay attachment")
			} else {
				result.client, result.err = overlay.Attach(p.PID, dll, session.Title, language)
			}
			select {
			case pending <- result:
			case <-cancel:
				if result.client != nil {
					result.client.Close()
				}
			}
		}()
		break
	}
	return false
}
func (n *native) OverlayCommand() uint32 {
	if n.overlay.client == nil {
		return 0
	}
	return n.overlay.client.Command()
}
func (n *native) OverlayLanguage(language string) { n.overlay.language = language }
func (n *native) ReleaseOverlay() {
	o := &n.overlay
	if o.cancel != nil {
		close(o.cancel)
	}
	if o.client != nil {
		o.client.Close()
	}
	language := o.language
	*o = nativeOverlay{language: language}
}
