package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Area is a polygon drawn on a map.
type Area struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	MapID       string      `json:"map_id"`
	Coordinates [][]float64 `json:"coordinates"`
}

// areaInput is the body of a POST request.
type areaInput struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Coordinates [][]float64 `json:"coordinates"`
}

// areaPatch holds the optional fields of a PATCH request; nil means "unchanged".
type areaPatch struct {
	Name        *string     `json:"name"`
	Description *string     `json:"description"`
	Coordinates [][]float64 `json:"coordinates"`
}

// areas is in-memory mock data, guarded by areasMu since handlers mutate it.
// Coordinates are [x, y] pixels on the 460x667 map image, origin top-left.
var (
	areasMu sync.Mutex
	areas   = []Area{
		{
			ID: "1", MapID: "1", Name: "Hospital",
			Description: "Hospital building with the red H and cross signs on the west side of the map.",
			Coordinates: [][]float64{{8, 378}, {92, 370}, {96, 525}, {10, 532}},
		},
		{
			ID: "2", MapID: "2", Name: "Roundabout",
			Description: "Green island of the roundabout in the lower middle of the map.",
			Coordinates: [][]float64{{322, 428}, {354, 441}, {367, 473}, {354, 505}, {322, 518}, {290, 505}, {277, 473}, {290, 441}},
		},
		{
			ID: "3", MapID: "3", Name: "Parking garage",
			Description: "Blue multi-storey parking garage near the bottom of the map.",
			Coordinates: [][]float64{{195, 548}, {258, 532}, {268, 625}, {205, 645}},
		},
	}
	nextAreaID = 4
)

const maxBodyBytes = 1 << 20

func handleListAreas(w http.ResponseWriter, r *http.Request) {
	mapID := r.PathValue("id")
	if !mapExists(mapID) {
		writeError(w, http.StatusNotFound, "map not found")
		return
	}

	areasMu.Lock()
	defer areasMu.Unlock()

	result := []Area{}
	for _, a := range areas {
		if a.MapID == mapID {
			result = append(result, a)
		}
	}
	writeJSON(w, http.StatusOK, result)
}

func handleGetArea(w http.ResponseWriter, r *http.Request) {
	mapID, areaID := r.PathValue("id"), r.PathValue("areaId")

	areasMu.Lock()
	defer areasMu.Unlock()

	for _, a := range areas {
		if a.MapID == mapID && a.ID == areaID {
			writeJSON(w, http.StatusOK, a)
			return
		}
	}
	writeError(w, http.StatusNotFound, "area not found")
}

func handleCreateArea(w http.ResponseWriter, r *http.Request) {
	mapID := r.PathValue("id")
	if !mapExists(mapID) {
		writeError(w, http.StatusNotFound, "map not found")
		return
	}

	var in areaInput
	if !decodeJSON(w, r, &in) {
		return
	}
	a := Area{
		Name:        strings.TrimSpace(in.Name),
		Description: in.Description,
		MapID:       mapID,
		Coordinates: in.Coordinates,
	}
	if msg := validateArea(a); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	areasMu.Lock()
	defer areasMu.Unlock()

	a.ID = strconv.Itoa(nextAreaID)
	nextAreaID++
	areas = append(areas, a)
	writeJSON(w, http.StatusCreated, a)
}

func handleUpdateArea(w http.ResponseWriter, r *http.Request) {
	mapID, areaID := r.PathValue("id"), r.PathValue("areaId")

	var p areaPatch
	if !decodeJSON(w, r, &p) {
		return
	}

	areasMu.Lock()
	defer areasMu.Unlock()

	for i, a := range areas {
		if a.MapID != mapID || a.ID != areaID {
			continue
		}
		if p.Name != nil {
			a.Name = strings.TrimSpace(*p.Name)
		}
		if p.Description != nil {
			a.Description = *p.Description
		}
		if p.Coordinates != nil {
			a.Coordinates = p.Coordinates
		}
		if msg := validateArea(a); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		areas[i] = a
		writeJSON(w, http.StatusOK, a)
		return
	}
	writeError(w, http.StatusNotFound, "area not found")
}

// validateArea returns an error message, or "" if the area is valid.
func validateArea(a Area) string {
	if a.Name == "" {
		return "name is required"
	}
	if len(a.Coordinates) < 3 {
		return "coordinates must contain at least 3 points"
	}
	for _, pt := range a.Coordinates {
		if len(pt) != 2 {
			return "each coordinate must be a [x, y] pair"
		}
	}
	return ""
}

func mapExists(id string) bool {
	for _, m := range maps {
		if m.ID == id {
			return true
		}
	}
	return false
}

// decodeJSON reads the request body into v, writing a 400 on failure.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return false
	}
	return true
}
