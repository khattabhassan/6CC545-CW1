package state

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestWeek03_SameIDExactlyOneWins(t *testing.T) {
	s := NewRegistry()
	const writers = 50
	var ok, dup int32
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start // release all goroutines together
			switch err := s.Add(peer("same")); {
			case err == nil:
				atomic.AddInt32(&ok, 1)
			case errors.Is(err, ErrDuplicate):
				atomic.AddInt32(&dup, 1)
			default:
				t.Error(err)
			}
		}()
	}
	close(start)
	wg.Wait()
	if ok != 1 || dup != writers-1 {
		t.Fatalf("want 1 success and %d duplicates, got %d and %d", writers-1, ok, dup)
	}
	if n := len(s.Snapshot()); n != 1 {
		t.Fatalf("want 1 peer, got %d", n)
	}
}