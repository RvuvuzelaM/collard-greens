package server

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const (
	statusTodo = "TODO"
	statusDone = "DONE"
)

// validationError is a client mistake in an area, reported as a 400.
type validationError struct{ msg string }

func (e validationError) Error() string { return e.msg }

func invalid(format string, args ...any) error {
	return validationError{fmt.Sprintf(format, args...)}
}

func validateArea(a Area) error {
	if strings.TrimSpace(a.Title) == "" {
		return invalid(`"title" is required`)
	}
	if a.Status != statusTodo && a.Status != statusDone {
		return invalid(`"status" must be %q or %q`, statusTodo, statusDone)
	}
	if len(a.Color) != 3 {
		return invalid(`"color" must be [r, g, b]`)
	}
	for _, c := range a.Color {
		if c < 0 || c > 255 {
			return invalid(`"color" values must be between 0 and 255`)
		}
	}
	if len(a.Coords) == 0 {
		return invalid(`"coords" must contain at least one [x, y] point`)
	}
	for i, p := range a.Coords {
		if len(p) != 2 {
			return invalid(`"coords"[%d] must be [x, y]`, i)
		}
		if p[0] < 0 || p[1] < 0 {
			return invalid(`"coords"[%d] must not be negative`, i)
		}
	}
	return nil
}

// areaInput is the body of POST and PATCH. Nil fields are left unchanged on PATCH.
type areaInput struct {
	Title  *string  `json:"title"`
	Status *string  `json:"status"`
	Color  *[]int   `json:"color"`
	Coords *[][]int `json:"coords"`
}

func (in areaInput) applyTo(a *Area) {
	if in.Title != nil {
		a.Title = *in.Title
	}
	if in.Status != nil {
		a.Status = *in.Status
	}
	if in.Color != nil {
		a.Color = *in.Color
	}
	if in.Coords != nil {
		a.Coords = *in.Coords
	}
}

// handleListAreas returns all areas of a map.
func handleListAreas(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		areas, ok, err := s.listAreas(r.PathValue("id"))
		if err != nil {
			slog.Error("list areas", "err", err)
			writeError(w, http.StatusInternalServerError, "could not list areas")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "map not found")
			return
		}
		writeJSON(w, http.StatusOK, areas)
	}
}

// handleCreateArea adds an area to a map. Status defaults to TODO.
func handleCreateArea(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in areaInput
		if !decodeJSON(w, r, &in) {
			return
		}

		a := Area{
			ID:        rand.Text(),
			MapID:     r.PathValue("id"),
			Status:    statusTodo,
			CreatedAt: time.Now().UTC(),
		}
		in.applyTo(&a)
		if err := validateArea(a); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		ok, err := s.addArea(a)
		if err != nil {
			slog.Error("save area", "err", err)
			writeError(w, http.StatusInternalServerError, "could not save area")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "map not found")
			return
		}
		writeJSON(w, http.StatusCreated, a)
	}
}

// handleUpdateArea changes only the fields present in the body.
func handleUpdateArea(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in areaInput
		if !decodeJSON(w, r, &in) {
			return
		}

		a, ok, err := s.updateArea(r.PathValue("id"), func(a *Area) error {
			updated := *a
			in.applyTo(&updated)
			if err := validateArea(updated); err != nil {
				return err
			}
			*a = updated
			return nil
		})
		var verr validationError
		switch {
		case errors.As(err, &verr):
			writeError(w, http.StatusBadRequest, verr.Error())
		case err != nil:
			slog.Error("update area", "err", err)
			writeError(w, http.StatusInternalServerError, "could not update area")
		case !ok:
			writeError(w, http.StatusNotFound, "area not found")
		default:
			writeJSON(w, http.StatusOK, a)
		}
	}
}

// handleDeleteArea removes an area.
func handleDeleteArea(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, err := s.deleteArea(r.PathValue("id"))
		if err != nil {
			slog.Error("delete area", "err", err)
			writeError(w, http.StatusInternalServerError, "could not delete area")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "area not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
