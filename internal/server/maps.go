package server

import (
	"crypto/rand"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const maxImageBytes = 10 << 20

// mapResponse is a Map without the image bytes.
type mapResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ContentType string    `json:"contentType"`
	CreatedAt   time.Time `json:"createdAt"`
	Areas       []Area    `json:"areas"`
}

// handleUploadMap accepts multipart/form-data with an "image" file and an optional "name".
func handleUploadMap(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxImageBytes)
		if err := r.ParseMultipartForm(maxImageBytes); err != nil {
			writeError(w, http.StatusBadRequest, "invalid multipart form: "+err.Error())
			return
		}

		file, header, err := r.FormFile("image")
		if err != nil {
			writeError(w, http.StatusBadRequest, `missing "image" file field`)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			writeError(w, http.StatusBadRequest, "could not read image: "+err.Error())
			return
		}

		contentType := http.DetectContentType(data)
		if !strings.HasPrefix(contentType, "image/") {
			writeError(w, http.StatusBadRequest, "file is not an image: "+contentType)
			return
		}
		// Areas are checked against the image size, so it must be readable.
		if _, _, err := imageSize(data); err != nil {
			writeError(w, http.StatusBadRequest, "unsupported image, use PNG, JPEG or GIF: "+contentType)
			return
		}

		name := r.FormValue("name")
		if name == "" {
			name = header.Filename
		}

		m := Map{
			ID:          rand.Text(),
			Name:        name,
			ContentType: contentType,
			CreatedAt:   time.Now().UTC(),
			Image:       data,
		}
		if err := s.addMap(m); err != nil {
			slog.Error("save map", "err", err)
			writeError(w, http.StatusInternalServerError, "could not save map")
			return
		}

		writeJSON(w, http.StatusCreated, mapResponse{
			ID:          m.ID,
			Name:        m.Name,
			ContentType: m.ContentType,
			CreatedAt:   m.CreatedAt,
			Areas:       []Area{},
		})
	}
}

// handleDownloadMap returns the raw image bytes of a map.
func handleDownloadMap(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m, ok, err := s.getMap(r.PathValue("id"))
		if err != nil {
			slog.Error("load map", "err", err)
			writeError(w, http.StatusInternalServerError, "could not load map")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "map not found")
			return
		}

		w.Header().Set("Content-Type", m.ContentType)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(m.Image); err != nil {
			slog.Error("write image", "err", err)
		}
	}
}

// handleListMaps returns metadata and areas for all maps, without the image bytes.
func handleListMaps(s *store) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		d, err := s.all()
		if err != nil {
			slog.Error("list maps", "err", err)
			writeError(w, http.StatusInternalServerError, "could not list maps")
			return
		}

		areasByMap := map[string][]Area{}
		for _, a := range d.Areas {
			areasByMap[a.MapID] = append(areasByMap[a.MapID], a)
		}

		resp := make([]mapResponse, 0, len(d.Maps))
		for _, m := range d.Maps {
			areas := areasByMap[m.ID]
			if areas == nil {
				areas = []Area{}
			}
			resp = append(resp, mapResponse{
				ID:          m.ID,
				Name:        m.Name,
				ContentType: m.ContentType,
				CreatedAt:   m.CreatedAt,
				Areas:       areas,
			})
		}
		writeJSON(w, http.StatusOK, resp)
	}
}
