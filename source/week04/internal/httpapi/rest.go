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
	// TODO W02_DECODE: implement the contract described in TASKS.md.
	return &InputError{http.StatusNotImplemented, "TODO W02_DECODE"}
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
		// TODO W02_STATUS: implement the contract described in TASKS.md.
		ErrorJSON(w, http.StatusNotImplemented, "TODO W02_STATUS")
	}
}

// Preserve imports while student bodies are replaced. These declarations are not solutions.
var _ = bytes.NewReader
var _ = json.NewDecoder
var _ = io.EOF
var _ = mime.ParseMediaType
