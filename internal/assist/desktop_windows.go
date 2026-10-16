//go:build windows

package assist

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	ui             = windows.NewLazySystemDLL("user32.dll")
	gdi            = windows.NewLazySystemDLL("gdi32.dll")
	shell          = windows.NewLazySystemDLL("shell32.dll")
	defWindowProc  = ui.NewProc("DefWindowProcW")
	postMessage    = ui.NewProc("PostMessageW")
	getForeground  = ui.NewProc("GetForegroundWindow")
	getWindowPID   = ui.NewProc("GetWindowThreadProcessId")
	showWindow     = ui.NewProc("ShowWindow")
	setWindowPos   = ui.NewProc("SetWindowPos")
	invalidateRect = ui.NewProc("InvalidateRect")
	sendInput      = ui.NewProc("SendInput")
	activeDesktop  *desktop
	windowCallback = windows.NewCallback(windowProc)
)

type point struct{ X, Y int32 }
type rect struct{ Left, Top, Right, Bottom int32 }
type message struct {
	Window         uintptr
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Point          point
	Private        uint32
}
type windowClass struct {
	Size, Style                        uint32
	Proc                               uintptr
	ClassExtra, WindowExtra            int32
	Instance, Icon, Cursor, Background uintptr
	MenuName, ClassName                *uint16
	SmallIcon                          uintptr
}
type paintStruct struct {
	DC                 uintptr
	Erase              int32
	Paint              rect
	Restore, IncUpdate int32
	Reserved           [32]byte
}
type monitorInfo struct {
	Size          uint32
	Monitor, Work rect
	Flags         uint32
}
type notifyIcon struct {
	Size                uint32
	Window              uintptr
	ID, Flags, Callback uint32
	Icon                uintptr
	Tip                 [128]uint16
	State, StateMask    uint32
	Info                [256]uint16
	Timeout             uint32
	InfoTitle           [64]uint16
	InfoFlags           uint32
	GUID                [16]byte
	Balloon             uintptr
}
type mouseInput struct {
	X, Y              int32
	Data, Flags, Time uint32
	Extra             uintptr
}
type nativeInput struct {
	Kind  uint32
	Mouse mouseInput
}
type keyboardInput struct {
	VK, Scan    uint16
	Flags, Time uint32
	Extra       uintptr
}
type padState struct {
	Packet         uint32
	Buttons        uint16
	LT, RT         byte
	LX, LY, RX, RY int16
}
type command struct {
	request Request
	reply   chan Status
}

type desktop struct {
	window                        uintptr
	mode                          Mode
	target, lastExternal          uintptr
	hostPID                       uint32
	hostSeen                      time.Time
	session                       bool
	menu                          uint64
	row, col                      int
	upper, symbols                bool
	previous                      uint16
	device                        int
	shortcuts                     Shortcuts
	repeat                        navigationRepeat
	lastTick                      time.Time
	scrollRemainder, moveX, moveY float64
	leftHeld                      bool
	armed                         bool
	top                           bool
	err                           string
	commands                      chan command
	stop                          chan struct{}
	tray                          notifyIcon
	font                          uintptr
	taskbarCreated                uint32
	xinput                        *windows.LazyProc
	extended                      uintptr
	nextCheck                     [4]time.Time
}

func wide(value string) *uint16 { result, _ := windows.UTF16PtrFromString(value); return result }

func Run() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pipe, err := createPipe()
	if err != nil {
		// Starting another Gemu instance reuses the existing companion.
		if _, live := exchange(Request{Action: "status"}); live == nil {
			return nil
		}
		return fmt.Errorf("create input companion: %w", err)
	}
	app := &desktop{mode: Off, device: -1, commands: make(chan command, 8), stop: make(chan struct{})}
	activeDesktop = app
	defer func() { activeDesktop = nil; app.release(); close(app.stop) }()
	// Per-monitor V2 keeps cursor/window coordinates in the same physical space.
	if proc := ui.NewProc("SetProcessDpiAwarenessContext"); proc.Find() == nil {
		proc.Call(^uintptr(3))
	}
	instance, _, _ := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetModuleHandleW").Call(0)
	icon, _, _ := ui.NewProc("LoadIconW").Call(0, 32512)
	cursor, _, _ := ui.NewProc("LoadCursorW").Call(0, 32512)
	class := windowClass{Size: uint32(unsafe.Sizeof(windowClass{})), Proc: windowCallback, Instance: uintptr(instance), Icon: icon, Cursor: cursor, ClassName: wide("GemuInputCompanion")}
	ok, _, err := ui.NewProc("RegisterClassExW").Call(uintptr(unsafe.Pointer(&class)))
	if ok == 0 {
		windows.CloseHandle(pipe)
		return err
	}
	// WS_EX_NOACTIVATE | TOOLWINDOW | TOPMOST. The editor/login retains keyboard focus.
	window, _, err := ui.NewProc("CreateWindowExW").Call(0x08000088, uintptr(unsafe.Pointer(class.ClassName)), uintptr(unsafe.Pointer(wide("Gemu · Asistencia"))), 0x80000000, 0, 0, 900, 400, 0, 0, uintptr(instance), 0)
	if window == 0 {
		windows.CloseHandle(pipe)
		return err
	}
	app.window = window
	defer ui.NewProc("DestroyWindow").Call(window)
	app.font, _, _ = gdi.NewProc("CreateFontW").Call(20, 0, 0, 0, 400, 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(wide("Segoe UI"))))
	defer gdi.NewProc("DeleteObject").Call(app.font)
	app.tray = notifyIcon{Size: uint32(unsafe.Sizeof(notifyIcon{})), Window: window, ID: 1, Flags: 1 | 2 | 4, Callback: 0x8002, Icon: icon}
	copy(app.tray.Tip[:], utf16.Encode([]rune("Gemu · Asistencia con mando")))
	app.addTray()
	defer shell.NewProc("Shell_NotifyIconW").Call(2, uintptr(unsafe.Pointer(&app.tray)))
	app.taskbarCreated = uint32Result(ui.NewProc("RegisterWindowMessageW").Call(uintptr(unsafe.Pointer(wide("TaskbarCreated")))))
	dll := windows.NewLazySystemDLL("xinput1_4.dll")
	app.xinput = dll.NewProc("XInputGetState")
	if dll.Load() == nil {
		app.extended, _ = windows.GetProcAddressByOrdinal(windows.Handle(dll.Handle()), 100)
	}
	go servePipe(pipe, app.stop, func(request Request) Status {
		cmd := command{request: request, reply: make(chan Status, 1)}
		select {
		case app.commands <- cmd:
		case <-app.stop:
			return Status{Version: protocolVersion, Error: "companion stopped"}
		default:
			return Status{Version: protocolVersion, Error: "companion busy"}
		}
		postMessage.Call(window, 0x8001, 0, 0)
		select {
		case status := <-cmd.reply:
			return status
		case <-time.After(200 * time.Millisecond):
			return Status{Version: protocolVersion, Error: "companion did not respond"}
		}
	})
	ui.NewProc("SetTimer").Call(window, 1, 16, 0)
	defer ui.NewProc("KillTimer").Call(window, 1)
	var msg message
	for {
		result, _, err := ui.NewProc("GetMessageW").Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(result) == -1 {
			return err
		}
		if result == 0 {
			return nil
		}
		ui.NewProc("TranslateMessage").Call(uintptr(unsafe.Pointer(&msg)))
		ui.NewProc("DispatchMessageW").Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func uint32Result(value uintptr, _ uintptr, _ error) uint32 { return uint32(value) }
func (d *desktop) addTray() {
	shell.NewProc("Shell_NotifyIconW").Call(0, uintptr(unsafe.Pointer(&d.tray)))
}
func windowProc(window uintptr, msg uint32, wparam, lparam uintptr) uintptr {
	d := activeDesktop
	if d != nil {
		if msg == d.taskbarCreated && msg != 0 {
			d.addTray()
			return 0
		}
		switch msg {
		case 0x21:
			return 3 // MA_NOACTIVATE
		case 0x14:
			return 1 // background is painted in one pass
		case 0xf:
			d.paint(window)
			return 0
		case 0x113:
			d.tick()
			return 0
		case 0x8001:
			for {
				select {
				case cmd := <-d.commands:
					cmd.reply <- d.dispatch(cmd.request)
				default:
					return 0
				}
			}
		case 0x8002:
			if lparam == 0x202 || lparam == 0x205 {
				d.trayMenu()
			}
			return 0
		case 0x201:
			d.click(int(int16(lparam&0xffff)), int(int16((lparam>>16)&0xffff)))
			return 0
		case 0x10:
			d.setMode(Off)
			return 0
		case 0x2:
			ui.NewProc("PostQuitMessage").Call(0)
			return 0
		}
	}
	result, _, _ := defWindowProc.Call(window, uintptr(msg), wparam, lparam)
	return result
}

func (d *desktop) dispatch(r Request) Status {
	var err string
	switch r.Action {
	case "status":
		if r.HostPID != 0 {
			d.hostPID = r.HostPID
			d.hostSeen = time.Now()
			d.session = r.Session
		}
	case "pointer":
		d.setMode(Pointer)
	case "keyboard":
		d.setMode(Keyboard)
	case "hide":
		d.setMode(Off)
	default:
		err = "unknown input companion action"
	}
	return Status{Version: protocolVersion, Mode: d.mode, Menu: d.menu, Error: err}
}

func windowPID(window uintptr) uint32 {
	var pid uint32
	getWindowPID.Call(window, uintptr(unsafe.Pointer(&pid)))
	return pid
}
func (d *desktop) foreground() uintptr { value, _, _ := getForeground.Call(); return value }
func (d *desktop) rememberTarget() {
	window := d.foreground()
	if window != 0 && window != d.window && windowPID(window) != d.hostPID && windowPID(window) != uint32(os.Getpid()) {
		d.lastExternal = window
	}
}
func (d *desktop) setMode(mode Mode) {
	d.release()
	d.err = ""
	d.moveX = 0
	d.moveY = 0
	d.scrollRemainder = 0
	if mode == Off {
		d.mode = Off
		showWindow.Call(d.window, 0)
		return
	}
	d.rememberTarget()
	window := d.foreground()
	if windowPID(window) == d.hostPID || windowPID(window) == uint32(os.Getpid()) {
		window = d.lastExternal
	}
	if window == 0 {
		d.err = "Selecciona primero la ventana que quieres controlar."
		mode = Pointer
	}
	d.target = window
	d.mode = mode
	d.armed = false
	d.row = 1
	d.col = 0
	d.previous = 0xffff
	d.repeat = navigationRepeat{}
	if window != 0 && d.foreground() != window {
		ok, _, _ := ui.NewProc("SetForegroundWindow").Call(window)
		if ok == 0 {
			d.err = "Windows no permitió enfocar la ventana. Selecciónala para continuar."
		}
	}
	d.place()
	d.redraw()
}
func (d *desktop) place() {
	monitor, _, _ := ui.NewProc("MonitorFromWindow").Call(d.target, 2)
	info := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
	ui.NewProc("GetMonitorInfoW").Call(monitor, uintptr(unsafe.Pointer(&info)))
	width := int32(math.Min(1000, float64(info.Work.Right-info.Work.Left-24)))
	height := int32(390)
	if d.mode == Pointer {
		height = 108
	}
	height = min(height, info.Work.Bottom-info.Work.Top-24)
	x := info.Work.Left + (info.Work.Right-info.Work.Left-width)/2
	y := info.Work.Bottom - height - 12
	if d.top {
		y = info.Work.Top + 12
	}
	setWindowPos.Call(d.window, ^uintptr(0), uintptr(x), uintptr(y), uintptr(width), uintptr(height), 0x10|0x40)
}
func (d *desktop) redraw() { invalidateRect.Call(d.window, 0, 0) }

func (d *desktop) pads(now time.Time) (int, padState) {
	var chosen padState
	chosenIndex := -1
	if d.xinput.Find() != nil {
		return chosenIndex, chosen
	}
	for slot := 0; slot < 4; slot++ {
		if now.Before(d.nextCheck[slot]) {
			continue
		}
		var state padState
		var status uintptr
		if d.extended != 0 {
			status, _, _ = syscall.SyscallN(d.extended, uintptr(slot), uintptr(unsafe.Pointer(&state)))
		} else {
			status, _, _ = d.xinput.Call(uintptr(slot), uintptr(unsafe.Pointer(&state)))
		}
		if status != 0 {
			d.nextCheck[slot] = now.Add(2 * time.Second)
			continue
		}
		if chosenIndex < 0 || slot == d.device {
			chosen = state
			chosenIndex = slot
		}
		// Keep ownership during assistance; a second player must not take over.
		if slot == d.device && d.mode != Off {
			return slot, state
		}
		if d.mode == Off && state.Buttons != 0 {
			return slot, state
		}
	}
	return chosenIndex, chosen
}
func (d *desktop) tick() {
	now := time.Now()
	elapsed := math.Min(.032, now.Sub(d.lastTick).Seconds())
	d.lastTick = now
	d.rememberTarget()
	slot, pad := d.pads(now)
	if slot != d.device {
		d.release()
		if d.device >= 0 {
			d.setMode(Off)
		}
		d.device = slot
		d.previous = 0xffff
		d.shortcuts.Reset()
		d.repeat = navigationRepeat{}
	}
	if slot < 0 {
		return
	}
	foreground := d.foreground()
	// The existing HTML keyboard remains available inside Gemu. Native chords
	// belong to the companion only while controlling another application.
	hostForeground := foreground != 0 && windowPID(foreground) == d.hostPID
	mode, menu, suppress := d.shortcuts.Sample(pad.Buttons)
	if mode != "" && !hostForeground {
		if d.mode == mode {
			d.setMode(Off)
		} else {
			d.setMode(mode)
		}
	}
	if menu && d.session && now.Sub(d.hostSeen) < 2*time.Second {
		d.setMode(Off)
		d.menu++
	}
	pressed := pad.Buttons &^ d.previous
	d.previous = pad.Buttons
	if suppress || d.mode == Off {
		return
	}
	if !d.armed {
		if pad.Buttons == 0 {
			d.armed = true
		}
		return
	}
	if d.mode == Keyboard && foreground != d.target {
		d.setMode(Off)
		d.shortcuts.Reset()
		return
	}
	if pressed&B != 0 {
		d.setMode(Off)
		return
	}
	if d.mode == Pointer {
		if pressed&Y != 0 {
			d.setMode(Keyboard)
			return
		}
		d.moveX += velocity(float64(pad.LX)/32767) * 1000 * elapsed
		d.moveY -= velocity(float64(pad.LY)/32767) * 1000 * elapsed
		x, y := int32(d.moveX), int32(d.moveY)
		d.moveX -= float64(x)
		d.moveY -= float64(y)
		if x != 0 || y != 0 {
			d.inject([]nativeInput{{Mouse: mouseInput{X: x, Y: y, Flags: 1}}})
		}
		left := pad.Buttons&A != 0
		if left != d.leftHeld {
			flag := uint32(4)
			if left {
				flag = 2
			}
			if d.inject([]nativeInput{{Mouse: mouseInput{Flags: flag}}}) {
				d.leftHeld = left
			}
		}
		if pressed&X != 0 {
			d.inject([]nativeInput{{Mouse: mouseInput{Flags: 8}}, {Mouse: mouseInput{Flags: 16}}})
		}
		d.scrollRemainder += velocity(float64(pad.RY)/32767) * 960 * elapsed
		wheel := int32(d.scrollRemainder)
		d.scrollRemainder -= float64(wheel)
		if wheel != 0 {
			d.inject([]nativeInput{{Mouse: mouseInput{Data: uint32(wheel), Flags: 0x800}}})
		}
		if pressed&LB != 0 {
			d.key("tab", true)
		}
		if pressed&RB != 0 {
			d.key("tab", false)
		}
		return
	}
	if pressed&X != 0 {
		d.key("backspace", false)
	}
	if pressed&Y != 0 {
		d.upper = !d.upper
		d.redraw()
	}
	if pressed&LB != 0 {
		d.symbols = !d.symbols
		d.clampSelection()
		d.redraw()
	}
	if pressed&RB != 0 {
		d.text(" ")
	}
	if pressed&A != 0 {
		d.activate(keyRows(d.symbols, d.upper)[d.row][d.col])
		return
	}
	direction := pad.Buttons & (Up | Down | Left | Right)
	if direction == 0 {
		switch {
		case pad.LX > 16000:
			direction = Right
		case pad.LX < -16000:
			direction = Left
		case pad.LY > 16000:
			direction = Up
		case pad.LY < -16000:
			direction = Down
		}
	}
	if d.repeat.sample(direction, now) {
		rows := keyRows(d.symbols, d.upper)
		switch {
		case direction&Up != 0:
			d.row = (d.row + len(rows) - 1) % len(rows)
		case direction&Down != 0:
			d.row = (d.row + 1) % len(rows)
		case direction&Left != 0:
			d.col = (d.col + len(rows[d.row]) - 1) % len(rows[d.row])
		case direction&Right != 0:
			d.col = (d.col + 1) % len(rows[d.row])
		}
		d.clampSelection()
		d.redraw()
	}
}
func (d *desktop) clampSelection() { d.col = min(d.col, len(keyRows(d.symbols, d.upper)[d.row])-1) }
func (d *desktop) release() {
	if d.leftHeld {
		d.inject([]nativeInput{{Mouse: mouseInput{Flags: 4}}})
		d.leftHeld = false
	}
}
func (d *desktop) inject(events []nativeInput) bool {
	if len(events) == 0 {
		return true
	}
	inserted, _, _ := sendInput.Call(uintptr(len(events)), uintptr(unsafe.Pointer(&events[0])), unsafe.Sizeof(nativeInput{}))
	runtime.KeepAlive(events)
	if int(inserted) != len(events) {
		d.err = "No se pudo enviar la entrada. Comprueba el foco y los permisos de la ventana."
		d.redraw()
		return false
	}
	return true
}
func keyInput(vk, scan uint16, flags uint32) nativeInput {
	input := nativeInput{Kind: 1}
	*(*keyboardInput)(unsafe.Pointer(&input.Mouse)) = keyboardInput{VK: vk, Scan: scan, Flags: flags}
	return input
}
func (d *desktop) text(value string) {
	if d.foreground() != d.target {
		d.setMode(Off)
		return
	}
	events := []nativeInput{}
	for _, unit := range utf16.Encode([]rune(value)) {
		events = append(events, keyInput(0, unit, 4), keyInput(0, unit, 4|2))
	}
	d.inject(events)
}
func (d *desktop) key(command string, shift bool) {
	keys := map[string]uint16{"tab": 9, "enter": 13, "escape": 27, "backspace": 8, "left": 37, "up": 38, "right": 39, "down": 40}
	vk := keys[command]
	if vk == 0 {
		return
	}
	flags := uint32(0)
	if vk >= 37 && vk <= 40 {
		flags = 1
	}
	events := []nativeInput{}
	if shift {
		events = append(events, keyInput(16, 0, 0))
	}
	events = append(events, keyInput(vk, 0, flags), keyInput(vk, 0, flags|2))
	if shift {
		events = append(events, keyInput(16, 0, 2))
	}
	d.inject(events)
}
func (d *desktop) activate(key Key) {
	switch key.Command {
	case "shift":
		d.upper = !d.upper
	case "symbols":
		d.symbols = !d.symbols
		d.clampSelection()
	case "pointer":
		d.setMode(Pointer)
	case "close":
		d.setMode(Off)
	case "move":
		d.top = !d.top
		d.place()
	case "":
		d.text(key.Text)
	default:
		d.key(key.Command, false)
	}
	d.redraw()
}

func (d *desktop) trayMenu() {
	menu, _, _ := ui.NewProc("CreatePopupMenu").Call()
	defer ui.NewProc("DestroyMenu").Call(menu)
	for index, label := range []string{"Ratón virtual", "Teclado virtual", "Ocultar asistencia", "Salir del asistente"} {
		ui.NewProc("AppendMenuW").Call(menu, 0, uintptr(index+1), uintptr(unsafe.Pointer(wide(label))))
	}
	var position point
	ui.NewProc("GetCursorPos").Call(uintptr(unsafe.Pointer(&position)))
	// Tray menus need a foreground owner. Remember the actual target first.
	d.rememberTarget()
	ui.NewProc("SetForegroundWindow").Call(d.window)
	choice, _, _ := ui.NewProc("TrackPopupMenu").Call(menu, 0x100|0x2, uintptr(position.X), uintptr(position.Y), 0, d.window, 0)
	switch choice {
	case 1:
		d.setMode(Pointer)
	case 2:
		d.setMode(Keyboard)
	case 3:
		d.setMode(Off)
	case 4:
		ui.NewProc("DestroyWindow").Call(d.window)
	}
	postMessage.Call(d.window, 0, 0, 0)
}

func fill(dc uintptr, bounds rect, color uint32) {
	brush, _, _ := gdi.NewProc("CreateSolidBrush").Call(uintptr(color))
	ui.NewProc("FillRect").Call(dc, uintptr(unsafe.Pointer(&bounds)), brush)
	gdi.NewProc("DeleteObject").Call(brush)
}
func drawText(dc uintptr, value string, bounds rect, color uint32, flags uintptr) {
	gdi.NewProc("SetTextColor").Call(dc, uintptr(color))
	ui.NewProc("DrawTextW").Call(dc, uintptr(unsafe.Pointer(wide(value))), ^uintptr(0), uintptr(unsafe.Pointer(&bounds)), flags)
}
func (d *desktop) clientRect() rect {
	var bounds rect
	ui.NewProc("GetClientRect").Call(d.window, uintptr(unsafe.Pointer(&bounds)))
	return bounds
}
func (d *desktop) paint(window uintptr) {
	var paint paintStruct
	dc, _, _ := ui.NewProc("BeginPaint").Call(window, uintptr(unsafe.Pointer(&paint)))
	defer ui.NewProc("EndPaint").Call(window, uintptr(unsafe.Pointer(&paint)))
	bounds := d.clientRect()
	fill(dc, bounds, 0x201810)
	gdi.NewProc("SetBkMode").Call(dc, 1)
	oldFont, _, _ := gdi.NewProc("SelectObject").Call(dc, d.font)
	defer gdi.NewProc("SelectObject").Call(dc, oldFont)
	title := "Ratón virtual · Y: teclado · B: salir"
	if d.mode == Keyboard {
		title = "Teclado virtual · A: escribir · X: borrar · Y: mayúsculas · B: salir"
	}
	drawText(dc, title, rect{16, 8, bounds.Right - 16, 38}, 0x3fffD9, 0x24)
	if d.mode == Keyboard {
		rows := keyRows(d.symbols, d.upper)
		for r, row := range rows {
			for c, key := range row {
				box := d.keyRect(r, c, bounds)
				color := uint32(0x342c24)
				text := uint32(0xe6eff2)
				if r == d.row && c == d.col {
					color = 0x3fffD9
					text = 0x071913
				}
				fill(dc, box, color)
				drawText(dc, key.Label, box, text, 0x25)
			}
		}
	}
	help := "Stick izq.: mover · A: mantener/arrastrar · X: clic derecho · Stick der.: desplazar · LB/RB: Tab"
	if d.mode == Keyboard {
		help = "Cruceta/stick: elegir · LB: símbolos · RB: espacio · Home + R3: ratón · Home + X: teclado"
	}
	if d.err != "" {
		help = d.err
	}
	drawText(dc, help, rect{16, bounds.Bottom - 52, bounds.Right - 16, bounds.Bottom - 6}, 0xc1b8b0, 0x10)
}
func (d *desktop) keyRect(row, col int, bounds rect) rect {
	rows := keyRows(d.symbols, d.upper)
	height := (bounds.Bottom - 104) / int32(len(rows))
	width := (bounds.Right - 24) / int32(len(rows[row]))
	return rect{12 + int32(col)*width, 44 + int32(row)*height, 12 + int32(col+1)*width - 4, 44 + int32(row+1)*height - 4}
}
func (d *desktop) click(x, y int) {
	if d.mode != Keyboard {
		return
	}
	bounds := d.clientRect()
	for r, row := range keyRows(d.symbols, d.upper) {
		for c, key := range row {
			box := d.keyRect(r, c, bounds)
			if int32(x) >= box.Left && int32(x) < box.Right && int32(y) >= box.Top && int32(y) < box.Bottom {
				d.row = r
				d.col = c
				d.activate(key)
				return
			}
		}
	}
}
