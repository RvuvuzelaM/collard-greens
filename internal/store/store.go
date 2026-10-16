// Package store keeps maps (image + points) in memory, backed by a JSON file.
package store

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

// Point is a polygon drawn on the map: its vertices as pixel coordinates, with a label.
type Point struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Coordinates [][]float64 `json:"coordinates"` // [[x, y], [x, y], [x, y], ...]
}

// Map is an uploaded image (base64 data URL) with points drawn on it.
type Map struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	Points      []Point   `json:"points"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ErrNotFound is returned for an unknown map id.
var ErrNotFound = errors.New("map not found")

// Store is a mutex-guarded slice of maps persisted to <dir>/maps.json.
type Store struct {
	path string
	mu   sync.RWMutex
	maps []Map
}

// Open creates dir if needed and loads maps.json when present.
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(dir, "maps.json")}
	data, err := os.ReadFile(s.path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// fresh data dir
	case err != nil:
		return nil, err
	default:
		if err := json.Unmarshal(data, &s.maps); err != nil {
			return nil, fmt.Errorf("parse %s: %w", s.path, err)
		}
	}
	return s, nil
}

// ListMaps returns all maps in creation order (never nil).
func (s *Store) ListMaps() []Map {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.maps)
}

// GetMap looks up one map by id.
func (s *Store) GetMap(id string) (Map, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if i := s.index(id); i >= 0 {
		return s.maps[i], true
	}
	return Map{}, false
}

// CreateMap stores m with a fresh id and timestamps.
func (s *Store) CreateMap(m Map) (Map, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	m.ID = rand.Text()
	m.CreatedAt, m.UpdatedAt = now, now
	if m.Points == nil {
		m.Points = []Point{}
	}
	s.maps = append(s.maps, m)
	if err := s.flush(); err != nil {
		s.maps = s.maps[:len(s.maps)-1]
		return Map{}, err
	}
	return m, nil
}

// UpdateMap applies change to the map with the given id and persists it.
func (s *Store) UpdateMap(id string, change func(*Map)) (Map, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.index(id)
	if i < 0 {
		return Map{}, ErrNotFound
	}
	old := s.maps[i]
	m := old
	change(&m)
	m.ID, m.CreatedAt = old.ID, old.CreatedAt
	m.UpdatedAt = time.Now().UTC()
	if m.Points == nil {
		m.Points = []Point{}
	}
	s.maps[i] = m
	if err := s.flush(); err != nil {
		s.maps[i] = old
		return Map{}, err
	}
	return m, nil
}

func (s *Store) index(id string) int {
	return slices.IndexFunc(s.maps, func(m Map) bool { return m.ID == id })
}

// flush writes maps.json atomically. Caller must hold the write lock.
func (s *Store) flush() error {
	data, err := json.MarshalIndent(s.maps, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
