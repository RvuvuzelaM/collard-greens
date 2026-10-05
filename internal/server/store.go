package server

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Map is an uploaded map image. Image is stored base64-encoded in the JSON file.
type Map struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ContentType string    `json:"contentType"`
	CreatedAt   time.Time `json:"createdAt"`
	Image       []byte    `json:"image"`
}

// db is the shape of the JSON file on disk.
type db struct {
	Maps []Map `json:"maps"`
}

// store is a single JSON file. It is re-read on every call, so the file can be
// edited by hand while the server is running.
type store struct {
	mu   sync.Mutex
	path string
}

func newStore(path string) *store {
	return &store{path: path}
}

func (s *store) load() (db, error) {
	var d db
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return d, nil
	}
	if err != nil {
		return d, err
	}
	if len(data) == 0 {
		return d, nil
	}
	err = json.Unmarshal(data, &d)
	return d, err
}

func (s *store) save(d db) error {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *store) addMap(m Map) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, err := s.load()
	if err != nil {
		return err
	}
	d.Maps = append(d.Maps, m)
	return s.save(d)
}

func (s *store) getMap(id string) (Map, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, err := s.load()
	if err != nil {
		return Map{}, false, err
	}
	for _, m := range d.Maps {
		if m.ID == id {
			return m, true, nil
		}
	}
	return Map{}, false, nil
}
