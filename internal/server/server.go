// Package server wires up the HTTP router, handlers and middleware for the API.
package server

import "net/http"

// New builds the root http.Handler with all routes and middleware attached.
func New() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /maps", handleListMaps)
	mux.HandleFunc("GET /maps/{id}", handleGetMap)
	mux.HandleFunc("/", handleNotFound)

	return logging(cors(mux))
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleNotFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "not found")
}
