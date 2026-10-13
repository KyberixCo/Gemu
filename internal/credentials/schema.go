package credentials

import (
	"fmt"
	"strings"
)

var fields = map[string][]string{
	// Retain legacy developer fields so existing encrypted stores remain readable.
	// ScreenScraper obtains application access exclusively from its build config.
	"steamgriddb": {"apiKey"}, "screenscraper": {"developerId", "developerPassword", "username", "password"},
	"igdb": {"clientId", "clientSecret"}, "steam": {"apiKey"}, "minimax": {"apiKey"}, "openrouter": {"apiKey"}, "compatible": {"apiKey"},
}

type Status struct {
	Provider    string       `json:"provider"`
	SavedFields []string     `json:"savedFields"`
	Storage     Capabilities `json:"storage"`
}
type Request struct {
	Provider string            `json:"provider"`
	Fields   map[string]string `json:"fields"`
}

func validate(provider string, values map[string]string) error {
	names, ok := fields[provider]
	if !ok {
		return fmt.Errorf("proveedor de credenciales no compatible")
	}
	for name, value := range values {
		valid := false
		for _, allowed := range names {
			valid = valid || name == allowed
		}
		if !valid || len(value) > 4096 || strings.ContainsRune(value, 0) {
			return fmt.Errorf("campo de credencial inválido")
		}
	}
	return nil
}
