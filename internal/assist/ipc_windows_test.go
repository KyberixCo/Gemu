//go:build windows

package assist

import (
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestNativeInputLayout(t *testing.T) {
	want, offset := uintptr(40), uintptr(8)
	if unsafe.Sizeof(uintptr(0)) == 4 {
		want = 28
		offset = 4
	}
	if unsafe.Sizeof(nativeInput{}) != want || unsafe.Offsetof(nativeInput{}.Mouse) != offset {
		t.Fatalf("INPUT layout: size=%d union=%d", unsafe.Sizeof(nativeInput{}), unsafe.Offsetof(nativeInput{}.Mouse))
	}
	if unsafe.Sizeof(padState{}) != 16 {
		t.Fatal("XINPUT_STATE layout changed")
	}
	input := keyInput(0, 0x00f1, 4)
	key := (*keyboardInput)(unsafe.Pointer(&input.Mouse))
	if input.Kind != 1 || key.Scan != 0xf1 || key.Flags != 4 {
		t.Fatal("Unicode keyboard event incorrectly encoded")
	}
}

func TestLocalPipeRoundTripAndSingleInstance(t *testing.T) {
	pipe, err := createPipe()
	if err != nil {
		t.Skipf("companion already running or interactive pipe unavailable: %v", err)
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	var count atomic.Int32
	go func() {
		defer close(done)
		servePipe(pipe, stop, func(request Request) Status {
			count.Add(1)
			return Status{Version: protocolVersion, Mode: Pointer, Menu: uint64(request.HostPID)}
		})
	}()
	defer func() {
		close(stop)
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("pipe server did not stop")
		}
	}()
	if second, err := createPipe(); err == nil {
		windows.CloseHandle(second)
		t.Fatal("second companion acquired the same pipe")
	}
	for i := 0; i < 3; i++ {
		status, err := exchange(Request{Action: "status", HostPID: 42})
		if err != nil {
			t.Fatal(err)
		}
		if status.Mode != Pointer || status.Menu != 42 {
			t.Fatalf("unexpected response: %+v", status)
		}
	}
	if count.Load() != 3 {
		t.Fatal("request was lost or repeated")
	}
	// A client that connects but never sends must not prevent future clients
	// from reaching the desktop companion.
	name, _, err := pipeIdentity()
	if err != nil {
		t.Fatal(err)
	}
	path, _ := windows.UTF16PtrFromString(name)
	stalled, err := windows.CreateFile(path, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OVERLAPPED, 0)
	if err == windows.ERROR_PIPE_BUSY {
		waitNamedPipe.Call(uintptr(unsafe.Pointer(path)), 100)
		stalled, err = windows.CreateFile(path, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OVERLAPPED, 0)
	}
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(stalled)
	time.Sleep(350 * time.Millisecond)
	if _, err := exchange(Request{Action: "status"}); err != nil {
		t.Fatalf("stalled client blocked the server: %v", err)
	}
}
