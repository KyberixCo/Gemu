package retroarch

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bodgit/sevenzip"
)

const buildbot = "https://buildbot.libretro.com"

var versionPattern = regexp.MustCompile(`^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`)
var releaseLinks = regexp.MustCompile(`href=["'](?:/stable/)?([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})/["']`)

func officialURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" {
		return false
	}
	return u.Host == "buildbot.libretro.com" || u.Host == "raw.githubusercontent.com" && strings.HasPrefix(u.Path, "/libretro/libretro-core-info/master/")
}
func (s *Service) request(ctx context.Context, method, raw string) (*http.Response, error) {
	if !officialURL(raw) {
		return nil, fmt.Errorf("RetroArch download URL is not an official source")
	}
	req, err := http.NewRequestWithContext(ctx, method, raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Gemu/0.1 RetroArch-setup")
	response, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("RetroArch download HTTP %d", response.StatusCode)
	}
	return response, nil
}
func (s *Service) latest(ctx context.Context) (string, error) {
	response, err := s.request(ctx, "GET", buildbot+"/stable/")
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024+1))
	if err != nil {
		return "", err
	}
	if len(data) > 2*1024*1024 {
		return "", fmt.Errorf("RetroArch release list is too large")
	}
	versions := []string{}
	for _, match := range releaseLinks.FindAllStringSubmatch(string(data), -1) {
		versions = append(versions, match[1])
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("official RetroArch stable release could not be resolved")
	}
	sort.Slice(versions, func(i, j int) bool {
		a := strings.Split(versions[i], ".")
		b := strings.Split(versions[j], ".")
		for n := 0; n < 3; n++ {
			x, _ := strconv.Atoi(a[n])
			y, _ := strconv.Atoi(b[n])
			if x != y {
				return x > y
			}
		}
		return false
	})
	return versions[0], nil
}

type countedReader struct {
	ctx    context.Context
	r      io.Reader
	update func(int64)
}

func (r countedReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := r.r.Read(p)
	if n > 0 && r.update != nil {
		r.update(int64(n))
	}
	return n, err
}

func (s *Service) download(ctx context.Context, asset Download, path string, progress func(int64, int64)) (string, error) {
	response, err := s.request(ctx, "GET", asset.URL)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	limit := int64(768 * 1024 * 1024)
	if strings.HasSuffix(asset.Name, ".info") {
		limit = 1024 * 1024
	}
	if strings.HasSuffix(asset.Name, ".dll.zip") {
		limit = 128 * 1024 * 1024
	}
	if response.ContentLength > limit {
		return "", fmt.Errorf("RetroArch download exceeds the size limit")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	var read int64
	r := countedReader{ctx: ctx, r: io.LimitReader(response.Body, limit+1), update: func(n int64) { read += n; progress(read, response.ContentLength) }}
	n, err := io.Copy(io.MultiWriter(f, hash), r)
	if err == nil && n > limit {
		err = fmt.Errorf("RetroArch download exceeds the size limit")
	}
	if err == nil && response.ContentLength >= 0 && n != response.ContentLength {
		err = io.ErrUnexpectedEOF
	}
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
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Validate names even on platforms that would otherwise accept Windows device
// names, alternate data streams, drive paths or case-insensitive collisions.
func archiveName(raw string) (string, error) {
	name := strings.ReplaceAll(raw, `\`, "/")
	if strings.HasPrefix(name, "/") || strings.Contains(name, ":") || strings.ContainsAny(name, "\x00\r\n") {
		return "", fmt.Errorf("unsafe RetroArch archive path")
	}
	parts := strings.Split(strings.TrimSuffix(name, "/"), "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || strings.TrimRight(part, " .") != part {
			return "", fmt.Errorf("unsafe RetroArch archive path")
		}
		base := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || regexp.MustCompile(`^(COM|LPT)[1-9]$`).MatchString(base) {
			return "", fmt.Errorf("unsafe RetroArch archive path")
		}
	}
	return filepath.FromSlash(strings.Join(parts, "/")), nil
}
func extract7z(ctx context.Context, path, destination string, progress func(string)) error {
	r, err := sevenzip.OpenReader(path)
	if err != nil {
		return err
	}
	defer r.Close()
	if len(r.File) > 100000 {
		return fmt.Errorf("RetroArch archive contains too many files")
	}
	if err = os.MkdirAll(destination, 0700); err != nil {
		return err
	}
	root, err := os.OpenRoot(destination)
	if err != nil {
		return err
	}
	defer root.Close()
	seen := map[string]bool{}
	var expanded uint64
	for _, file := range r.File {
		if err = ctx.Err(); err != nil {
			return err
		}
		name, e := archiveName(file.Name)
		if e != nil {
			return e
		}
		mode := file.Mode()
		if !mode.IsRegular() && !mode.IsDir() {
			return fmt.Errorf("RetroArch archive contains a link or special file")
		}
		key := strings.ToLower(name)
		if seen[key] {
			return fmt.Errorf("duplicate RetroArch archive path")
		}
		seen[key] = true
		if mode.IsDir() {
			if err = root.MkdirAll(name, 0700); err != nil {
				return err
			}
			continue
		}
		expanded += file.UncompressedSize
		if expanded > 4*1024*1024*1024 || file.UncompressedSize > 1024*1024*1024 {
			return fmt.Errorf("RetroArch archive expands beyond the size limit")
		}
		if err = root.MkdirAll(filepath.Dir(name), 0700); err != nil {
			return err
		}
		input, e := file.Open()
		if e != nil {
			return e
		}
		output, e := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			input.Close()
			return e
		}
		hash := crc32.NewIEEE()
		n, e := io.Copy(io.MultiWriter(output, hash), countedReader{ctx: ctx, r: io.LimitReader(input, int64(file.UncompressedSize)+1)})
		inputErr := input.Close()
		closeErr := output.Close()
		if e != nil {
			return e
		}
		if inputErr != nil {
			return inputErr
		}
		if closeErr != nil {
			return closeErr
		}
		if uint64(n) != file.UncompressedSize || hash.Sum32() != file.CRC32 {
			return fmt.Errorf("RetroArch archive integrity check failed")
		}
		progress(file.Name)
	}
	return nil
}
func extractCore(ctx context.Context, path, destination, expected string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer r.Close()
	if len(r.File) > 100 {
		return fmt.Errorf("invalid RetroArch core package")
	}
	found := false
	for _, file := range r.File {
		name, e := archiveName(file.Name)
		if e != nil {
			return e
		}
		if filepath.Base(name) != expected || file.FileInfo().IsDir() {
			continue
		}
		if found || !file.Mode().IsRegular() || file.UncompressedSize64 > 256*1024*1024 {
			return fmt.Errorf("invalid RetroArch core package")
		}
		found = true
		in, e := file.Open()
		if e != nil {
			return e
		}
		out, e := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			in.Close()
			return e
		}
		n, e := io.Copy(out, countedReader{ctx: ctx, r: io.LimitReader(in, 256*1024*1024+1)})
		inErr := in.Close()
		outErr := out.Close()
		if e != nil {
			return e
		}
		if inErr != nil {
			return inErr
		}
		if outErr != nil {
			return outErr
		}
		if uint64(n) != file.UncompressedSize64 {
			return fmt.Errorf("RetroArch core integrity check failed")
		}
	}
	if !found {
		return fmt.Errorf("RetroArch core DLL is missing from its package")
	}
	return nil
}

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: 20 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) > 5 || !officialURL(req.URL.String()) {
			return fmt.Errorf("RetroArch download redirect rejected")
		}
		return nil
	}}
}
