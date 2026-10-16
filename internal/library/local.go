package library

import (
	"database/sql"
	"sort"
)

// PublishLocal atomically commits a local connector and its configuration.
// Existing metadata/art/player state survives rescanning the same stable IDs.
func (s *Service) PublishLocal(source SourceConfig, entries []Entry, saveConfig func(*sql.Tx) error) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot, err := s.Load()
	if err != nil {
		return Snapshot{}, err
	}
	previous := map[string]Entry{}
	games := []Entry{}
	for _, entry := range snapshot.Games {
		if entry.SourceID == source.ID {
			previous[entry.Game.ID] = entry
		} else {
			games = append(games, entry)
		}
	}
	for _, entry := range entries {
		if old, ok := previous[entry.Game.ID]; ok {
			entry.Metadata = old.Metadata
			entry.Player = old.Player
		}
		games = append(games, entry)
	}
	sources := []SourceConfig{}
	for _, item := range snapshot.Sources {
		if item.ID != source.ID {
			sources = append(sources, item)
		}
	}
	snapshot.Sources = append(sources, source)
	snapshot.Games = games
	sort.SliceStable(snapshot.Games, func(i, j int) bool { return snapshot.Games[i].Game.Title < snapshot.Games[j].Game.Title })
	if err = s.saveSQLiteWith(snapshot, saveConfig); err != nil {
		return Snapshot{}, err
	}
	snapshot.revision++
	return snapshot, nil
}
