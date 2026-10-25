package httpapi

import (
	"example.edu/6cc545/cw1/internal/state"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWeek03_PeerAPI_GetEmptyAndInvalid(t *testing.T) {
	h := Peers(state.NewRegistry())

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/peers", nil))
	if rr.Code != 200 || strings.TrimSpace(rr.Body.String()) != "[]" {
		t.Fatalf("empty registry: %d %q", rr.Code, rr.Body.String())
	}

	bad := `{"id":"x","http_address":"broken","grpc_address":"127.0.0.1:9092","status":"unknown"}`
	r := httptest.NewRequest(http.MethodPost, "/v1/peers", strings.NewReader(bad))
	r.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, r)
	if rr.Code != 400 {
		t.Fatalf("want 400 got %d %s", rr.Code, rr.Body.String())
	}
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/peers", nil))
	if strings.TrimSpace(rr.Body.String()) != "[]" {
		t.Fatal("rejected peer changed the registry")
	}
}