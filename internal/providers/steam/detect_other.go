//go:build !windows

package steam

import "fmt"

func DetectRoot() (string, error) {
	return "", fmt.Errorf("la detección automática de Steam requiere Windows; indica una carpeta")
}
