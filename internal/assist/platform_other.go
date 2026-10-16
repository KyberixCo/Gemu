//go:build !windows

package assist

import (
	"context"
	"fmt"
)

func Connect(context.Context) error  { return nil }
func Healthy() bool                  { return false }
func CurrentMode() Mode              { return Off }
func MenuShortcut(bool) (bool, bool) { return false, false }
func Open(Mode) error                { return fmt.Errorf("La asistencia de escritorio requiere Windows") }
func Run() error                     { return fmt.Errorf("La asistencia de escritorio requiere Windows") }
