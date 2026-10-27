package httpapi

import (
	"errors"
	"example.edu/6cc545/cw1/internal/model"
	"example.edu/6cc545/cw1/internal/state"
	"net/http"
)

func Peers(s *state.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO W03_PEER_HTTP: implement the contract described in TASKS.md.
		ErrorJSON(w, http.StatusNotImplemented, "TODO W03_PEER_HTTP")
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

var _ = errors.Is
var _ = model.Peer{}
