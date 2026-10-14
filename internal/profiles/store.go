package profiles

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"
)

type Profile struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar,omitempty"`
}
type Store struct {
	dir string
	mu  sync.Mutex
}

func New(dir string) *Store { return &Store{dir: dir} }
func (s *Store) List() ([]Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.list()
}
func (s *Store) list() ([]Profile, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, "profiles.json"))
	if os.IsNotExist(err) {
		return []Profile{}, nil
	}
	if err != nil {
		return nil, err
	}
	var items []Profile
	if err = json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("perfiles guardados inválidos: %w", err)
	}
	seen := map[string]bool{}
	for _, p := range items {
		decoded, e := hex.DecodeString(p.ID)
		if e != nil || len(decoded) != 16 || seen[p.ID] || strings.TrimSpace(p.Name) == "" || validateAvatar(p.Avatar) != nil {
			return nil, fmt.Errorf("perfiles guardados inválidos")
		}
		seen[p.ID] = true
	}
	if items == nil {
		items = []Profile{}
	}
	return items, nil
}
func (s *Store) Create(name string) (Profile, error) {
	return s.CreateWithAvatar(name, "")
}
func (s *Store) CreateWithAvatar(name, avatar string) (Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 32 || strings.ContainsAny(name, "\r\n\t") {
		return Profile{}, fmt.Errorf("usa un nombre de 1 a 32 caracteres")
	}
	if err := validateAvatar(avatar); err != nil {
		return Profile{}, err
	}
	items, err := s.list()
	if err != nil {
		return Profile{}, err
	}
	for _, p := range items {
		if strings.EqualFold(p.Name, name) {
			return Profile{}, fmt.Errorf("ya existe un perfil con ese nombre")
		}
	}
	var id [16]byte
	if _, err = rand.Read(id[:]); err != nil {
		return Profile{}, err
	}
	p := Profile{ID: hex.EncodeToString(id[:]), Name: name, Avatar: avatar}
	items = append(items, p)
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return Profile{}, err
	}
	if err = os.MkdirAll(s.dir, 0700); err != nil {
		return Profile{}, err
	}
	f, err := os.CreateTemp(s.dir, "profiles-*.tmp")
	if err != nil {
		return Profile{}, err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return Profile{}, err
	}
	if closeErr != nil {
		return Profile{}, closeErr
	}
	if err = os.Rename(f.Name(), filepath.Join(s.dir, "profiles.json")); err != nil {
		return Profile{}, err
	}
	return p, nil
}

func validateAvatar(avatar string) error {
	if avatar == "" {
		return nil
	}
	mime, encoded, ok := strings.Cut(avatar, ";base64,")
	if !ok || (mime != "data:image/jpeg" && mime != "data:image/png") || len(encoded) > (512*1024+2)/3*4 {
		return fmt.Errorf("imagen de perfil inválida")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(data) > 512*1024 {
		return fmt.Errorf("imagen de perfil inválida")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 1024 || config.Height > 1024 ||
		(format == "jpeg" && mime != "data:image/jpeg") || (format == "png" && mime != "data:image/png") ||
		(format != "jpeg" && format != "png") {
		return fmt.Errorf("imagen de perfil inválida")
	}
	return nil
}
func (s *Store) Directory(id string) (string, error) {
	items, err := s.List()
	if err != nil {
		return "", err
	}
	for _, p := range items {
		if p.ID == id {
			return filepath.Join(s.dir, "profiles", id), nil
		}
	}
	return "", fmt.Errorf("el perfil no existe")
}
