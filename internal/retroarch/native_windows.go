//go:build windows

package retroarch

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func freeSpace(path string) (int64, error) {
	for {
		if _, err := os.Stat(path); err == nil {
			break
		}
		parent := filepath.Dir(path)
		if parent == path {
			return -1, fmt.Errorf("cannot inspect RetroArch installation volume")
		}
		path = parent
	}
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return -1, err
	}
	var available, total, free uint64
	if err = windows.GetDiskFreeSpaceEx(ptr, &available, &total, &free); err != nil {
		return -1, err
	}
	return int64(available), nil
}

func registryCandidates() []string {
	paths := []string{}
	for _, hive := range []registry.Key{registry.CURRENT_USER, registry.LOCAL_MACHINE} {
		for _, view := range []uint32{registry.WOW64_64KEY, registry.WOW64_32KEY} {
			key, err := registry.OpenKey(hive, `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\retroarch.exe`, registry.QUERY_VALUE|view)
			if err == nil {
				value, _, e := key.GetStringValue("")
				key.Close()
				if e == nil {
					paths = append(paths, strings.Trim(value, `"`))
				}
			}
		}
	}
	return paths
}
func installLock(root string) (func(), error) {
	runtime.LockOSThread()
	name, _ := windows.UTF16PtrFromString(fmt.Sprintf(`Local\GemuRetroArch-%x`, sha256.Sum256([]byte(strings.ToLower(root)))))
	h, err := windows.CreateMutex(nil, false, name)
	if err != nil {
		runtime.UnlockOSThread()
		return nil, err
	}
	state, err := windows.WaitForSingleObject(h, 0)
	if err != nil || (state != windows.WAIT_OBJECT_0 && state != windows.WAIT_ABANDONED) {
		windows.CloseHandle(h)
		runtime.UnlockOSThread()
		return nil, fmt.Errorf("another RetroArch installation is already running")
	}
	return func() { windows.ReleaseMutex(h); windows.CloseHandle(h); runtime.UnlockOSThread() }, nil
}
