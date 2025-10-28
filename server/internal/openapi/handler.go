package openapi

import (
	"net/http"
)

func MountHandlers(srv *http.ServeMux) {
	srv.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(GetSpecYAML())
	})

	srv.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		resp, err := GetSpecJSON()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(resp)
	})
}
