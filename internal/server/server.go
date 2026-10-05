// Package server wires up the HTTP router, handlers and middleware for the API.
package server

import (
	"net/http"
	"os"
)

// New builds the root http.Handler with all routes and middleware attached.
func New() http.Handler {
	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = "data/db.json"
	}
	s := newStore(dataFile)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /maps", handleListMaps(s))
	mux.HandleFunc("POST /maps", handleUploadMap(s))
	mux.HandleFunc("GET /maps/{id}/image", handleDownloadMap(s))
	mux.HandleFunc("GET /maps/{id}/areas", handleListAreas(s))
	mux.HandleFunc("POST /maps/{id}/areas", handleCreateArea(s))
	mux.HandleFunc("PATCH /areas/{id}", handleUpdateArea(s))
	mux.HandleFunc("DELETE /areas/{id}", handleDeleteArea(s))
	mux.HandleFunc("/", handleNotFound)

	return logging(cors(mux))
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleNotFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "not found")
}
