//go:build windows && amd64

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"kyberix.com/gemu/internal/overlay"
)

func main() {
	exclusive := flag.Bool("exclusive", false, "test fullscreen exclusive at the current display resolution")
	flag.Parse()
	executable, _ := os.Executable()
	dir := filepath.Dir(executable)
	report := filepath.Join(dir, "overlay-report.json")
	mode := "--windowed"
	if *exclusive {
		mode = "--exclusive"
	}
	command := exec.Command(filepath.Join(dir, "overlay-smoke.exe"), mode, report)
	if err := command.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	finished := make(chan error, 1)
	go func() { finished <- command.Wait() }()
	var client *overlay.Client
	start := time.Now()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-finished:
			if client != nil {
				client.Close()
			}
			if data, e := os.ReadFile(report); e == nil {
				fmt.Print(string(data))
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		case <-ticker.C:
			if client == nil && overlay.HasD3D11(uint32(command.Process.Pid)) {
				var err error
				client, err = overlay.Attach(uint32(command.Process.Pid), filepath.Join(dir, "gemu-overlay.dll"), "Gemu · DirectX 11 test", "es")
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
			}
			if client != nil {
				elapsed := time.Since(start).Seconds()
				client.Update("es", "running", elapsed > 3 && elapsed < 6 || elapsed > 8 && elapsed < 10)
			}
		}
	}
}
