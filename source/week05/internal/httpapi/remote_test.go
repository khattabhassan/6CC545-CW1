package httpapi

import (
	"encoding/json"
	pb "example.edu/6cc545/cw1/api/files/v1"
	"example.edu/6cc545/cw1/internal/model"
	"example.edu/6cc545/cw1/internal/rpcfiles"
	"example.edu/6cc545/cw1/internal/state"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func startRPCForTest(t *testing.T, nodeID string, files []model.FileMetadata) string {
	t.Helper()
	cat := state.NewCatalog()
	if err := cat.Replace(files); err != nil {
		t.Fatal(err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer()
	pb.RegisterFileServiceServer(s, &rpcfiles.Service{NodeID: nodeID, Catalog: cat})
	done := make(chan struct{})
	go func() { defer close(done); _ = s.Serve(ln) }()
	t.Cleanup(func() { s.Stop(); _ = ln.Close(); <-done })
	return ln.Addr().String()
}

func closedAddress(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

func TestWeek05_MapRPCError(t *testing.T) {
	cases := []struct {
		code codes.Code
		want int
	}{
		{codes.InvalidArgument, 400},
		{codes.Unavailable, 503},
		{codes.DeadlineExceeded, 504},
		{codes.Canceled, 503},
		{codes.Internal, 502},
	}
	for _, tc := range cases {
		got, label := mapRPCError(status.Error(tc.code, "x"))
		if got != tc.want || label == "" {
			t.Errorf("%v: want %d got %d (%q)", tc.code, tc.want, got, label)
		}
	}
}

func TestWeek05_RemoteFilesHTTP(t *testing.T) {
	goodAddr := startRPCForTest(t, "node-b", []model.FileMetadata{
		{Name: "diagram.txt", Size: 25, SHA256: strings.Repeat("b", 64), ModifiedUnix: 1},
	})
	reg := state.NewRegistry()
	for id, addr := range map[string]string{"node-b": goodAddr, "node-c": closedAddress(t)} {
		if err := reg.Add(model.Peer{ID: id, HTTPAddress: "127.0.0.1:8082", GRPCAddress: addr, Status: "unknown"}); err != nil {
			t.Fatal(err)
		}
	}
	h := RemoteFiles(reg, "node-a", 300*time.Millisecond)

	tests := []struct {
		name, method, target string
		want                 []int
	}{
		{"missing peer_id", "GET", "/v1/remote-files", []int{400}},
		{"malformed peer_id", "GET", "/v1/remote-files?peer_id=bad%20id!", []int{400}},
		{"unknown peer", "GET", "/v1/remote-files?peer_id=node-z", []int{404}},
		{"wrong method", "POST", "/v1/remote-files?peer_id=node-b", []int{405}},
		{"peer down", "GET", "/v1/remote-files?peer_id=node-c", []int{503, 504}},
		{"success", "GET", "/v1/remote-files?peer_id=node-b", []int{200}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(tc.method, tc.target, nil))
			ok := false
			for _, c := range tc.want {
				if rr.Code == c {
					ok = true
				}
			}
			if !ok {
				t.Fatalf("want one of %v, got %d: %s", tc.want, rr.Code, rr.Body.String())
			}
			if rr.Code == http.StatusOK {
				var got remoteFilesResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if got.NodeID != "node-b" || len(got.Files) != 1 || got.Files[0].Name != "diagram.txt" {
					t.Fatalf("wrong body: %s", rr.Body.String())
				}
			}
		})
	}
}