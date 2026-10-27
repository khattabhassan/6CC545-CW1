package rpcfiles

import (
	"context"
	pb "example.edu/6cc545/cw1/api/files/v1"
	"example.edu/6cc545/cw1/internal/model"
	"example.edu/6cc545/cw1/internal/state"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"strings"
	"testing"
	"time"
)

func startTestServer(t *testing.T, nodeID string, cat Catalog) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer()
	pb.RegisterFileServiceServer(s, &Service{NodeID: nodeID, Catalog: cat})
	done := make(chan struct{})
	go func() { defer close(done); _ = s.Serve(ln) }()
	t.Cleanup(func() { s.Stop(); _ = ln.Close(); <-done })
	return ln.Addr().String()
}

func catalogWith(t *testing.T, names ...string) *state.Catalog {
	t.Helper()
	var files []model.FileMetadata
	for _, n := range names {
		files = append(files, model.FileMetadata{Name: n, Size: 1, SHA256: strings.Repeat("c", 64), ModifiedUnix: 1})
	}
	c := state.NewCatalog()
	if err := c.Replace(files); err != nil {
		t.Fatal(err)
	}
	return c
}

// An empty catalogue is a SUCCESSFUL empty result, not a failure.
func TestWeek05_EmptyCatalogIsSuccess(t *testing.T) {
	addr := startTestServer(t, "node-b", catalogWith(t))
	got, err := Fetch(context.Background(), addr, "node-a", time.Second)
	if err != nil {
		t.Fatalf("empty catalogue returned an error: %v", err)
	}
	if len(got.GetFiles()) != 0 || got.GetNodeId() != "node-b" {
		t.Fatalf("unexpected result: %v", got)
	}
}

// Two nodes with different catalogues must return DIFFERENT data (proves a real remote call).
func TestWeek05_RemoteDiffersFromLocal(t *testing.T) {
	addrA := startTestServer(t, "node-a", catalogWith(t, "alpha.txt"))
	addrB := startTestServer(t, "node-b", catalogWith(t, "notes.txt", "diagram.txt"))

	b, err := Fetch(context.Background(), addrB, "node-a", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if b.GetNodeId() != "node-b" || len(b.GetFiles()) != 2 ||
		b.Files[0].GetName() != "diagram.txt" || b.Files[1].GetName() != "notes.txt" {
		t.Fatalf("wrong data from B: %v", b)
	}
	a, err := Fetch(context.Background(), addrA, "node-b", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if a.GetNodeId() != "node-a" || len(a.GetFiles()) != 1 || a.Files[0].GetName() != "alpha.txt" {
		t.Fatalf("wrong data from A: %v", a)
	}
}

func TestWeek05_InvalidRequesterIDs(t *testing.T) {
	addr := startTestServer(t, "node-b", catalogWith(t, "x.txt"))
	for _, id := range []string{"", "bad id!", strings.Repeat("x", 33)} {
		_, err := Fetch(context.Background(), addr, id, time.Second)
		if status.Code(err) != codes.InvalidArgument {
			t.Errorf("requester %q: want InvalidArgument, got %v", id, err)
		}
	}
}

func TestWeek05_FetchRejectsBadAddress(t *testing.T) {
	for _, addr := range []string{"", "not-an-address", "8.8.8.8:9"} {
		_, err := Fetch(context.Background(), addr, "node-a", time.Second)
		if status.Code(err) != codes.InvalidArgument {
			t.Errorf("address %q: want InvalidArgument, got %v", addr, err)
		}
	}
}