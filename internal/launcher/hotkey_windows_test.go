//go:build windows

package launcher

import (
	"runtime"
	"testing"

	"golang.org/x/sys/windows"
)

func TestControllerMenuShortcutSupportsGuide(t *testing.T) {
	for _, buttons := range []uint32{0, 0x10, 0x20, 0x4000} {
		if controllerMenuShortcut(buttons) {
			t.Errorf("buttons %#x must not open Gemu's menu", buttons)
		}
	}
	for _, buttons := range []uint32{0x30, 0x400, 0x400 | 0x30} {
		if !controllerMenuShortcut(buttons) {
			t.Errorf("buttons %#x must open Gemu's menu", buttons)
		}
	}
}

func TestNativeHotKeyQueueAndRelease(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	n := &native{state: user32.NewProc("GemuMissingXInputProc")}
	defer n.Shortcut(false)
	n.Shortcut(true)
	if !n.hotKeyRegistered {
		t.Skip("F10 is already registered by another application")
	}
	postThreadMessage := user32.NewProc("PostThreadMessageW")
	postHotKey := func(id uintptr) {
		t.Helper()
		ok, _, err := postThreadMessage.Call(uintptr(windows.GetCurrentThreadId()), wmHotKey, id, vkF10<<16)
		if ok == 0 {
			t.Fatalf("post hotkey: %v", err)
		}
	}
	// A queued press remains available after the key has been released.
	postHotKey(menuHotKeyID)
	if !n.Shortcut(true) || n.Shortcut(true) {
		t.Fatal("queued F10 must trigger exactly once")
	}
	postHotKey(menuHotKeyID + 1)
	if n.Shortcut(true) {
		t.Fatal("unrelated hotkey opened the menu")
	}
	postHotKey(menuHotKeyID)
	if n.Shortcut(false) || n.hotKeyRegistered {
		t.Fatal("inactive session must release F10 and discard queued presses")
	}
	// Prove the registration was actually released, not just the Go flag.
	ok, _, err := registerHotKey.Call(0, menuHotKeyID+1, modNoRepeat, vkF10)
	if ok == 0 {
		t.Fatalf("F10 was not released: %v", err)
	}
	unregisterHotKey.Call(0, menuHotKeyID+1)
	n.Shortcut(true)
	if !n.hotKeyRegistered || n.Shortcut(true) {
		t.Fatal("next session must register F10 without replaying an old press")
	}
}
