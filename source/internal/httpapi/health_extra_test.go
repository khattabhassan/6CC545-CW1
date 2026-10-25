package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWeek01_WrongMethodReturnsJSONError(t *testing.T) {
	rr := httptest.NewRecorder()
	Health("node-a").ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/healthz", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rr.Code)
	}
	if got := rr.Header().Get("Allow"); got != "GET" {
		t.Fatalf("want Allow: GET, got %q", got)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if body["error"] != "method_not_allowed" {
		t.Fatalf("unexpected error body: %v", body)
	}
}