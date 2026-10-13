//go:build windows

package credentials

import (
	"fmt"
	"golang.org/x/sys/windows"
	"runtime"
	"unsafe"
)

func crypt(data, entropy []byte, decrypt bool) ([]byte, error) {
	if len(data) == 0 || len(data) > maxSize {
		return nil, fmt.Errorf("datos inválidos")
	}
	input := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	scope := windows.DataBlob{Size: uint32(len(entropy)), Data: &entropy[0]}
	output := windows.DataBlob{}
	var err error
	if decrypt {
		err = windows.CryptUnprotectData(&input, nil, &scope, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &output)
	} else {
		err = windows.CryptProtectData(&input, nil, &scope, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &output)
	}
	runtime.KeepAlive(data)
	runtime.KeepAlive(entropy)
	if err != nil {
		return nil, fmt.Errorf("protección de Windows no disponible")
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(output.Data)))
	if output.Data == nil || output.Size > maxSize {
		return nil, fmt.Errorf("datos inválidos")
	}
	native := unsafe.Slice(output.Data, int(output.Size))
	result := append([]byte{}, native...)
	clear(native)
	return result, nil
}
func protect(data, entropy []byte) ([]byte, error)   { return crypt(data, entropy, false) }
func unprotect(data, entropy []byte) ([]byte, error) { return crypt(data, entropy, true) }
