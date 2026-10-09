// discover inspects a local Steam directory without starting any processes.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"kyberix.com/gemu/internal/providers/steam"
	"os"
)

func main() {
	root := flag.String("steam-root", "", "Steam installation directory (required)")
	flag.Parse()
	if *root == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/discover -steam-root <directory>")
		os.Exit(2)
	}
	if info, err := os.Stat(*root); err != nil || !info.IsDir() {
		fmt.Fprintln(os.Stderr, "Steam directory is not accessible")
		os.Exit(1)
	}
	games, diagnostics := steam.Discover(*root)
	for _, err := range diagnostics {
		fmt.Fprintln(os.Stderr, err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(games); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
