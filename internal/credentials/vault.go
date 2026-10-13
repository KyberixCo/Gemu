package credentials

import "errors"

var ErrUnavailable = errors.New("no hay un almacén seguro disponible en este sistema operativo")

// Vault is the OS-independent credential contract. Secrets are consumed only
// by Go; Status exposes presence and backend capabilities to the frontend.
// Implementations must preserve blank fields on Save, isolate the supplied
// profile scope, serialize writes, and never fall back to plaintext storage.
type Vault interface {
	Get(provider string) (map[string]string, error)
	Status(provider string) (Status, error)
	Save(Request) error
	Delete(provider string) error
	Capabilities() Capabilities
}

type Capabilities struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
}
