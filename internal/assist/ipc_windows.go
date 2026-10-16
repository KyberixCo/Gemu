//go:build windows

package assist

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var waitNamedPipe = windows.NewLazySystemDLL("kernel32.dll").NewProc("WaitNamedPipeW")

func pipeIdentity() (string, string, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return "", "", err
	}
	sid := user.User.Sid.String()
	var session uint32
	if err := windows.ProcessIdToSessionId(uint32(os.Getpid()), &session); err != nil {
		return "", "", err
	}
	return fmt.Sprintf(`\\.\pipe\GemuInput-%s-%d-v1`, sid, session), sid, nil
}

// Named pipes stay inside this user's interactive session. The ACL explicitly
// restricts clients to this user; remote connections are rejected by Windows.
func createPipe() (windows.Handle, error) {
	name, sid, err := pipeIdentity()
	if err != nil {
		return 0, err
	}
	sd, err := windows.SecurityDescriptorFromString("D:P(A;;GA;;;" + sid + ")")
	if err != nil {
		return 0, err
	}
	security := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	path, _ := windows.UTF16PtrFromString(name)
	return windows.CreateNamedPipe(path, windows.PIPE_ACCESS_DUPLEX|windows.FILE_FLAG_FIRST_PIPE_INSTANCE|windows.FILE_FLAG_OVERLAPPED,
		windows.PIPE_TYPE_MESSAGE|windows.PIPE_READMODE_MESSAGE|windows.PIPE_REJECT_REMOTE_CLIENTS,
		1, 4096, 4096, 250, &security)
}

func exchange(request Request) (Status, error) {
	var status Status
	name, _, err := pipeIdentity()
	if err != nil {
		return status, err
	}
	path, _ := windows.UTF16PtrFromString(name)
	request.Version = protocolVersion
	encoded, err := json.Marshal(request)
	if err != nil {
		return status, err
	}
	handle, err := windows.CreateFile(path, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OVERLAPPED, 0)
	if err == windows.ERROR_PIPE_BUSY {
		waitNamedPipe.Call(uintptr(unsafe.Pointer(path)), 100)
		handle, err = windows.CreateFile(path, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OVERLAPPED, 0)
	}
	if err != nil {
		return status, err
	}
	defer windows.CloseHandle(handle)
	if _, err = pipeIO(handle, encoded, true, 100); err != nil {
		return status, err
	}
	buffer := make([]byte, 4096)
	read, err := pipeIO(handle, buffer, false, 400)
	if err != nil {
		return status, err
	}
	// An explicit acknowledgement lets the server recycle the single instance
	// promptly, without flushing synchronously or discarding an unread response.
	if _, err = pipeIO(handle, []byte{1}, true, 100); err != nil {
		return status, err
	}
	if err := json.Unmarshal(buffer[:read], &status); err != nil {
		return status, err
	}
	if status.Version != protocolVersion {
		return status, fmt.Errorf("incompatible input companion")
	}
	if status.Error != "" {
		return status, errors.New(status.Error)
	}
	return status, nil
}

func servePipe(handle windows.Handle, stop <-chan struct{}, dispatch func(Request) Status) {
	defer windows.CloseHandle(handle)
	buffer := make([]byte, 4096)
	for {
		select {
		case <-stop:
			return
		default:
		}
		event, err := windows.CreateEvent(nil, 1, 0, nil)
		if err != nil {
			return
		}
		op := windows.Overlapped{HEvent: event}
		err = windows.ConnectNamedPipe(handle, &op)
		if err == windows.ERROR_PIPE_CONNECTED {
			err = nil
		}
		_, err = finishIO(handle, &op, err, 100)
		windows.CloseHandle(event)
		if err != nil {
			windows.DisconnectNamedPipe(handle)
			continue
		}
		read, err := pipeIO(handle, buffer, false, 300)
		if err == nil && read > 0 {
			request := Request{}
			status := Status{Version: protocolVersion, Error: "invalid input companion request"}
			if json.Unmarshal(buffer[:read], &request) == nil && request.Version == protocolVersion {
				status = dispatch(request)
			}
			encoded, _ := json.Marshal(status)
			if _, err := pipeIO(handle, encoded, true, 100); err == nil {
				pipeIO(handle, buffer[:1], false, 100)
			}
		}
		windows.DisconnectNamedPipe(handle)
	}
}

func pipeIO(handle windows.Handle, buffer []byte, write bool, timeout uint32) (uint32, error) {
	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(event)
	op := windows.Overlapped{HEvent: event}
	var count uint32
	if write {
		err = windows.WriteFile(handle, buffer, &count, &op)
	} else {
		err = windows.ReadFile(handle, buffer, &count, &op)
	}
	count, err = finishIO(handle, &op, err, timeout)
	runtime.KeepAlive(buffer)
	return count, err
}

func finishIO(handle windows.Handle, op *windows.Overlapped, err error, timeout uint32) (uint32, error) {
	var count uint32
	if err == windows.ERROR_IO_PENDING {
		result, waitErr := windows.WaitForSingleObject(op.HEvent, timeout)
		if waitErr != nil || result != windows.WAIT_OBJECT_0 {
			windows.CancelIoEx(handle, op)
			// Drain cancellation before the OVERLAPPED and its buffer leave scope.
			windows.GetOverlappedResult(handle, op, &count, true)
			return 0, fmt.Errorf("input companion I/O timed out after %s", time.Duration(timeout)*time.Millisecond)
		}
		err = nil
	}
	if err != nil {
		return 0, err
	}
	err = windows.GetOverlappedResult(handle, op, &count, false)
	return count, err
}
