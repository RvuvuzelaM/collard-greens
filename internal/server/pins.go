package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Pin statuses.
const (
	PinStatusTodo = "TODO"
	PinStatusDone = "DONE"
)

// Pin is a set of points on a map, optionally linked to an area.
type Pin struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Status      string      `json:"status"`
	MapID       string      `json:"map_id"`
	AreaID      *string     `json:"area_id"`
	Coordinates [][]float64 `json:"coordinates"`
}

// pinInput is the body of a POST request. Status is not accepted: new pins
// start as TODO and change only through the status endpoint.
type pinInput struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	AreaID      *string     `json:"area_id"`
	Coordinates [][]float64 `json:"coordinates"`
}

// pinPatch holds the optional fields of a PATCH request; nil means "unchanged".
// Status is not accepted here; use the status endpoint.
type pinPatch struct {
	Name        *string        `json:"name"`
	Description *string        `json:"description"`
	AreaID      nullableString `json:"area_id"`
	Coordinates [][]float64    `json:"coordinates"`
}

// pinStatusInput is the body of a PUT .../status request.
type pinStatusInput struct {
	Status string `json:"status"`
}

// nullableString tells apart a missing field (Set == false) from an explicit
// null (Set == true, Value == nil), so PATCH can unlink a pin from its area.
type nullableString struct {
	Set   bool
	Value *string
}

func (n *nullableString) UnmarshalJSON(b []byte) error {
	n.Set = true
	if string(b) == "null" {
		n.Value = nil
		return nil
	}
	return json.Unmarshal(b, &n.Value)
}

// pins is in-memory mock data, guarded by pinsMu since handlers mutate it.
// Coordinates are [x, y] pixels on the 460x667 map image, origin top-left.
var (
	pinsMu sync.Mutex
	pins   = []Pin{
		{
			ID: "1", MapID: "1", AreaID: ptr("1"), Status: PinStatusTodo,
			Name:        "Check ambulance access",
			Description: "Make sure the ambulance bay at the hospital entrance is clear.",
			Coordinates: [][]float64{{14, 500}, {60, 497}, {62, 528}, {16, 531}},
		},
		{
			ID: "2", MapID: "2", AreaID: ptr("2"), Status: PinStatusDone,
			Name:        "Plant flowers",
			Description: "Flower bed in the middle of the roundabout island.",
			Coordinates: [][]float64{{305, 466}, {340, 466}, {342, 492}, {303, 492}},
		},
		{
			ID: "3", MapID: "3", AreaID: nil, Status: PinStatusTodo,
			Name:        "Repaint crossing",
			Description: "Zebra crossing on the road left of the parking garage has faded.",
			Coordinates: [][]float64{{146, 590}, {176, 588}, {177, 612}, {147, 614}},
		},
	}
	nextPinID = 4
)

func ptr[T any](v T) *T { return &v }

func handleListPins(w http.ResponseWriter, r *http.Request) {
	mapID := r.PathValue("id")
	if !mapExists(mapID) {
		writeError(w, http.StatusNotFound, "map not found")
		return
	}

	pinsMu.Lock()
	defer pinsMu.Unlock()

	result := []Pin{}
	for _, p := range pins {
		if p.MapID == mapID {
			result = append(result, p)
		}
	}
	writeJSON(w, http.StatusOK, result)
}

func handleGetPin(w http.ResponseWriter, r *http.Request) {
	mapID, pinID := r.PathValue("id"), r.PathValue("pinId")

	pinsMu.Lock()
	defer pinsMu.Unlock()

	for _, p := range pins {
		if p.MapID == mapID && p.ID == pinID {
			writeJSON(w, http.StatusOK, p)
			return
		}
	}
	writeError(w, http.StatusNotFound, "pin not found")
}

func handleCreatePin(w http.ResponseWriter, r *http.Request) {
	mapID := r.PathValue("id")
	if !mapExists(mapID) {
		writeError(w, http.StatusNotFound, "map not found")
		return
	}

	var in pinInput
	if !decodeJSON(w, r, &in) {
		return
	}
	p := Pin{
		Name:        strings.TrimSpace(in.Name),
		Description: in.Description,
		Status:      PinStatusTodo,
		MapID:       mapID,
		AreaID:      in.AreaID,
		Coordinates: in.Coordinates,
	}
	// Hold pinsMu while validating so an area can't be deleted between the
	// area_id check and the insert.
	pinsMu.Lock()
	defer pinsMu.Unlock()

	if msg := validatePin(p); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	p.ID = strconv.Itoa(nextPinID)
	nextPinID++
	pins = append(pins, p)
	writeJSON(w, http.StatusCreated, p)
}

func handleUpdatePin(w http.ResponseWriter, r *http.Request) {
	mapID, pinID := r.PathValue("id"), r.PathValue("pinId")

	var patch pinPatch
	if !decodeJSON(w, r, &patch) {
		return
	}

	pinsMu.Lock()
	defer pinsMu.Unlock()

	for i, p := range pins {
		if p.MapID != mapID || p.ID != pinID {
			continue
		}
		if patch.Name != nil {
			p.Name = strings.TrimSpace(*patch.Name)
		}
		if patch.Description != nil {
			p.Description = *patch.Description
		}
		if patch.AreaID.Set {
			p.AreaID = patch.AreaID.Value
		}
		if patch.Coordinates != nil {
			p.Coordinates = patch.Coordinates
		}
		if msg := validatePin(p); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		pins[i] = p
		writeJSON(w, http.StatusOK, p)
		return
	}
	writeError(w, http.StatusNotFound, "pin not found")
}

func handleUpdatePinStatus(w http.ResponseWriter, r *http.Request) {
	mapID, pinID := r.PathValue("id"), r.PathValue("pinId")

	var in pinStatusInput
	if !decodeJSON(w, r, &in) {
		return
	}
	if msg := validatePinStatus(in.Status); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	pinsMu.Lock()
	defer pinsMu.Unlock()

	for i, p := range pins {
		if p.MapID == mapID && p.ID == pinID {
			pins[i].Status = in.Status
			writeJSON(w, http.StatusOK, pins[i])
			return
		}
	}
	writeError(w, http.StatusNotFound, "pin not found")
}

func handleDeletePin(w http.ResponseWriter, r *http.Request) {
	mapID, pinID := r.PathValue("id"), r.PathValue("pinId")

	pinsMu.Lock()
	defer pinsMu.Unlock()

	for i, p := range pins {
		if p.MapID == mapID && p.ID == pinID {
			pins = append(pins[:i], pins[i+1:]...)
			deleteComments(mapID, commentTargetPin, pinID)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	writeError(w, http.StatusNotFound, "pin not found")
}

// validatePin returns an error message, or "" if the pin is valid.
func validatePin(p Pin) string {
	if p.Name == "" {
		return "name is required"
	}
	if msg := validatePinStatus(p.Status); msg != "" {
		return msg
	}
	if len(p.Coordinates) == 0 {
		return "coordinates must contain at least 1 point"
	}
	for _, pt := range p.Coordinates {
		if len(pt) != 2 {
			return "each coordinate must be a [x, y] pair"
		}
	}
	if p.AreaID != nil && !areaExists(p.MapID, *p.AreaID) {
		return "area_id does not reference an area on this map"
	}
	return ""
}

// validatePinStatus returns an error message, or "" if the status is valid.
func validatePinStatus(status string) string {
	if status != PinStatusTodo && status != PinStatusDone {
		return "status must be TODO or DONE"
	}
	return ""
}

// areaExists reports whether the area is on the map. Callers may hold pinsMu;
// lock order is always pinsMu before areasMu.
func areaExists(mapID, areaID string) bool {
	areasMu.Lock()
	defer areasMu.Unlock()

	return areaExistsLocked(mapID, areaID)
}

// areaExistsLocked is areaExists for callers that already hold areasMu.
func areaExistsLocked(mapID, areaID string) bool {
	for _, a := range areas {
		if a.MapID == mapID && a.ID == areaID {
			return true
		}
	}
	return false
}

// pinExistsLocked reports whether the pin is on the map. The caller must hold pinsMu.
func pinExistsLocked(mapID, pinID string) bool {
	for _, p := range pins {
		if p.MapID == mapID && p.ID == pinID {
			return true
		}
	}
	return false
}
