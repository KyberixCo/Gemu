package main

import (
	"embed"
	"kyberix.com/gemu/internal/library"
	"kyberix.com/gemu/internal/profiles"
	"kyberix.com/gemu/internal/providers/steam"
	"kyberix.com/gemu/internal/retroarch"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	dir := os.Getenv("GEMU_DATA_DIR")
	if dir == "" {
		dir = filepath.Join(configDir, "Gemu")
	}
	retro := retroarch.New(dir)
	sources := map[string]library.Source{"steam": steam.NewProvider(), "retroarch": retroarch.NewSource(retro, "")}
	service := library.New(dir, sources)
	app := &App{profiles: profiles.New(dir), sources: sources, dir: dir, retro: retro}
	if err := wails.Run(&options.App{
		Title:      "Gemu",
		Fullscreen: true,
		Width:      1280, Height: 720,
		MinWidth: 960, MinHeight: 540,
		AssetServer:      &assetserver.Options{Assets: assets, Handler: service, Middleware: service.Middleware},
		Bind:             []interface{}{app},
		OnStartup:        app.startup,
		BackgroundColour: &options.RGBA{R: 10, G: 12, B: 18, A: 255},
	}); err != nil {
		log.Fatal(err)
	}
}
