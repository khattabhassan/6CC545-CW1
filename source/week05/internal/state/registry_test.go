package state

import (
	"errors"
	"example.edu/6cc545/cw1/internal/model"
	"fmt"
	"sync"
	"testing"
)

func peer(id string) model.Peer {
	return model.Peer{ID: id, HTTPAddress: "127.0.0.1:8082", GRPCAddress: "127.0.0.1:9092", Status: "unknown"}
}
func TestWeek03_Registry(t *testing.T) {
	s := NewRegistry()
	for _, id := range []string{"z", "a"} {
		if e := s.Add(peer(id)); e != nil {
			t.Fatal(e)
		}
	}
	if !errors.Is(s.Add(peer("a")), ErrDuplicate) {
		t.Fatal("duplicate policy must be explicit")
	}
	p := s.Snapshot()
	if len(p) != 2 || p[0].ID != "a" {
		t.Fatal("snapshot must sort by ID")
	}
	p[0].ID = "changed"
	if s.Snapshot()[0].ID != "a" {
		t.Fatal("snapshot aliases live state")
	}
}
func TestWeek03_ConcurrentRegistry(t *testing.T) {
	s := NewRegistry()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if e := s.Add(peer(fmt.Sprintf("n%02d", i))); e != nil {
				t.Error(e)
			}
			_ = s.Snapshot()
		}(i)
	}
	wg.Wait()
	if n := len(s.Snapshot()); n != 50 {
		t.Fatalf("want 50 peers, got %d", n)
	}
}
func TestWeek03_InvalidPeer(t *testing.T) {
	s := NewRegistry()
	p := peer("x")
	p.HTTPAddress = "broken"
	if s.Add(p) == nil {
		t.Fatal("invalid address accepted")
	}
	if len(s.Snapshot()) != 0 {
		t.Fatal("invalid input changed state")
	}
}
