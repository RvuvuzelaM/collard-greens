// Package server wires up the HTTP router, handlers and middleware for the API.
package server

import (
	"net/http"

	"green-field/backend/internal/store"
)

type server struct {
	store *store.Store
}

// New builds the root http.Handler with all routes and middleware attached.
func New(st *store.Store) http.Handler {
	s := &server{store: st}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /maps", s.handleListMaps)
	mux.HandleFunc("POST /maps", s.handleCreateMap)
	mux.HandleFunc("GET /maps/{id}", s.handleGetMap)
	mux.HandleFunc("PUT /maps/{id}", s.handleUpdateMap)
	mux.HandleFunc("/", handleNotFound)

	return logging(cors(mux))
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleNotFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "not found")
}
