package retroarch

import (
	"archive/zip"
	"bufio"
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/library"
)

func hasExtension(path string, extensions []string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, item := range extensions {
		if ext == item {
			return true
		}
	}
	return false
}
func contentPath(root, path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	if !within(root, resolved) {
		return "", fmt.Errorf("RetroArch content references a file outside its collection")
	}
	stat, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !stat.Mode().IsRegular() {
		return "", fmt.Errorf("RetroArch content is not a regular file")
	}
	return resolved, nil
}

var cueFile = regexp.MustCompile(`(?im)^\s*FILE\s+(?:"([^"]+)"|(\S+))\s+\S+\s*$`)

func validateContent(root, path string, sys System) ([]string, error) {
	if _, err := contentPath(root, path); err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".zip" {
		r, err := zip.OpenReader(path)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		if len(r.File) > 1000 {
			return nil, fmt.Errorf("ROM archive contains too many files")
		}
		matches := 0
		for _, file := range r.File {
			if _, err = archiveName(file.Name); err != nil {
				return nil, err
			}
			if !file.FileInfo().IsDir() && hasExtension(file.Name, sys.Extensions) {
				if file.UncompressedSize64 > 512*1024*1024 {
					return nil, fmt.Errorf("ROM archive entry is too large")
				}
				matches++
			}
		}
		if matches != 1 {
			return nil, fmt.Errorf("ROM archive must contain exactly one compatible game")
		}
		return nil, nil
	}
	if ext != ".cue" && ext != ".m3u" {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := readSmall(f)
	if err != nil {
		return nil, err
	}
	refs := []string{}
	if ext == ".cue" {
		for _, match := range cueFile.FindAllStringSubmatch(string(data), -1) {
			name := match[1]
			if name == "" {
				name = match[2]
			}
			refs = append(refs, name)
		}
	} else {
		scanner := bufio.NewScanner(strings.NewReader(strings.TrimPrefix(string(data), "\ufeff")))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				refs = append(refs, line)
			}
		}
		if err = scanner.Err(); err != nil {
			return nil, err
		}
	}
	if len(refs) == 0 || len(refs) > 100 {
		return nil, fmt.Errorf("disc descriptor is empty or has too many files")
	}
	resolved := []string{}
	for _, ref := range refs {
		if filepath.IsAbs(ref) || strings.Contains(ref, ":") {
			return nil, fmt.Errorf("disc descriptor requires relative file names")
		}
		full, err := contentPath(root, filepath.Join(filepath.Dir(path), filepath.FromSlash(strings.ReplaceAll(ref, `\`, "/"))))
		if err != nil {
			return nil, err
		}
		if ext == ".m3u" {
			if !hasExtension(full, []string{".cue", ".chd", ".pbp"}) {
				return nil, fmt.Errorf("playlist must reference CUE, CHD or PBP discs")
			}
			if strings.EqualFold(filepath.Ext(full), ".cue") {
				if _, err = validateContent(root, full, sys); err != nil {
					return nil, err
				}
			}
		}
		resolved = append(resolved, full)
	}
	return resolved, nil
}

func (s *Service) inspectFirmware(config Config, sys System) ([]FirmwareStatus, []string) {
	items := []FirmwareStatus{}
	diagnostics := []string{}
	for _, firmware := range sys.Firmware {
		item := FirmwareStatus{Name: firmware.Name, Region: firmware.Region, State: "missing"}
		path := filepath.Join(config.SystemDirectory, firmware.Name)
		if f, err := os.Open(path); err == nil {
			stat, e := f.Stat()
			if e == nil && stat.Mode().IsRegular() && stat.Size() <= 4*1024*1024 {
				hash := md5.New()
				if _, e = io.Copy(hash, f); e == nil && hex.EncodeToString(hash.Sum(nil)) == firmware.MD5 {
					item.State = "verified"
				} else {
					item.State = "invalid"
				}
			} else {
				item.State = "invalid"
			}
			f.Close()
		}
		items = append(items, item)
		if item.State == "invalid" {
			diagnostics = append(diagnostics, "BIOS checksum does not match: "+firmware.Name)
		}
	}
	return items, diagnostics
}
func (s *Service) scan(ctx context.Context, config Config) (Preview, error) {
	preview := Preview{Games: []library.Entry{}, Checks: []Check{}, Diagnostics: []string{}, CanPublish: true}
	seenGames := map[string]bool{}
	for _, collection := range config.Collections {
		sys, _ := system(collection.SystemID)
		core, err := s.corePath(config, sys.Core)
		firmware, diagnostics := s.inspectFirmware(config, sys)
		preview.Checks = append(preview.Checks, Check{SystemID: sys.ID, CorePath: core, CoreReady: err == nil, Firmware: firmware})
		preview.Diagnostics = append(preview.Diagnostics, diagnostics...)
		if len(diagnostics) > 0 {
			preview.CanPublish = false
		}
		if err != nil {
			preview.CanPublish = false
			preview.Diagnostics = append(preview.Diagnostics, err.Error())
		}
		paths := []string{}
		scanned := 0
		err = filepath.WalkDir(collection.Directory, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				preview.Diagnostics = append(preview.Diagnostics, path+": "+err.Error())
				return nil
			}
			if err = ctx.Err(); err != nil {
				return err
			}
			scanned++
			if scanned > 100000 {
				return fmt.Errorf("RetroArch collection exceeds the scan limit")
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return nil
			}
			if entry.IsDir() {
				if path != collection.Directory && !collection.Recursive {
					return fs.SkipDir
				}
				return nil
			}
			if hasExtension(path, sys.Extensions) || sys.ID != "psx" && strings.EqualFold(filepath.Ext(path), ".zip") {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			return preview, err
		}
		valid := []string{}
		referenced := map[string]bool{}
		for _, path := range paths {
			refs, err := validateContent(collection.Directory, path, sys)
			if err != nil {
				preview.Diagnostics = append(preview.Diagnostics, filepath.Base(path)+": "+err.Error())
				continue
			}
			valid = append(valid, path)
			if strings.EqualFold(filepath.Ext(path), ".m3u") {
				for _, ref := range refs {
					referenced[strings.ToLower(ref)] = true
				}
			}
		}
		for _, path := range valid {
			if referenced[strings.ToLower(path)] {
				continue
			}
			id := fmt.Sprintf("retroarch:%s:%x", sys.ID, sha256.Sum256([]byte(strings.ToLower(filepath.Clean(path)))))
			if seenGames[id] {
				continue
			}
			seenGames[id] = true
			title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			title = strings.ReplaceAll(title, "_", " ")
			preview.Games = append(preview.Games, library.Entry{Game: domain.Game{ID: id, Title: title, Installations: []domain.Installation{{Provider: "retroarch", ExternalID: sys.ID, Path: path, Installed: true}}}, SourceID: "retroarch", Metadata: library.Metadata{Platform: sys.Name, Genres: []string{}, Origin: "local", Status: "pending"}, Player: library.PlayerState{Profile: library.Profile{}, Status: "pending", Achievements: []library.Achievement{}, Message: ""}})
			if len(preview.Games) > 20000 {
				return preview, fmt.Errorf("RetroArch library exceeds 20000 games")
			}
		}
	}
	sort.Slice(preview.Games, func(i, j int) bool { return preview.Games[i].Game.Title < preview.Games[j].Game.Title })
	if len(preview.Diagnostics) > 200 {
		preview.Diagnostics = append(preview.Diagnostics[:200], "Additional diagnostics omitted")
	}
	return preview, nil
}
func (s *Service) Preview(ctx context.Context, id string, config Config) (Preview, error) {
	if err := safeProfileID(id); err != nil {
		return Preview{}, err
	}
	config, err := s.normalize(config, id)
	if err != nil {
		return Preview{}, err
	}
	baseline, err := s.setupData(id)
	if err != nil {
		return Preview{}, err
	}
	preview, err := s.scan(ctx, config)
	if err != nil {
		return Preview{}, err
	}
	preview.Token = rand.Text()
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, d := range s.drafts {
		if time.Now().After(d.expires) {
			delete(s.drafts, token)
		}
	}
	if len(s.drafts) >= 8 {
		return Preview{}, fmt.Errorf("too many RetroArch previews; close previous assistants")
	}
	s.drafts[preview.Token] = draft{owner: id, baseline: baseline, config: config, preview: preview, expires: time.Now().Add(15 * time.Minute)}
	return preview, nil
}
func (s *Service) Discard(id, token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if d, ok := s.drafts[token]; ok && d.owner == id {
		delete(s.drafts, token)
	}
}
func (s *Service) Commit(ctx context.Context, id, token string, service *library.Service) (library.Snapshot, error) {
	s.mu.Lock()
	d, ok := s.drafts[token]
	if !ok || d.owner != id || time.Now().After(d.expires) {
		s.mu.Unlock()
		return library.Snapshot{}, fmt.Errorf("RetroArch preview expired; scan again")
	}
	if !d.preview.CanPublish {
		s.mu.Unlock()
		return library.Snapshot{}, fmt.Errorf("resolve missing cores or invalid BIOS before publishing")
	}
	delete(s.drafts, token)
	s.mu.Unlock()
	config, err := s.normalize(d.config, id)
	if err != nil {
		return library.Snapshot{}, err
	}
	// Check paths and requirements again, but never silently publish games that
	// were added after review. A changed scan requires a fresh review.
	current, err := s.scan(ctx, config)
	if err != nil {
		return library.Snapshot{}, err
	}
	if !current.CanPublish || len(current.Games) != len(d.preview.Games) {
		return library.Snapshot{}, fmt.Errorf("RetroArch collection changed; review it again")
	}
	for i, entry := range current.Games {
		if entry.Game.ID != d.preview.Games[i].Game.ID {
			return library.Snapshot{}, fmt.Errorf("RetroArch collection changed; review it again")
		}
	}
	data, err := json.Marshal(config)
	if err != nil {
		return library.Snapshot{}, err
	}
	snapshot, err := service.PublishLocal(library.SourceConfig{ID: "retroarch", Name: "RetroArch", Root: config.Executable}, current.Games, func(tx *sql.Tx) error {
		var previous string
		err := tx.QueryRow("SELECT data FROM retroarch_setups WHERE scope=?", id).Scan(&previous)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if previous != d.baseline {
			return fmt.Errorf("RetroArch configuration changed; review it again")
		}
		_, err = tx.Exec("INSERT INTO retroarch_setups(scope,data) VALUES (?,?) ON CONFLICT(scope) DO UPDATE SET data=excluded.data", id, data)
		return err
	})
	return snapshot, err
}
