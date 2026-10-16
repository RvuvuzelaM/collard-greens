package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"green-field/backend/internal/store"
)

// mapSummary is what the list endpoint returns: everything except the (large) image and points.
type mapSummary struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PointCount  int       `json:"pointCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type createMapRequest struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Image       string        `json:"image"`
	Points      []store.Point `json:"points"`
}

// updateMapRequest: every field is optional, only the ones present are changed.
// Points, when present, replace the whole list.
type updateMapRequest struct {
	Name        *string        `json:"name"`
	Description *string        `json:"description"`
	Image       *string        `json:"image"`
	Points      *[]store.Point `json:"points"`
}

// handleListMaps returns a summary of every map.
func (s *server) handleListMaps(w http.ResponseWriter, _ *http.Request) {
	ms := s.store.ListMaps()
	out := make([]mapSummary, 0, len(ms))
	for _, m := range ms {
		out = append(out, mapSummary{
			ID: m.ID, Name: m.Name, Description: m.Description,
			PointCount: len(m.Points), CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// handleGetMap returns one full map including its image and points.
func (s *server) handleGetMap(w http.ResponseWriter, r *http.Request) {
	m, ok := s.store.GetMap(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// handleCreateMap adds a map from a name, description and base64 image.
func (s *server) handleCreateMap(w http.ResponseWriter, r *http.Request) {
	var req createMapRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Image == "" {
		writeError(w, http.StatusBadRequest, "image is required")
		return
	}
	image, err := normalizeImage(req.Image)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if msg := validatePoints(req.Points); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	m, err := s.store.CreateMap(store.Map{
		Name: req.Name, Description: req.Description, Image: image, Points: req.Points,
	})
	if err != nil {
		slog.Error("create map", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

// handleUpdateMap edits name, description, image and/or points of a map.
func (s *server) handleUpdateMap(w http.ResponseWriter, r *http.Request) {
	var req updateMapRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Name != nil {
		*req.Name = strings.TrimSpace(*req.Name)
		if *req.Name == "" {
			writeError(w, http.StatusBadRequest, "name must not be empty")
			return
		}
	}
	if req.Image != nil {
		image, err := normalizeImage(*req.Image)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		req.Image = &image
	}
	if req.Points != nil {
		if msg := validatePoints(*req.Points); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
	}

	m, err := s.store.UpdateMap(r.PathValue("id"), func(m *store.Map) {
		if req.Name != nil {
			m.Name = *req.Name
		}
		if req.Description != nil {
			m.Description = *req.Description
		}
		if req.Image != nil {
			m.Image = *req.Image
		}
		if req.Points != nil {
			m.Points = *req.Points
		}
	})
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		slog.Error("update map", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// normalizeImage accepts raw base64 or a data URL, checks it is an image and
// returns it as a data URL the frontend can drop straight into <img src>.
func normalizeImage(s string) (string, error) {
	raw := s
	if strings.HasPrefix(s, "data:") {
		i := strings.Index(s, ",")
		if i < 0 {
			return "", errors.New("image data URL is malformed")
		}
		raw = s[i+1:]
	}
	raw = strings.TrimSpace(raw)
	b, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		if b, err = base64.RawStdEncoding.DecodeString(raw); err != nil {
			return "", errors.New("image must be base64 encoded")
		}
	}
	ct := http.DetectContentType(b)
	if !strings.HasPrefix(ct, "image/") {
		return "", fmt.Errorf("image must be a PNG, JPEG, GIF or WebP, got %s", ct)
	}
	return "data:" + ct + ";base64," + raw, nil
}

// validatePoints trims names and returns a user-facing error message or "".
func validatePoints(ps []store.Point) string {
	for i := range ps {
		ps[i].Name = strings.TrimSpace(ps[i].Name)
		if ps[i].Name == "" {
			return fmt.Sprintf("points[%d].name is required", i)
		}
		if ps[i].Coordinates == nil {
			ps[i].Coordinates = [][]float64{}
		}
		for j, c := range ps[i].Coordinates {
			if len(c) != 2 {
				return fmt.Sprintf("points[%d].coordinates[%d] must be [x, y]", i, j)
			}
		}
	}
	return ""
}
