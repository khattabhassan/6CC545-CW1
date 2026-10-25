package httpapi

import (
	"example.edu/6cc545/cw1/internal/state"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func padded(n int) string {
	base := `{"status":"healthy"}`
	return base + strings.Repeat(" ", n-len(base))
}

func TestWeek02_BoundaryAndNonMutation(t *testing.T) {
	cases := []struct {
		name string
		body string
		code int
	}{
		{"exactly 4096 bytes accepted", padded(MaxBody), 200},
		{"4097 bytes rejected", padded(MaxBody + 1), 413},
		{"empty body", "", 400},
		{"valid JSON then junk", `{"status":"healthy"} junk`, 400},
		{"missing status field", `{}`, 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := state.NewStatus()
			if err := s.Set("degraded"); err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/v1/status", strings.NewReader(tc.body))
			r.Header.Set("Content-Type", "application/json")
			StatusResource(s).ServeHTTP(rr, r)
			if rr.Code != tc.code {
				t.Fatalf("want %d got %d: %s", tc.code, rr.Code, rr.Body.String())
			}
			want := "degraded"
			if tc.code == 200 {
				want = "healthy"
			}
			if got := s.Get(); got != want {
				t.Fatalf("state = %q, want %q", got, want)
			}
		})
	}
}