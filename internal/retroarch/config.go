package retroarch

import (
	"database/sql"
	"debug/pe"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"kyberix.com/gemu/internal/storage"
)

func supported() bool { return runtime.GOOS == "windows" && runtime.GOARCH == "amd64" }
func (s *Service) profileDir(id string) string {
	return filepath.Join(s.dir, "profiles", id, "retroarch")
}
func (s *Service) managedDir() string { return filepath.Join(s.dir, "emulators", "retroarch") }
func (s *Service) defaultConfig(id string) Config {
	return Config{SystemDirectory: filepath.Join(s.profileDir(id), "system"), Collections: []Collection{}, Fullscreen: true, Language: "es"}
}
func (s *Service) Load(id string) (Config, error) {
	config := s.defaultConfig(id)
	db, err := storage.Open(s.dir)
	if err != nil {
		return config, err
	}
	defer db.Close()
	var data []byte
	err = db.QueryRow("SELECT data FROM retroarch_setups WHERE scope=?", id).Scan(&data)
	if err == sql.ErrNoRows {
		return config, nil
	}
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

func (s *Service) setupData(id string) (string, error) {
	db, err := storage.Open(s.dir)
	if err != nil {
		return "", err
	}
	defer db.Close()
	var data string
	err = db.QueryRow("SELECT data FROM retroarch_setups WHERE scope=?", id).Scan(&data)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return data, err
}
func validatePE(path string, dll bool) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("RetroArch requires an absolute file path")
	}
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !stat.Mode().IsRegular() {
		return fmt.Errorf("RetroArch file is not a regular file")
	}
	f, err := pe.Open(path)
	if err != nil {
		return fmt.Errorf("invalid Windows executable: %w", err)
	}
	defer f.Close()
	if f.Machine != pe.IMAGE_FILE_MACHINE_AMD64 || (f.Characteristics&pe.IMAGE_FILE_DLL != 0) != dll {
		return fmt.Errorf("RetroArch and its cores must use Windows x64")
	}
	if _, ok := f.OptionalHeader.(*pe.OptionalHeader64); !ok || len(f.Sections) == 0 {
		return fmt.Errorf("invalid Windows executable")
	}
	return nil
}
func (s *Service) validateExecutable(path string) error {
	if !strings.EqualFold(filepath.Base(path), "retroarch.exe") {
		return fmt.Errorf("select retroarch.exe")
	}
	return s.checkPE(path, false)
}
func cleanDirectory(path string) (string, error) {
	if !filepath.IsAbs(path) || strings.ContainsAny(path, "\x00\r\n\"") {
		return "", fmt.Errorf("invalid absolute directory")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	stat, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("path is not a directory")
	}
	return resolved, nil
}
func (s *Service) normalize(config Config, id string) (Config, error) {
	if !supported() {
		return config, fmt.Errorf("RetroArch setup is available on Windows x64")
	}
	exe, err := filepath.EvalSymlinks(config.Executable)
	if err != nil {
		return config, err
	}
	config.Executable = exe
	if err = s.validateExecutable(exe); err != nil {
		return config, err
	}
	if config.Language != "en" && config.Language != "es" {
		config.Language = "es"
	}
	if config.JoypadDriver == "" {
		config.JoypadDriver = "auto"
	}
	if config.JoypadDriver != "auto" && config.JoypadDriver != "xinput" && config.JoypadDriver != "sdl2" && config.JoypadDriver != "dinput" {
		return config, fmt.Errorf("unsupported RetroArch controller driver")
	}
	if config.CoreDirectory != "" {
		config.CoreDirectory, err = cleanDirectory(config.CoreDirectory)
		if err != nil {
			return config, err
		}
	}
	if len(config.Collections) > len(Systems) {
		return config, fmt.Errorf("too many RetroArch collections")
	}
	if config.SystemDirectory == "" {
		config.SystemDirectory = s.defaultConfig(id).SystemDirectory
	}
	// The default BIOS directory may be created at commit; external directories
	// must already exist. Merely previewing never creates user folders.
	if strings.EqualFold(filepath.Clean(config.SystemDirectory), filepath.Clean(s.defaultConfig(id).SystemDirectory)) {
		config.SystemDirectory = s.defaultConfig(id).SystemDirectory
	} else {
		config.SystemDirectory, err = cleanDirectory(config.SystemDirectory)
		if err != nil {
			return config, err
		}
	}
	seen := map[string]bool{}
	for i, c := range config.Collections {
		if _, ok := system(c.SystemID); !ok || seen[c.SystemID] {
			return config, fmt.Errorf("invalid or duplicate RetroArch system")
		}
		seen[c.SystemID] = true
		config.Collections[i].Directory, err = cleanDirectory(c.Directory)
		if err != nil {
			return config, err
		}
	}
	return config, nil
}
func (s *Service) corePath(config Config, core string) (string, error) {
	name := core + "_libretro.dll"
	paths := []string{filepath.Join(s.managedDir(), "cores", name), filepath.Join(filepath.Dir(config.Executable), "cores", name)}
	if config.CoreDirectory != "" {
		paths = append([]string{filepath.Join(config.CoreDirectory, name)}, paths...)
	}
	for _, path := range paths {
		if err := s.checkPE(path, true); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("selected RetroArch core is not installed: %s", core)
}

func (s *Service) CreateFolders(id string, systems []string) (map[string]string, error) {
	if err := safeProfileID(id); err != nil {
		return nil, err
	}
	paths := map[string]string{}
	for _, key := range systems {
		if _, ok := system(key); !ok {
			return nil, fmt.Errorf("unknown RetroArch system")
		}
		paths[key] = filepath.Join(s.profileDir(id), "roms", key)
	}
	for _, path := range paths {
		if err := os.MkdirAll(path, 0700); err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(s.defaultConfig(id).SystemDirectory, 0700); err != nil {
		return nil, err
	}
	return paths, nil
}
func cfgValue(value string) (string, error) {
	if strings.ContainsAny(value, "\x00\r\n\"") {
		return "", fmt.Errorf("invalid RetroArch configuration value")
	}
	return `"` + filepath.ToSlash(value) + `"`, nil
}

// A new immutable config is generated for every launch. RetroArch may modify
// that file, but it never edits the user's existing retroarch.cfg or our DB.
func (s *Service) launchConfig(id string, config Config, gameID, core string) (string, error) {
	base := s.profileDir(id)
	dirs := map[string]string{"savefile_directory": filepath.Join(base, "saves", gameID), "savestate_directory": filepath.Join(base, "states", gameID), "screenshot_directory": filepath.Join(base, "screenshots"), "input_remapping_directory": filepath.Join(base, "remaps"), "core_options_path": filepath.Join(base, "core-options.cfg"), "log_dir": filepath.Join(base, "logs"), "system_directory": config.SystemDirectory, "libretro_directory": filepath.Dir(core), "libretro_info_path": filepath.Join(s.managedDir(), "info"), "joypad_autoconfig_dir": filepath.Join(filepath.Dir(config.Executable), "autoconfig"), "assets_directory": filepath.Join(filepath.Dir(config.Executable), "assets")}
	if _, err := os.Stat(filepath.Join(dirs["libretro_info_path"], strings.TrimSuffix(filepath.Base(core), ".dll")+".info")); err != nil {
		dirs["libretro_info_path"] = filepath.Join(filepath.Dir(config.Executable), "info")
	}
	for _, key := range []string{"savefile_directory", "savestate_directory", "screenshot_directory", "input_remapping_directory", "log_dir"} {
		if err := os.MkdirAll(dirs[key], 0700); err != nil {
			return "", err
		}
	}
	if config.SystemDirectory == s.defaultConfig(id).SystemDirectory {
		if err := os.MkdirAll(config.SystemDirectory, 0700); err != nil {
			return "", err
		}
	}
	values := map[string]string{"video_fullscreen": strconv.FormatBool(config.Fullscreen), "video_windowed_fullscreen": "true", "input_autodetect_enable": "true", "config_save_on_exit": "false", "rewind_enable": strconv.FormatBool(config.Rewind), "savestate_auto_save": "false", "savestate_auto_load": "false", "log_to_file": "true", "log_verbosity": "true", "input_menu_toggle": "f1", "input_exit_emulator": "escape"}
	values["input_menu_toggle_gamepad_combo"] = "2" // RetroArch INPUT_COMBO_L3_R3; independent of Gemu's View+Menu.
	if config.JoypadDriver != "" && config.JoypadDriver != "auto" {
		values["input_joypad_driver"], _ = cfgValue(config.JoypadDriver)
	}
	if config.Language == "es" {
		values["user_language"] = "3"
	} else {
		values["user_language"] = "0"
	}
	for key, path := range dirs {
		value, err := cfgValue(path)
		if err != nil {
			return "", err
		}
		values[key] = value
	}
	keys := []string{}
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var text strings.Builder
	for _, key := range keys {
		text.WriteString(key + " = " + values[key] + "\n")
	}
	if err := os.MkdirAll(filepath.Join(base, "sessions"), 0700); err != nil {
		return "", err
	}
	f, err := os.CreateTemp(filepath.Join(base, "sessions"), "launch-*.cfg")
	if err != nil {
		return "", err
	}
	_, err = f.WriteString(text.String())
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return "", err
	}
	if closeErr != nil {
		return "", closeErr
	}
	return f.Name(), nil
}

func (s *Service) baseConfig(id string) (string, error) {
	path := filepath.Join(s.profileDir(id), "retroarch.cfg")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if os.IsExist(err) {
		return path, nil
	}
	if err != nil {
		return "", err
	}
	_, err = f.WriteString("# RetroArch settings for this Gemu player. Save explicitly from RetroArch.\n")
	if err == nil {
		err = f.Sync()
	}
	closed := f.Close()
	if err != nil {
		return "", err
	}
	if closed != nil {
		return "", closed
	}
	return path, nil
}
