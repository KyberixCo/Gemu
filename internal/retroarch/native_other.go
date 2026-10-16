//go:build !windows

package retroarch

import "fmt"

func freeSpace(string) (int64, error) { return -1, nil }

func registryCandidates() []string { return nil }
func installLock(string) (func(), error) {
	return nil, fmt.Errorf("RetroArch setup is available on Windows x64")
}
