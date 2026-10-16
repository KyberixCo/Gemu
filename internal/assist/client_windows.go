//go:build windows

package assist

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"
)

var connected atomic.Bool
var sessionActive atomic.Bool
var menuPending atomic.Bool
var currentMode atomic.Value

func Healthy() bool     { return connected.Load() }
func CurrentMode() Mode { mode, _ := currentMode.Load().(Mode); return mode }
func MenuShortcut(active bool) (bool, bool) {
	sessionActive.Store(active)
	return menuPending.Swap(false) && active, connected.Load()
}

// Connect starts a per-user companion if necessary. The companion has its own
// lifecycle: cancellation stops this client, never the tray application.
func Connect(ctx context.Context) error {
	if _, err := exchange(Request{Action: "status"}); err != nil {
		executable, err := os.Executable()
		if err != nil {
			return err
		}
		path := filepath.Join(filepath.Dir(executable), "gemu-input.exe")
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("input companion unavailable: %w", err)
		}
		command := exec.Command(path)
		command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000008 | 0x00000200, HideWindow: true}
		if err := command.Start(); err != nil {
			return err
		}
		go command.Wait()
	}
	go func() {
		defer connected.Store(false)
		var previous uint64
		initialized := false
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			status, err := exchange(Request{Action: "status", HostPID: uint32(os.Getpid()), Session: sessionActive.Load()})
			if err != nil {
				connected.Store(false)
				currentMode.Store(Off)
				continue
			}
			if initialized && status.Menu != previous {
				menuPending.Store(true)
			}
			previous = status.Menu
			initialized = true
			currentMode.Store(status.Mode)
			connected.Store(true)
		}
	}()
	return nil
}

func Open(mode Mode) error {
	action := "pointer"
	if mode == Keyboard {
		action = "keyboard"
	}
	_, err := exchange(Request{Action: action})
	return err
}
