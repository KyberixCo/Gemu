//go:build windows

package credentials

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var header = []byte("GEMUVAULT\x01")

const maxSize = 1024 * 1024

type fileStore struct {
	dir     string
	entropy []byte
	mu      sync.Mutex
}
type payload struct {
	Version   int                          `json:"version"`
	Providers map[string]map[string]string `json:"providers"`
}

func newFileStore(dir string) *fileStore {
	absolute, _ := filepath.Abs(dir)
	scope := sha256.Sum256([]byte("Gemu.credentials.v1:" + strings.ToLower(filepath.Clean(absolute))))
	return &fileStore{dir: dir, entropy: scope[:]}
}
func (*fileStore) Capabilities() Capabilities {
	return Capabilities{Name: "Windows DPAPI", Available: true}
}
func wipe(data []byte) { clear(data) }
func (s *fileStore) read() (payload, error) {
	result := payload{Version: 1, Providers: map[string]map[string]string{}}
	file, err := os.Open(filepath.Join(s.dir, "credentials.bin"))
	if os.IsNotExist(err) {
		return result, nil
	}
	if err != nil {
		return result, fmt.Errorf("no se pudo abrir el almacén de credenciales")
	}
	defer file.Close()
	encrypted, err := io.ReadAll(io.LimitReader(file, maxSize+1))
	if err != nil || len(encrypted) > maxSize || !bytes.HasPrefix(encrypted, header) {
		return result, fmt.Errorf("almacén de credenciales inválido")
	}
	data, err := unprotect(encrypted[len(header):], s.entropy)
	if err != nil {
		return result, fmt.Errorf("no se pudieron descifrar las credenciales de este perfil")
	}
	defer wipe(data)
	result = payload{}
	if json.Unmarshal(data, &result) != nil || result.Version != 1 || result.Providers == nil {
		return result, fmt.Errorf("almacén de credenciales inválido")
	}
	for provider, values := range result.Providers {
		if validate(provider, values) != nil {
			return payload{}, fmt.Errorf("almacén de credenciales inválido")
		}
	}
	return result, nil
}
func (s *fileStore) write(value payload) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("no se pudieron preparar las credenciales")
	}
	defer wipe(data)
	encrypted, err := protect(data, s.entropy)
	if err != nil {
		return fmt.Errorf("no se pudieron proteger las credenciales; no se guardaron")
	}
	if len(encrypted) > maxSize {
		return fmt.Errorf("almacén de credenciales demasiado grande")
	}
	if os.MkdirAll(s.dir, 0700) != nil {
		return fmt.Errorf("no se pudo crear el almacén de credenciales")
	}
	file, err := os.CreateTemp(s.dir, "credentials-*.tmp")
	if err != nil {
		return fmt.Errorf("no se pudo guardar el almacén de credenciales")
	}
	defer os.Remove(file.Name())
	_, err = file.Write(append(append([]byte{}, header...), encrypted...))
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil || closeErr != nil || os.Rename(file.Name(), filepath.Join(s.dir, "credentials.bin")) != nil {
		return fmt.Errorf("no se pudo guardar el almacén de credenciales")
	}
	return nil
}
func (s *fileStore) Get(provider string) (map[string]string, error) {
	if err := validate(provider, nil); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	value, err := s.read()
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	for field, secret := range value.Providers[provider] {
		result[field] = secret
	}
	return result, nil
}
func (s *fileStore) Status(provider string) (Status, error) {
	values, err := s.Get(provider)
	if err != nil {
		return Status{}, err
	}
	status := Status{Provider: provider, SavedFields: []string{}, Storage: s.Capabilities()}
	for field, value := range values {
		if value != "" {
			status.SavedFields = append(status.SavedFields, field)
		}
	}
	sort.Strings(status.SavedFields)
	return status, nil
}

// Nonempty fields replace their saved value. Blank fields preserve it; removal
// is explicit so editing one secret cannot accidentally erase the others.
func (s *fileStore) Save(request Request) error {
	if err := validate(request.Provider, request.Fields); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	value, err := s.read()
	if err != nil {
		return err
	}
	current := value.Providers[request.Provider]
	if current == nil {
		current = map[string]string{}
	}
	changed := false
	for field, secret := range request.Fields {
		if secret != "" {
			current[field] = secret
			changed = true
		}
	}
	if !changed {
		return fmt.Errorf("introduce al menos una credencial nueva")
	}
	value.Providers[request.Provider] = current
	return s.write(value)
}
func (s *fileStore) Delete(provider string) error {
	if err := validate(provider, nil); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	value, err := s.read()
	if err != nil {
		return err
	}
	if _, exists := value.Providers[provider]; !exists {
		return nil
	}
	delete(value.Providers, provider)
	return s.write(value)
}

var _ Vault = (*fileStore)(nil)
