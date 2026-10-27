package state

import (
	"errors"
	"example.edu/6cc545/cw1/internal/model"
	"sort"
	"sync"
)

var ErrDuplicate = errors.New("duplicate peer ID")
var ErrMissing = errors.New("peer not found")

type Registry struct {
	mu    sync.RWMutex
	peers map[string]model.Peer
}

func NewRegistry() *Registry { return &Registry{peers: make(map[string]model.Peer)} }
func (s *Registry) Add(p model.Peer) error {
	// TODO W03_ADD: implement the contract described in TASKS.md.
	return errors.New("TODO W03_ADD")
}
func (s *Registry) Snapshot() []model.Peer {
	// TODO W03_SNAPSHOT: implement the contract described in TASKS.md.
	return nil
}
func (s *Registry) Get(id string) (model.Peer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.peers[id]
	if !ok {
		return p, ErrMissing
	}
	return p, nil
}

var _ = sort.Slice
