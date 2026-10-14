//go:build windows

package launcher

import (
	"fmt"
	"golang.org/x/sys/windows"
	"log"
	"os"
	"syscall"
	"time"
	"unsafe"
)

var user32 = windows.NewLazySystemDLL("user32.dll")
var enumWindows = user32.NewProc("EnumWindows")
var windowPID = user32.NewProc("GetWindowThreadProcessId")
var visible = user32.NewProc("IsWindowVisible")
var foreground = user32.NewProc("SetForegroundWindow")
var show = user32.NewProc("ShowWindow")
var iconic = user32.NewProc("IsIconic")
var windowClass = user32.NewProc("GetClassNameW")
var position = user32.NewProc("SetWindowPos")
var post = user32.NewProc("PostMessageW")
var key = user32.NewProc("GetAsyncKeyState")
var registerHotKey = user32.NewProc("RegisterHotKey")
var unregisterHotKey = user32.NewProc("UnregisterHotKey")
var peekMessage = user32.NewProc("PeekMessageW")

const (
	menuHotKeyID = 0x4745
	wmHotKey     = 0x0312
	vkF10        = 0x79
	modNoRepeat  = 0x4000
)

// MSG uses pointer-sized HWND/WPARAM/LPARAM and includes lPrivate.
type hotKeyMessage struct {
	Window  uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	X, Y    int32
	Private uint32
}

type native struct {
	overlay          nativeOverlay
	held             bool
	state            *windows.LazyProc
	extended         uintptr
	nextCheck        [4]time.Time
	hotKeyActive     bool
	hotKeyRegistered bool
}

func Native() Platform {
	dll := windows.NewLazySystemDLL("xinput1_4.dll")
	n := &native{state: dll.NewProc("XInputGetState")}
	// Optional compatibility path also used by Chromium: Guide is absent from
	// the public XInputGetState API. Always retain the documented fallback.
	if dll.Load() == nil {
		n.extended, _ = windows.GetProcAddressByOrdinal(windows.Handle(dll.Handle()), 100)
	}
	return n
}
func FocusHost() error {
	hwnd, err := hostWindow()
	if err != nil {
		return err
	}
	ok, _, _ := foreground.Call(hwnd)
	if ok == 0 {
		return fmt.Errorf("Windows no permitió enfocar Gemu; selecciónalo con Alt+Tab")
	}
	return nil
}

// The host can be hidden/minimised. Do not use the game's visible-window
// heuristic, which can select a WebView popup instead of the Wails main window.
// Register one callback: Windows callback trampolines are never released.
var hostWindowCallback = windows.NewCallback(func(hwnd, parameter uintptr) uintptr {
	var pid uint32
	windowPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid != uint32(os.Getpid()) {
		return 1
	}
	var class [256]uint16
	windowClass.Call(hwnd, uintptr(unsafe.Pointer(&class[0])), uintptr(len(class)))
	if windows.UTF16ToString(class[:]) == "wailsWindow" {
		*(*uintptr)(unsafe.Pointer(parameter)) = hwnd
		return 0
	}
	return 1
})

func hostWindow() (uintptr, error) {
	var found uintptr
	enumWindows.Call(hostWindowCallback, uintptr(unsafe.Pointer(&found)))
	if found == 0 {
		return 0, fmt.Errorf("No se encontró la ventana de Gemu")
	}
	return found, nil
}

// Wails skips WindowUnminimise while its fullscreen flag is set. Restore the
// actual HWND synchronously before Wails applies fullscreen and queues Show.
func RestoreHost() error {
	hwnd, err := hostWindow()
	if err != nil {
		return err
	}
	if minimised, _, _ := iconic.Call(hwnd); minimised != 0 {
		show.Call(hwnd, 9) // SW_RESTORE
	} else {
		show.Call(hwnd, 5) // SW_SHOW, including a hidden host
	}
	return nil
}

func SetHostOverlay(enabled bool) error {
	hwnd, err := hostWindow()
	if err != nil {
		return err
	}
	insertAfter := ^uintptr(1) // HWND_NOTOPMOST (-2)
	if enabled {
		insertAfter = ^uintptr(0) // HWND_TOPMOST (-1)
	}
	// Demoting the overlay must not reactivate Gemu after the game gets focus.
	ok, _, _ := position.Call(hwnd, insertAfter, 0, 0, 0, 0, 0x0001|0x0002|0x0010)
	if ok == 0 {
		return fmt.Errorf("No se pudo cambiar la prioridad de la ventana de Gemu")
	}
	return nil
}
func (*native) Open(uri string) error {
	verb, _ := windows.UTF16PtrFromString("open")
	file, err := windows.UTF16PtrFromString(uri)
	if err != nil {
		return err
	}
	return windows.ShellExecute(0, verb, file, nil, nil, 1)
}
func process(pid uint32) (Process, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return Process{}, err
	}
	defer windows.CloseHandle(h)
	var created, exit, kernel, user windows.Filetime
	if err = windows.GetProcessTimes(h, &created, &exit, &kernel, &user); err != nil {
		return Process{}, err
	}
	buf := make([]uint16, 32768)
	size := uint32(len(buf))
	if err = windows.QueryFullProcessImageName(h, 0, &buf[0], &size); err != nil {
		return Process{}, err
	}
	return Process{PID: pid, Created: uint64(created.HighDateTime)<<32 | uint64(created.LowDateTime), Path: windows.UTF16ToString(buf[:size])}, nil
}
func (*native) Processes() ([]Process, error) {
	h, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(h)
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	result := []Process{}
	err = windows.Process32First(h, &entry)
	for err == nil {
		if p, e := process(entry.ProcessID); e == nil {
			result = append(result, p)
		}
		err = windows.Process32Next(h, &entry)
	}
	if err != windows.ERROR_NO_MORE_FILES {
		return nil, err
	}
	return result, nil
}
func window(p Process) (uintptr, error) {
	current, err := process(p.PID)
	if err != nil || current != p {
		return 0, fmt.Errorf("El proceso del juego ya terminó")
	}
	var found uintptr
	callback := windows.NewCallback(func(hwnd, parameter uintptr) uintptr {
		var pid uint32
		windowPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
		v, _, _ := visible.Call(hwnd)
		if pid == p.PID && v != 0 {
			found = hwnd
			return 0
		}
		return 1
	})
	enumWindows.Call(callback, 0)
	if found == 0 {
		return 0, fmt.Errorf("No se encontró una ventana del juego")
	}
	return found, nil
}
func (*native) Focus(p Process) error {
	hwnd, err := window(p)
	if err != nil {
		return err
	}
	// Restoring an already visible window can undo its maximised/fullscreen size.
	if minimised, _, _ := iconic.Call(hwnd); minimised != 0 {
		show.Call(hwnd, 9)
	}
	ok, _, _ := foreground.Call(hwnd)
	if ok == 0 {
		return fmt.Errorf("Windows no permitió enfocar el juego; selecciónalo con Alt+Tab")
	}
	return nil
}
func (*native) Close(p Process) error {
	hwnd, err := window(p)
	if err != nil {
		return err
	}
	ok, _, _ := post.Call(hwnd, 0x0010, 0, 0)
	if ok == 0 {
		return fmt.Errorf("No se pudo solicitar el cierre del juego")
	}
	return nil
}
func (n *native) Shortcut(active bool) bool {
	if active != n.hotKeyActive {
		n.hotKeyActive = active
		if active {
			ok, _, err := registerHotKey.Call(0, menuHotKeyID, modNoRepeat, vkF10)
			n.hotKeyRegistered = ok != 0
			if !n.hotKeyRegistered {
				log.Printf("No se pudo registrar F10; se mantiene el sondeo de teclado: %v", err)
			}
		} else if n.hotKeyRegistered {
			unregisterHotKey.Call(0, menuHotKeyID)
			n.hotKeyRegistered = false
		}
	}
	triggered := false
	for {
		var message hotKeyMessage
		ok, _, _ := peekMessage.Call(uintptr(unsafe.Pointer(&message)), ^uintptr(0), wmHotKey, wmHotKey, 1)
		if ok == 0 {
			break
		}
		if active && message.Message == wmHotKey && message.WParam == menuHotKeyID {
			triggered = true
		}
	}
	if !active {
		n.held = false
		return false
	}
	down := false
	if !n.hotKeyRegistered {
		pressed, _, _ := key.Call(vkF10)
		down = pressed&0x8000 != 0
	}
	if n.state.Find() == nil {
		for slot := 0; slot < 4; slot++ {
			if time.Now().Before(n.nextCheck[slot]) {
				continue
			}
			var state [5]uint32
			var status uintptr
			if n.extended != 0 {
				status, _, _ = syscall.SyscallN(n.extended, uintptr(slot), uintptr(unsafe.Pointer(&state)))
			} else {
				status, _, _ = n.state.Call(uintptr(slot), uintptr(unsafe.Pointer(&state)))
			}
			if status == 0 && (state[1]&0x30 == 0x30 || state[1]&0x400 != 0) {
				down = true
			}
			if status != 0 {
				n.nextCheck[slot] = time.Now().Add(2 * time.Second)
			}
		}
	}
	edge := down && !n.held
	n.held = down
	return triggered || edge
}
