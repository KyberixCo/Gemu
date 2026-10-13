package library

import (
	"context"
	"kyberix.com/gemu/internal/domain"
)

type Profile struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Origin string `json:"origin"`
}
type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Unlocked    bool   `json:"unlocked"`
	UnlockedAt  int64  `json:"unlockedAt"`
}
type PlayerState struct {
	Profile      Profile       `json:"profile"`
	Status       string        `json:"status"`
	Achievements []Achievement `json:"achievements"`
	Message      string        `json:"message"`
}
type Metadata struct {
	Description string    `json:"description"`
	Genres      []string  `json:"genres"`
	Cover       string    `json:"cover"`
	Background  string    `json:"background"`
	Origin      string    `json:"origin"`
	Status      string    `json:"status"`
	Artworks    []Artwork `json:"artworks,omitempty"`
	Logo        string    `json:"logo,omitempty"`
	Icon        string    `json:"icon,omitempty"`
	Backgrounds []string  `json:"backgrounds,omitempty"`
	ManualArt   []string  `json:"manualArt,omitempty"`
}

// URL is remote only while preparing; published assets always use the local cache.
type Artwork struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	URL      string `json:"url"`
	Provider string `json:"provider"`
	Author   string `json:"author,omitempty"`
	Page     string `json:"page,omitempty"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}
type ContentRequest struct {
	ID          string            `json:"id"`
	Priority    int               `json:"priority"`
	Credentials map[string]string `json:"credentials"`
}
type ContentResult struct {
	Description string
	Genres      []string
	Artworks    []Artwork
	Diagnostics []string
}
type ContentProvider interface {
	Name() string
	ArtworkHosts() []string
	Check(context.Context, map[string]string) error
	Fetch(context.Context, domain.Game, map[string]string) (ContentResult, error)
}
type Entry struct {
	Game     domain.Game `json:"game"`
	SourceID string      `json:"sourceId"`
	Metadata Metadata    `json:"metadata"`
	Player   PlayerState `json:"player"`
}
type SourceConfig struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Root      string `json:"root"`
	ProfileID string `json:"profileId"`
}
type Snapshot struct {
	Version     int            `json:"version"`
	Sources     []SourceConfig `json:"sources"`
	Games       []Entry        `json:"games"`
	ArtworkUndo *ArtworkUndo   `json:"artworkUndo,omitempty"`
}
type Preview struct {
	Root        string        `json:"root"`
	Games       []domain.Game `json:"games"`
	Profiles    []Profile     `json:"profiles"`
	Diagnostics []string      `json:"diagnostics"`
}
type Request struct {
	SourceID         string           `json:"sourceId"`
	Root             string           `json:"root"`
	ProfileID        string           `json:"profileId"`
	APIKey           string           `json:"apiKey"`
	ContentProviders []ContentRequest `json:"contentProviders,omitempty"`
}
type Draft struct {
	Token       string       `json:"token"`
	Source      SourceConfig `json:"source"`
	Games       []Entry      `json:"games"`
	Diagnostics []string     `json:"diagnostics"`
}

// Sources implement the same import contract independently of their store/client.
type Source interface {
	Name() string
	ArtworkHosts() []string
	Discover(context.Context, string) (Preview, error)
	Metadata(context.Context, domain.Game) (Metadata, error)
	Player(context.Context, domain.Game, Profile, string) PlayerState
}
