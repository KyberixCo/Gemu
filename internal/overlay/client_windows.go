//go:build windows && amd64

package overlay

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

var kernel = windows.NewLazySystemDLL("kernel32.dll")
var allocate = kernel.NewProc("VirtualAllocEx")
var free = kernel.NewProc("VirtualFreeEx")
var createThread = kernel.NewProc("CreateRemoteThread")
var threadExit = kernel.NewProc("GetExitCodeThread")
var ticks = kernel.NewProc("GetTickCount")

type Client struct {
	mapping windows.Handle
	address uintptr
	shared  *Shared
}

func tick() uint32 { value, _, _ := ticks.Call(); return uint32(value) }

func HasD3D11(pid uint32) bool {
	_, err := moduleBase(pid, "d3d11.dll")
	return err == nil
}

func moduleBase(pid uint32, name string) (uintptr, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPMODULE, pid)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ModuleEntry32{Size: uint32(unsafe.Sizeof(windows.ModuleEntry32{}))}
	for err = windows.Module32First(snapshot, &entry); err == nil; err = windows.Module32Next(snapshot, &entry) {
		if strings.EqualFold(windows.UTF16ToString(entry.Module[:]), name) {
			return uintptr(unsafe.Pointer(entry.ModBaseAddr)), nil
		}
	}
	return 0, fmt.Errorf("module %s unavailable", name)
}

func Attach(pid uint32, dll, title, language string) (*Client, error) {
	if !filepath.IsAbs(dll) {
		return nil, fmt.Errorf("overlay DLL requires an absolute path")
	}
	process, err := windows.OpenProcess(windows.PROCESS_CREATE_THREAD|windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_OPERATION|windows.PROCESS_VM_WRITE|windows.PROCESS_VM_READ, false, pid)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(process)
	var machine, native uint16
	if err = windows.IsWow64Process2(process, &machine, &native); err != nil {
		return nil, err
	}
	if machine != 0 || native != 0x8664 {
		return nil, fmt.Errorf("overlay supports x64 processes only")
	}
	name, _ := windows.UTF16PtrFromString(fmt.Sprintf("Local\\GemuOverlay-%d", pid))
	mapping, err := windows.CreateFileMapping(windows.InvalidHandle, nil, windows.PAGE_READWRITE, 0, uint32(unsafe.Sizeof(Shared{})), name)
	if err != nil {
		if mapping != 0 {
			windows.CloseHandle(mapping)
		}
		return nil, err
	}
	address, err := windows.MapViewOfFile(mapping, windows.FILE_MAP_READ|windows.FILE_MAP_WRITE, 0, 0, unsafe.Sizeof(Shared{}))
	if err != nil {
		windows.CloseHandle(mapping)
		return nil, err
	}
	c := &Client{mapping: mapping, address: address, shared: (*Shared)(unsafe.Pointer(address))}
	*c.shared = Shared{Version: Version, Size: uint32(unsafe.Sizeof(Shared{}))}
	c.shared.configure(title, language, "running")
	c.Update(language, "running", false)
	if err = inject(process, pid, dll); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

func inject(process windows.Handle, pid uint32, dll string) error {
	path, err := windows.UTF16FromString(dll)
	if err != nil {
		return err
	}
	loaderDLL := windows.NewLazySystemDLL("kernelbase.dll")
	loader := loaderDLL.NewProc("LoadLibraryW")
	if err = loader.Find(); err != nil {
		return err
	}
	remoteBase, err := moduleBase(pid, "kernelbase.dll")
	if err != nil {
		return err
	}
	start := remoteBase + loader.Addr() - loaderDLL.Handle()
	size := uintptr(len(path)) * 2
	address, _, err := allocate.Call(uintptr(process), 0, size, windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_READWRITE)
	if address == 0 {
		return err
	}
	finished := true
	defer func() {
		if finished {
			free.Call(uintptr(process), address, 0, windows.MEM_RELEASE)
		}
	}()
	if err = windows.WriteProcessMemory(process, address, (*byte)(unsafe.Pointer(&path[0])), size, nil); err != nil {
		return err
	}
	thread, _, err := createThread.Call(uintptr(process), 0, 0, start, address, 0, 0)
	if thread == 0 {
		return err
	}
	defer windows.CloseHandle(windows.Handle(thread))
	status, err := windows.WaitForSingleObject(windows.Handle(thread), 3000)
	if err != nil || status != windows.WAIT_OBJECT_0 {
		// The remote loader may still be reading the path; never free it early.
		finished = false
		return fmt.Errorf("overlay loader did not finish: %v", err)
	}
	var result uint32
	threadExit.Call(thread, uintptr(unsafe.Pointer(&result)))
	if result == 0 {
		return fmt.Errorf("overlay DLL could not be loaded")
	}
	return nil
}

func (c *Client) Ready() bool {
	return c != nil && c.shared != nil && atomic.LoadUint32(&c.shared.Ready) == 1 && tick()-atomic.LoadUint32(&c.shared.RendererTick) < 3000
}
func (c *Client) Update(language, state string, visible bool) {
	if c == nil || c.shared == nil {
		return
	}
	c.shared.configure("", language, state)
	atomic.StoreUint32(&c.shared.HostTick, tick())
	value := uint32(0)
	if visible {
		value = 1
	}
	atomic.StoreUint32(&c.shared.Visible, value)
}
func (c *Client) Command() uint32 {
	if c == nil || c.shared == nil {
		return CommandNone
	}
	return atomic.SwapUint32(&c.shared.Command, CommandNone)
}
func (c *Client) Dimensions() (uint32, uint32) {
	if c == nil || c.shared == nil {
		return 0, 0
	}
	return atomic.LoadUint32(&c.shared.Width), atomic.LoadUint32(&c.shared.Height)
}
func (c *Client) Close() {
	if c == nil || c.shared == nil {
		return
	}
	atomic.StoreUint32(&c.shared.Visible, 0)
	atomic.StoreUint32(&c.shared.Stop, 1)
	windows.UnmapViewOfFile(c.address)
	windows.CloseHandle(c.mapping)
	c.shared = nil
}
