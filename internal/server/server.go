// Package server wires up the HTTP router, handlers and middleware for the API.
package server

import "net/http"

// New builds the root http.Handler with all routes and middleware attached.
func New() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /maps", handleListMaps)
	mux.HandleFunc("GET /maps/{id}", handleGetMap)
	mux.HandleFunc("GET /maps/{id}/areas", handleListAreas)
	mux.HandleFunc("POST /maps/{id}/areas", handleCreateArea)
	mux.HandleFunc("GET /maps/{id}/areas/{areaId}", handleGetArea)
	mux.HandleFunc("PATCH /maps/{id}/areas/{areaId}", handleUpdateArea)
	mux.HandleFunc("DELETE /maps/{id}/areas/{areaId}", handleDeleteArea)
	mux.HandleFunc("GET /maps/{id}/areas/{areaId}/comments", handleListAreaComments)
	mux.HandleFunc("POST /maps/{id}/areas/{areaId}/comments", handleCreateAreaComment)
	mux.HandleFunc("GET /maps/{id}/pins", handleListPins)
	mux.HandleFunc("POST /maps/{id}/pins", handleCreatePin)
	mux.HandleFunc("GET /maps/{id}/pins/{pinId}", handleGetPin)
	mux.HandleFunc("PATCH /maps/{id}/pins/{pinId}", handleUpdatePin)
	mux.HandleFunc("DELETE /maps/{id}/pins/{pinId}", handleDeletePin)
	mux.HandleFunc("PUT /maps/{id}/pins/{pinId}/status", handleUpdatePinStatus)
	mux.HandleFunc("GET /maps/{id}/pins/{pinId}/comments", handleListPinComments)
	mux.HandleFunc("POST /maps/{id}/pins/{pinId}/comments", handleCreatePinComment)
	mux.HandleFunc("/", handleNotFound)

	return logging(cors(mux))
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleNotFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "not found")
}
