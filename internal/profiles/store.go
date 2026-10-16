package profiles

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"kyberix.com/gemu/internal/storage"
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
	db, err := s.database()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return listProfiles(db)
}

type querier interface {
	Query(string, ...any) (*sql.Rows, error)
}

func listProfiles(db querier) ([]Profile, error) {
	rows, err := db.Query("SELECT id,name,avatar FROM profiles ORDER BY position")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Profile{}
	for rows.Next() {
		var p Profile
		if err = rows.Scan(&p.ID, &p.Name, &p.Avatar); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (s *Store) database() (*sql.DB, error) {
	db, err := storage.Open(s.dir)
	if err != nil {
		return nil, err
	}
	err = storage.Update(db, func(tx *sql.Tx) error {
		done, err := storage.Imported(tx, storage.Global, "profiles.json")
		if err != nil || done {
			return err
		}
		items, err := s.loadLegacy()
		if err != nil {
			return err
		}
		for i, p := range items {
			if _, err = tx.Exec("INSERT INTO scopes(id) VALUES (?)", p.ID); err != nil {
				return err
			}
			if _, err = tx.Exec("INSERT INTO profiles(id,name,avatar,position) VALUES (?,?,?,?)", p.ID, p.Name, p.Avatar, i); err != nil {
				return err
			}
		}
		return storage.MarkImported(tx, storage.Global, "profiles.json")
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func (s *Store) loadLegacy() ([]Profile, error) {
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
	db, err := s.database()
	if err != nil {
		return Profile{}, err
	}
	defer db.Close()
	var id [16]byte
	if _, err = rand.Read(id[:]); err != nil {
		return Profile{}, err
	}
	p := Profile{ID: hex.EncodeToString(id[:]), Name: name, Avatar: avatar}
	err = storage.Update(db, func(tx *sql.Tx) error {
		items, err := listProfiles(tx)
		if err != nil {
			return err
		}
		for _, existing := range items {
			if strings.EqualFold(existing.Name, name) {
				return fmt.Errorf("ya existe un perfil con ese nombre")
			}
		}
		if _, err = tx.Exec("INSERT INTO scopes(id) VALUES (?)", p.ID); err != nil {
			return err
		}
		_, err = tx.Exec("INSERT INTO profiles(id,name,avatar,position) VALUES (?,?,?,?)", p.ID, p.Name, p.Avatar, len(items))
		return err
	})
	if err != nil {
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
