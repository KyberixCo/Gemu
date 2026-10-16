package content

import (
	"context"
	"errors"
	"testing"

	"kyberix.com/gemu/internal/domain"
	"kyberix.com/gemu/internal/library"
)

func TestScreenScraperDoesNotSearchRetroROMsAsPCWindows(t *testing.T) {
	p := NewScreenScraper()
	_, err := p.Fetch(context.Background(), domain.Game{Title: "Example", Installations: []domain.Installation{{Provider: "retroarch", ExternalID: "nes"}}}, nil)
	if err == nil || !errors.Is(err, library.ErrContentUnavailable) {
		t.Fatal("retro ROM incorrectly eligible for PC Windows search", err)
	}
}
