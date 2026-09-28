package server

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Kinds of objects a comment can be attached to.
const (
	commentTargetArea = "area"
	commentTargetPin  = "pin"
)

// Comment is a text note attached to an area or a pin.
type Comment struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`

	mapID      string
	targetKind string
	targetID   string
}

// commentInput is the body of a POST request.
type commentInput struct {
	Text string `json:"text"`
}

// comments is in-memory mock data, guarded by commentsMu since handlers mutate it.
// Lock order is always pinsMu before areasMu before commentsMu.
var (
	commentsMu    sync.Mutex
	comments      = []Comment{}
	nextCommentID = 1
)

func handleListAreaComments(w http.ResponseWriter, r *http.Request) {
	mapID, areaID := r.PathValue("id"), r.PathValue("areaId")

	areasMu.Lock()
	defer areasMu.Unlock()

	if !areaExistsLocked(mapID, areaID) {
		writeError(w, http.StatusNotFound, "area not found")
		return
	}
	writeJSON(w, http.StatusOK, listComments(mapID, commentTargetArea, areaID))
}

func handleCreateAreaComment(w http.ResponseWriter, r *http.Request) {
	mapID, areaID := r.PathValue("id"), r.PathValue("areaId")

	var in commentInput
	if !decodeJSON(w, r, &in) {
		return
	}

	// Hold areasMu until the insert so the area can't be deleted in between.
	areasMu.Lock()
	defer areasMu.Unlock()

	if !areaExistsLocked(mapID, areaID) {
		writeError(w, http.StatusNotFound, "area not found")
		return
	}
	createComment(w, in, mapID, commentTargetArea, areaID)
}

func handleListPinComments(w http.ResponseWriter, r *http.Request) {
	mapID, pinID := r.PathValue("id"), r.PathValue("pinId")

	pinsMu.Lock()
	defer pinsMu.Unlock()

	if !pinExistsLocked(mapID, pinID) {
		writeError(w, http.StatusNotFound, "pin not found")
		return
	}
	writeJSON(w, http.StatusOK, listComments(mapID, commentTargetPin, pinID))
}

func handleCreatePinComment(w http.ResponseWriter, r *http.Request) {
	mapID, pinID := r.PathValue("id"), r.PathValue("pinId")

	var in commentInput
	if !decodeJSON(w, r, &in) {
		return
	}

	// Hold pinsMu until the insert so the pin can't be deleted in between.
	pinsMu.Lock()
	defer pinsMu.Unlock()

	if !pinExistsLocked(mapID, pinID) {
		writeError(w, http.StatusNotFound, "pin not found")
		return
	}
	createComment(w, in, mapID, commentTargetPin, pinID)
}

// listComments returns the comments on one target, oldest first.
func listComments(mapID, kind, targetID string) []Comment {
	commentsMu.Lock()
	defer commentsMu.Unlock()

	result := []Comment{}
	for _, c := range comments {
		if c.mapID == mapID && c.targetKind == kind && c.targetID == targetID {
			result = append(result, c)
		}
	}
	return result
}

// createComment validates the input and stores a new comment on the target.
// The caller must hold the target's mutex and have checked it exists.
func createComment(w http.ResponseWriter, in commentInput, mapID, kind, targetID string) {
	c := Comment{
		Text:       strings.TrimSpace(in.Text),
		CreatedAt:  time.Now().UTC(),
		mapID:      mapID,
		targetKind: kind,
		targetID:   targetID,
	}
	if c.Text == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}

	commentsMu.Lock()
	defer commentsMu.Unlock()

	c.ID = strconv.Itoa(nextCommentID)
	nextCommentID++
	comments = append(comments, c)
	writeJSON(w, http.StatusCreated, c)
}

// deleteComments removes every comment on the target. The caller must hold
// the target's mutex.
func deleteComments(mapID, kind, targetID string) {
	commentsMu.Lock()
	defer commentsMu.Unlock()

	kept := comments[:0]
	for _, c := range comments {
		if c.mapID != mapID || c.targetKind != kind || c.targetID != targetID {
			kept = append(kept, c)
		}
	}
	comments = kept
}
