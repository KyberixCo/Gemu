package library

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func artworkDimensions(data []byte) (int, int, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	return config.Width, config.Height, err
}

func artworkURL(raw string, hosts []string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, allowed := range hosts {
		if host == allowed {
			return true
		}
	}
	return false
}
func (s *Service) cacheArtwork(ctx context.Context, raw string, hosts []string) (string, error) {
	if !artworkURL(raw, hosts) {
		return "", fmt.Errorf("origen de arte no permitido")
	}
	name := fmt.Sprintf("%x", sha256.Sum256([]byte(raw))) + ".img"
	path := filepath.Join(s.artDir, name)
	if _, err := os.Stat(path); err == nil {
		return "/library-assets/" + name, nil
	}
	client := &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 || !artworkURL(req.URL.String(), hosts) {
			return fmt.Errorf("redirección de arte rechazada")
		}
		return nil
	}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("no se pudo descargar arte")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("arte HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024+1))
	if err != nil || len(data) > 8*1024*1024 {
		return "", fmt.Errorf("arte demasiado grande o inválido")
	}
	mime := http.DetectContentType(data)
	if mime != "image/jpeg" && mime != "image/png" && mime != "image/webp" {
		return "", fmt.Errorf("formato de arte no permitido")
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", err
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		return "", err
	}
	return "/library-assets/" + name, nil
}

var assetName = regexp.MustCompile(`^[a-f0-9]{64}\.img$`)

// Source imports may have a cached cover without a gallery entry. Measure that
// file so older high-resolution covers can participate in scenery too.
func (s *Service) includeSceneryCover(m *Metadata) {
	name := strings.TrimPrefix(m.Cover, "/library-assets/")
	if name == m.Cover || !assetName.MatchString(name) {
		return
	}
	for _, art := range m.Artworks {
		if art.URL == m.Cover {
			return
		}
	}
	file, err := os.Open(filepath.Join(s.artDir, name))
	if err != nil {
		return
	}
	defer file.Close()
	config, _, err := image.DecodeConfig(file)
	if err != nil || !BackgroundSize(config.Width, config.Height) {
		return
	}
	m.Artworks = append(m.Artworks, Artwork{ID: "stored-cover", Kind: "cover", URL: m.Cover, Provider: m.Origin, Width: config.Width, Height: config.Height})
}

// Route cached images before Vite's HTML fallback in development, too.
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/library-assets/") {
			s.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/library-assets/")
	if !strings.HasPrefix(r.URL.Path, "/library-assets/") || !assetName.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	file, err := os.Open(filepath.Join(s.artDir, name))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var header [512]byte
	n, _ := file.Read(header[:])
	w.Header().Set("Content-Type", http.DetectContentType(header[:n]))
	file.Seek(0, 0)
	http.ServeContent(w, r, name, info.ModTime(), file)
}
