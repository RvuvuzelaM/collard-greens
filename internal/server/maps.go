package server

import "net/http"

// Map is a map available to the frontend.
type Map struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"imageUrl"`
}

// maps is in-memory mock data.
var maps = []Map{
	{ID: "1", Name: "Zielone Wzgorze", ImageURL: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQAqGrEcRWvJqK9uN4PlUx1mp8hNulf7YgR_yQ1PiWS5g&s=10"},
	{ID: "2", Name: "Kamienna Góra", ImageURL: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQAqGrEcRWvJqK9uN4PlUx1mp8hNulf7YgR_yQ1PiWS5g&s=10"},
	{ID: "3", Name: "Słoneczne Wybrzeże", ImageURL: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQAqGrEcRWvJqK9uN4PlUx1mp8hNulf7YgR_yQ1PiWS5g&s=10"},
}

func handleListMaps(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, maps)
}

func handleGetMap(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, m := range maps {
		if m.ID == id {
			writeJSON(w, http.StatusOK, m)
			return
		}
	}
	writeError(w, http.StatusNotFound, "map not found")
}
