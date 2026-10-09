package domain

// Game is the catalog identity; installations are separate to allow multiple stores.
type Game struct {
	ID            string
	Title         string
	Installations []Installation
}
type Installation struct {
	Provider   string
	ExternalID string
	Path       string
	Installed  bool
}
type LaunchSpec struct {
	URI        string
	Confidence string
}
