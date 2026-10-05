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

// Area is a marked region on a map. Coords are [x, y] pixel positions on the
// map image and Color is [r, g, b].
type Area struct {
	ID        string    `json:"id"`
	MapID     string    `json:"mapId"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Color     []int     `json:"color"`
	Coords    [][]int   `json:"coords"`
	CreatedAt time.Time `json:"createdAt"`
}

// db is the shape of the JSON file on disk.
type db struct {
	Maps  []Map  `json:"maps"`
	Areas []Area `json:"areas"`
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

// all returns the whole database, so maps and their areas come from one read.
func (s *store) all() (db, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load()
}

func findMap(d db, id string) (Map, bool) {
	for _, m := range d.Maps {
		if m.ID == id {
			return m, true
		}
	}
	return Map{}, false
}

// addArea stores a and reports whether its map exists. Nothing is saved if the
// map doesn't exist or check returns an error, which is passed through unchanged.
func (s *store) addArea(a Area, check func(Map) error) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, err := s.load()
	if err != nil {
		return false, err
	}
	m, ok := findMap(d, a.MapID)
	if !ok {
		return false, nil
	}
	if err := check(m); err != nil {
		return true, err
	}
	d.Areas = append(d.Areas, a)
	return true, s.save(d)
}

// listAreas returns the areas of a map and reports whether the map exists.
func (s *store) listAreas(mapID string) ([]Area, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, err := s.load()
	if err != nil {
		return nil, false, err
	}
	if _, ok := findMap(d, mapID); !ok {
		return nil, false, nil
	}
	areas := []Area{}
	for _, a := range d.Areas {
		if a.MapID == mapID {
			areas = append(areas, a)
		}
	}
	return areas, true, nil
}

// updateArea runs apply on the stored area and the map it belongs to, then saves
// it, unless apply returns an error, which is passed through unchanged.
func (s *store) updateArea(id string, apply func(*Area, Map) error) (Area, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, err := s.load()
	if err != nil {
		return Area{}, false, err
	}
	for i := range d.Areas {
		if d.Areas[i].ID != id {
			continue
		}
		m, _ := findMap(d, d.Areas[i].MapID)
		if err := apply(&d.Areas[i], m); err != nil {
			return Area{}, true, err
		}
		return d.Areas[i], true, s.save(d)
	}
	return Area{}, false, nil
}

func (s *store) deleteArea(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, err := s.load()
	if err != nil {
		return false, err
	}
	for i, a := range d.Areas {
		if a.ID == id {
			d.Areas = append(d.Areas[:i], d.Areas[i+1:]...)
			return true, s.save(d)
		}
	}
	return false, nil
}
