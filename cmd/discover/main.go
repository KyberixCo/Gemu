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
	root := flag.String("steam-root", "", "Steam installation/library directory (empty: detect on Windows)")
	flag.Parse()
	report, err := steam.Inspect(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
