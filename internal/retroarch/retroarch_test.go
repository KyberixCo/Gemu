package retroarch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/library"
	"kyberix.com/gemu/internal/profiles"
	"kyberix.com/gemu/internal/storage"
)

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}
func fixture(t *testing.T) (*Service, string, Config, *library.Service) {
	t.Helper()
	if !supported() {
		t.Skip("Windows x64 feature")
	}
	s := New(t.TempDir())
	p, err := profiles.New(s.dir).Create("Player")
	if err != nil {
		t.Fatal(err)
	}
	s.checkPE = func(path string, _ bool) error {
		if !filepath.IsAbs(path) {
			return fmt.Errorf("absolute path required")
		}
		stat, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !stat.Mode().IsRegular() {
			return fmt.Errorf("not a file")
		}
		return nil
	}
	config := s.defaultConfig(p.ID)
	config.Executable = filepath.Join(t.TempDir(), "Existing RetroArch", "retroarch.exe")
	writeFile(t, config.Executable, []byte("existing executable"))
	writeFile(t, filepath.Join(filepath.Dir(config.Executable), "retroarch.cfg"), []byte("original user configuration"))
	roms := t.TempDir()
	config.Collections = []Collection{{SystemID: "nes", Directory: roms, Recursive: true}}
	writeFile(t, filepath.Join(s.managedDir(), "cores", "fceumm_libretro.dll"), []byte("test core"))
	service := library.New(filepath.Join(s.dir, "profiles", p.ID), nil).WithPlayerStorage(s.dir, p.ID)
	return s, p.ID, config, service
}

func TestScanPublicationPersistenceAndLaunchArguments(t *testing.T) {
	s, id, config, service := fixture(t)
	ctx := context.Background()
	writeFile(t, filepath.Join(config.Collections[0].Directory, "Game one.nes"), []byte("ROM"))
	writeFile(t, filepath.Join(config.Collections[0].Directory, "subfolder", "Game one.nes"), []byte("ROM"))
	preview, err := s.Preview(ctx, id, config)
	if err != nil || len(preview.Games) != 2 || !preview.CanPublish {
		t.Fatal(preview, err)
	}
	if _, err = s.Commit(ctx, "00112233445566778899aabbccddeeff", preview.Token, service); err == nil {
		t.Fatal("another player used preview")
	}
	snapshot, err := s.Commit(ctx, id, preview.Token, service)
	if err != nil || len(snapshot.Games) != 2 {
		t.Fatal(snapshot, err)
	}
	if _, err = s.Commit(ctx, id, preview.Token, service); err == nil {
		t.Fatal("preview reused")
	}
	loaded, err := s.Load(id)
	if err != nil || loaded.Executable != config.Executable {
		t.Fatal(loaded, err)
	}
	paths := map[string]bool{}
	for _, entry := range snapshot.Games {
		plan, err := s.BuildLaunch(id, entry.Game.ID, entry.Game.Installations[0])
		if err != nil {
			t.Fatal(err)
		}
		if plan.Executable != config.Executable || len(plan.Args) != 8 || plan.Args[0] != "--config" || plan.Args[5] != "-L" || plan.Args[7] != entry.Game.Installations[0].Path {
			t.Fatal(plan)
		}
		data, err := os.ReadFile(plan.Args[3])
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		for _, want := range []string{filepath.ToSlash(s.profileDir(id)), "input_menu_toggle_gamepad_combo = 2", "config_save_on_exit = false", "savestate_auto_load = false"} {
			if !strings.Contains(text, want) {
				t.Fatalf("missing %s in %s", want, text)
			}
		}
		for _, line := range strings.Split(text, "\n") {
			if strings.HasPrefix(line, "savefile_directory") {
				paths[line] = true
			}
		}
	}
	if len(paths) != 2 {
		t.Fatal("same-named games share a save directory")
	}
	data, _ := os.ReadFile(filepath.Join(filepath.Dir(config.Executable), "retroarch.cfg"))
	if string(data) != "original user configuration" {
		t.Fatal("changed external config")
	}
	if _, err = os.Stat(filepath.Join(filepath.Dir(config.Executable), "assets")); !os.IsNotExist(err) {
		t.Fatal("created files in an external installation")
	}
	page, err := service.Search(library.SearchRequest{Query: "game"})
	if err != nil || page.Total != 2 {
		t.Fatal("FTS import", page, err)
	}
	bad := snapshot.Games[0].Game.Installations[0]
	bad.Path = filepath.Join(t.TempDir(), "outside.nes")
	writeFile(t, bad.Path, []byte("ROM"))
	if _, err = s.BuildLaunch(id, "forged", bad); err == nil {
		t.Fatal("launched outside collection")
	}
	other, err := profiles.New(s.dir).Create("Other")
	if err != nil {
		t.Fatal(err)
	}
	otherConfig, err := s.Load(other.ID)
	if err != nil || otherConfig.Executable != "" {
		t.Fatal("config leaked", otherConfig, err)
	}
	status, err := s.Detect(id, config.Executable)
	if err != nil || len(status.Installations) < 1 || status.Installations[0].Managed {
		t.Fatal(status, err)
	}
}

func TestChangedCollectionAndMissingCoreDoNotPublish(t *testing.T) {
	s, id, config, service := fixture(t)
	ctx := context.Background()
	file := filepath.Join(config.Collections[0].Directory, "First.nes")
	writeFile(t, file, []byte("ROM"))
	preview, err := s.Preview(ctx, id, config)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(config.Collections[0].Directory, "Added.nes"), []byte("ROM"))
	if _, err = s.Commit(ctx, id, preview.Token, service); err == nil {
		t.Fatal("published a game that was never reviewed")
	}
	snapshot, err := service.Load()
	if err != nil || len(snapshot.Games) != 0 {
		t.Fatal(snapshot, err)
	}
	config.Collections[0].SystemID = "snes"
	preview, err = s.Preview(ctx, id, config)
	if err != nil || preview.CanPublish {
		t.Fatal("missing core marked ready", preview, err)
	}
	if _, err = s.Commit(ctx, id, preview.Token, service); err == nil {
		t.Fatal("published without core")
	}
	config.Collections[0].SystemID = "nes"
	config.Collections[0].Recursive = false
	writeFile(t, filepath.Join(config.Collections[0].Directory, "subfolder", "Hidden.nes"), []byte("ROM"))
	preview, err = s.Preview(ctx, id, config)
	if err != nil || len(preview.Games) != 2 {
		t.Fatal("non-recursive scan", preview, err)
	}
	s.Discard(id, preview.Token)
	if _, err = s.Commit(ctx, id, preview.Token, service); err == nil {
		t.Fatal("discarded preview published")
	}
}

func TestDiscReferencesAndPlaylists(t *testing.T) {
	s, id, config, _ := fixture(t)
	root := config.Collections[0].Directory
	config.Collections[0].SystemID = "psx"
	writeFile(t, filepath.Join(s.managedDir(), "cores", "mednafen_psx_hw_libretro.dll"), []byte("test core"))
	for _, name := range []string{"Disc 1", "Disc 2"} {
		writeFile(t, filepath.Join(root, name+".bin"), []byte("TRACK"))
		writeFile(t, filepath.Join(root, name+".cue"), []byte("FILE \""+name+".bin\" BINARY\n TRACK 01 MODE2/2352\n"))
	}
	writeFile(t, filepath.Join(root, "My game.m3u"), []byte("#EXTM3U\nDisc 1.cue\nDisc 2.cue\n"))
	writeFile(t, filepath.Join(root, "Broken.cue"), []byte("FILE \"missing.bin\" BINARY\n"))
	writeFile(t, filepath.Join(root, "Escape.m3u"), []byte("../outside.chd\n"))
	preview, err := s.Preview(context.Background(), id, config)
	if err != nil || len(preview.Games) != 1 || preview.Games[0].Game.Title != "My game" || len(preview.Diagnostics) < 2 {
		t.Fatal(preview, err)
	}
	if len(preview.Checks[0].Firmware) != 3 || preview.Checks[0].Firmware[0].State != "missing" {
		t.Fatal("missing BIOS claimed valid", preview.Checks)
	}
	writeFile(t, filepath.Join(config.SystemDirectory, "scph5500.bin"), []byte("invalid BIOS"))
	preview, err = s.Preview(context.Background(), id, config)
	if err != nil || preview.CanPublish || preview.Checks[0].Firmware[0].State != "invalid" {
		t.Fatal("wrong BIOS accepted", preview, err)
	}
}

func TestRescanPreservesPublishedArtworkAndOtherSources(t *testing.T) {
	s, id, config, service := fixture(t)
	writeFile(t, filepath.Join(config.Collections[0].Directory, "One.nes"), []byte("ROM"))
	preview, err := s.Preview(context.Background(), id, config)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := s.Commit(context.Background(), id, preview.Token, service)
	if err != nil {
		t.Fatal(err)
	}
	metadata := snapshot.Games[0].Metadata
	metadata.Icon = "keep-icon"
	metadata.ManualArt = []string{"icon"}
	data, _ := json.Marshal(metadata)
	db, err := storage.Open(s.dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("UPDATE games SET metadata=? WHERE scope=?", data, id); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if _, err = service.PublishLocal(library.SourceConfig{ID: "steam", Name: "Steam"}, []library.Entry{{SourceID: "steam", Game: domain.Game{ID: "steam:42", Title: "Other source"}}}, nil); err != nil {
		t.Fatal(err)
	}
	preview, err = s.Preview(context.Background(), id, config)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err = s.Commit(context.Background(), id, preview.Token, service)
	if err != nil || len(snapshot.Games) != 2 {
		t.Fatal(snapshot, err)
	}
	found := false
	for _, entry := range snapshot.Games {
		if entry.SourceID == "retroarch" {
			found = true
			if entry.Metadata.Icon != "keep-icon" {
				t.Fatal("manual art lost")
			}
		}
	}
	if !found {
		t.Fatal("missing RetroArch game")
	}
}

func TestConfigurationChangedAfterPreviewDoesNotOverwriteNewerSetup(t *testing.T) {
	s, id, config, service := fixture(t)
	ctx := context.Background()
	first, err := s.Preview(ctx, id, config)
	if err != nil {
		t.Fatal(err)
	}
	older, err := s.Preview(ctx, id, config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Commit(ctx, id, first.Token, service); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Commit(ctx, id, older.Token, service); err == nil {
		t.Fatal("overwrote a configuration confirmed after preview")
	}
}

func TestWindowsExecutableValidation(t *testing.T) {
	if !supported() {
		t.Skip("Windows PE fixture")
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err = validatePE(executable, false); err != nil {
		t.Fatal(err)
	}
	if err = validatePE(executable, true); err == nil {
		t.Fatal("accepted EXE as core DLL")
	}
	path := filepath.Join(t.TempDir(), "retroarch.exe")
	writeFile(t, path, []byte("not an executable"))
	if err = validatePE(path, false); err == nil {
		t.Fatal("accepted non-PE executable")
	}
}

func TestLaunchKeepsExplicitRetroArchUserConfiguration(t *testing.T) {
	s, id, config, service := fixture(t)
	writeFile(t, filepath.Join(config.Collections[0].Directory, "Game.nes"), []byte("ROM"))
	preview, err := s.Preview(context.Background(), id, config)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := s.Commit(context.Background(), id, preview.Token, service)
	if err != nil {
		t.Fatal(err)
	}
	entry := snapshot.Games[0]
	first, err := s.BuildLaunch(id, entry.Game.ID, entry.Game.Installations[0])
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, first.Args[1], []byte("input_player1_a_btn = 7\n"))
	next, err := s.BuildLaunch(id, entry.Game.ID, entry.Game.Installations[0])
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(next.Args[1])
	if err != nil || string(data) != "input_player1_a_btn = 7\n" || first.Args[3] == next.Args[3] {
		t.Fatal("did not preserve base config and fresh launch overrides", err)
	}
}
