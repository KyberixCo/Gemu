// library prepares a source import without publishing it unless -commit is given.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"kyberix.com/gemu/internal/library"
	"kyberix.com/gemu/internal/providers/steam"
	"os"
	"path/filepath"
	"time"
)

func main() {
	source := flag.String("source", "steam", "source adapter")
	root := flag.String("root", "", "installation/library root")
	profile := flag.String("profile", "", "player profile ID (optional)")
	dir := flag.String("data-dir", "", "cache and catalog directory")
	commit := flag.Bool("commit", false, "publish the prepared import")
	flag.Parse()
	if *dir == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			fail(err)
		}
		*dir = filepath.Join(base, "Gemu")
	}
	service := library.New(*dir, map[string]library.Source{"steam": steam.NewProvider()})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	preview, err := service.Preview(ctx, *source, *root)
	if err != nil {
		fail(err)
	}
	if *profile == "" && len(preview.Profiles) == 1 {
		*profile = preview.Profiles[0].ID
	}
	draft, err := service.Prepare(ctx, library.Request{SourceID: *source, Root: preview.Root, ProfileID: *profile, APIKey: os.Getenv("GEMU_STEAM_API_KEY")})
	if err != nil {
		fail(err)
	}
	if *commit {
		if _, err = service.Commit(draft.Token); err != nil {
			fail(err)
		}
	}
	type summary struct {
		Title        string
		Metadata     string
		Art          bool
		Profile      bool
		Achievements string
	}
	games := []summary{}
	for _, entry := range draft.Games {
		games = append(games, summary{entry.Game.Title, entry.Metadata.Status, entry.Metadata.Cover != "" || entry.Metadata.Background != "", entry.Player.Profile.ID != "", entry.Player.Status})
	}
	if err = json.NewEncoder(os.Stdout).Encode(struct {
		Games       []summary
		Diagnostics []string
		Published   bool
	}{games, draft.Diagnostics, *commit}); err != nil {
		fail(err)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
