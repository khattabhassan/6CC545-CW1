package state

import (
	"example.edu/6cc545/cw1/internal/model"
	"fmt"
	"sort"
	"sync"
)

// Catalog contains metadata, not mutable file contents. File I/O belongs in CW2.
type Catalog struct {
	mu    sync.RWMutex
	files []model.FileMetadata
}

func NewCatalog() *Catalog { return &Catalog{files: []model.FileMetadata{}} }

// Replace is all-or-nothing: validate a private copy first, commit only if everything is valid.
func (c *Catalog) Replace(input []model.FileMetadata) error {
	candidate := make([]model.FileMetadata, len(input))
	copy(candidate, input)
	seen := make(map[string]struct{}, len(candidate))
	for i, f := range candidate {
		if err := f.Validate(); err != nil {
			return fmt.Errorf("file %d (%q): %w", i, f.Name, err)
		}
		if _, dup := seen[f.Name]; dup {
			return fmt.Errorf("duplicate file name %q", f.Name)
		}
		seen[f.Name] = struct{}{}
	}
	sort.Slice(candidate, func(i, j int) bool { return candidate[i].Name < candidate[j].Name })
	c.mu.Lock()
	c.files = candidate
	c.mu.Unlock()
	return nil
}

// Snapshot returns an independent copy, already sorted by name.
func (c *Catalog) Snapshot() []model.FileMetadata {
	c.mu.RLock()
	out := make([]model.FileMetadata, len(c.files))
	copy(out, c.files)
	c.mu.RUnlock()
	return out
}