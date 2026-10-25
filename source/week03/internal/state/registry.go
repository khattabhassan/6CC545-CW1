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

// Add validates first, then checks for a duplicate and inserts under ONE write lock.
func (s *Registry) Add(p model.Peer) error {
	if err := p.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.peers[p.ID]; exists {
		return ErrDuplicate
	}
	s.peers[p.ID] = p
	return nil
}

// Snapshot returns an independent copy sorted by ID.
func (s *Registry) Snapshot() []model.Peer {
	s.mu.RLock()
	out := make([]model.Peer, 0, len(s.peers))
	for _, p := range s.peers {
		out = append(out, p)
	}
	s.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
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