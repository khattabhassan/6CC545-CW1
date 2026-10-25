package httpapi

import (
	"errors"
	"example.edu/6cc545/cw1/internal/model"
	"example.edu/6cc545/cw1/internal/state"
	"net/http"
)

func Peers(s *state.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			WriteJSON(w, http.StatusOK, s.Snapshot())
		case http.MethodPost:
			var p model.Peer // a fresh value for each request
			if err := DecodeOne(w, r, &p); err != nil {
				RejectInput(w, err) // reuses the Week 2 decoder: 415 / 413 / 400
				return
			}
			if err := s.Add(p); err != nil {
				if errors.Is(err, state.ErrDuplicate) {
					ErrorJSON(w, http.StatusConflict, "duplicate_peer")
					return
				}
				ErrorJSON(w, http.StatusBadRequest, "invalid_peer")
				return
			}
			WriteJSON(w, http.StatusCreated, p)
		default:
			w.Header().Set("Allow", "GET, POST")
			ErrorJSON(w, http.StatusMethodNotAllowed, "method_not_allowed")
		}
	}
}
func Files(c *state.Catalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.Header().Set("Allow", "GET")
			ErrorJSON(w, 405, "method_not_allowed")
			return
		}
		WriteJSON(w, 200, c.Snapshot())
	}
}