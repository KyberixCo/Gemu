//go:build !windows

package launcher

import "fmt"

type native struct{}

func Native() Platform                       { return native{} }
func FocusHost() error                       { return nil }
func RestoreHost() error                     { return nil }
func SetHostOverlay(bool) error              { return nil }
func (native) Open(string) error             { return fmt.Errorf("El lanzamiento está disponible en Windows") }
func (native) Processes() ([]Process, error) { return nil, nil }
func (native) Focus(Process) error           { return fmt.Errorf("No disponible") }
func (native) Close(Process) error           { return fmt.Errorf("No disponible") }
func (native) Shortcut(bool) bool            { return false }
