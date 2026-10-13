package content

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kyberix.com/gemu/internal/library"
	"net/http"
	"strings"
	"time"
	"unicode"
)

func client() *http.Client {
	return &http.Client{Timeout: 12 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return fmt.Errorf("redirección rechazada") }}
}

// Do not propagate response bodies or transport errors: they may contain credentials.
func getJSON(ctx context.Context, c *http.Client, address, key string, out any, classify ...func(int) error) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return fmt.Errorf("solicitud inválida")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Gemu/0.1")
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("no se pudo conectar con el proveedor")
	}
	defer resp.Body.Close()
	if len(classify) > 0 {
		if err := classify[0](resp.StatusCode); err != nil {
			return err
		}
	}
	switch resp.StatusCode {
	case 401, 403:
		return fmt.Errorf("credenciales rechazadas o acceso no permitido: %w", library.ErrContentUnavailable)
	case 423, 426, 429, 430, 431:
		return fmt.Errorf("servicio ocupado o cuota agotada; inténtalo más tarde: %w", library.ErrContentUnavailable)
	case 404:
		return fmt.Errorf("juego o recurso no encontrado")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("servicio HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024+1))
	if err != nil || len(data) > 4*1024*1024 || json.Unmarshal(data, out) != nil {
		return fmt.Errorf("respuesta del proveedor inválida")
	}
	return nil
}

func normalizedTitle(title string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, title)
}

// ScreenScraper encodes identifiers and dimensions as either JSON strings or numbers.
type textNumber string

func (n *textNumber) UnmarshalJSON(data []byte) error {
	var value string
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
	} else {
		value = string(data)
	}
	*n = textNumber(value)
	return nil
}
