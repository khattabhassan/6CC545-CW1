package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"example.edu/6cc545/cw1/internal/state"
	"io"
	"mime"
	"net/http"
)

const MaxBody = 4096

// DecodeOne enforces a single bounded JSON object. No state is mutated here.
func DecodeOne(w http.ResponseWriter, r *http.Request, dst any) error {
	mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mt != "application/json" {
		return &InputError{http.StatusUnsupportedMediaType, "unsupported_media_type"}
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxBody)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			return &InputError{http.StatusRequestEntityTooLarge, "body_too_large"}
		}
		return &InputError{http.StatusBadRequest, "unreadable_body"}
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return &InputError{http.StatusBadRequest, "empty_body"}
	}
	if trimmed[0] != '{' {
		return &InputError{http.StatusBadRequest, "expected_json_object"}
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return &InputError{http.StatusBadRequest, "invalid_json"}
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return &InputError{http.StatusBadRequest, "trailing_data"}
	}
	return nil
}

type InputError struct {
	Status int
	Label  string
}

func (e *InputError) Error() string { return e.Label }
func RejectInput(w http.ResponseWriter, e error) {
	var in *InputError
	if errors.As(e, &in) {
		ErrorJSON(w, in.Status, in.Label)
		return
	}
	ErrorJSON(w, 400, "invalid_input")
}

func StatusResource(s *state.Status) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			WriteJSON(w, http.StatusOK, map[string]string{"status": s.Get()})
		case http.MethodPost:
			var req struct {
				Status string `json:"status"`
			}
			if err := DecodeOne(w, r, &req); err != nil {
				RejectInput(w, err)
				return
			}
			if err := s.Set(req.Status); err != nil {
				ErrorJSON(w, http.StatusBadRequest, "invalid_status")
				return
			}
			WriteJSON(w, http.StatusOK, map[string]string{"status": s.Get()})
		default:
			w.Header().Set("Allow", "GET, POST")
			ErrorJSON(w, http.StatusMethodNotAllowed, "method_not_allowed")
		}
	}
}