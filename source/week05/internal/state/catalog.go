package state

import (
	"errors"
	"example.edu/6cc545/cw1/internal/model"
	"sort"
	"sync"
)

// Catalog contains metadata, not mutable file contents. File I/O belongs in CW2.
type Catalog struct {
	mu    sync.RWMutex
	files []model.FileMetadata
}

func NewCatalog() *Catalog { return &Catalog{files: []model.FileMetadata{}} }
func (c *Catalog) Replace(input []model.FileMetadata) error {
	// TODO W03_CATALOG_REPLACE: implement the contract described in TASKS.md.
	return errors.New("TODO W03_CATALOG_REPLACE")
}
func (c *Catalog) Snapshot() []model.FileMetadata {
	// TODO W03_CATALOG_SNAPSHOT: implement the contract described in TASKS.md.
	return nil
}

var _ = errors.New
var _ = sort.Slice
