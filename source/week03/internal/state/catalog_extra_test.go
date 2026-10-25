package state

import (
	"example.edu/6cc545/cw1/internal/model"
	"strings"
	"testing"
)

func TestWeek03_CatalogDuplicateNameKeepsOld(t *testing.T) {
	c := NewCatalog()
	h := strings.Repeat("a", 64)
	if err := c.Replace([]model.FileMetadata{{Name: "keep.txt", Size: 1, SHA256: h}}); err != nil {
		t.Fatal(err)
	}
	dup := []model.FileMetadata{
		{Name: "x.txt", Size: 1, SHA256: h},
		{Name: "x.txt", Size: 2, SHA256: h},
	}
	if c.Replace(dup) == nil {
		t.Fatal("duplicate names accepted")
	}
	got := c.Snapshot()
	if len(got) != 1 || got[0].Name != "keep.txt" {
		t.Fatalf("rejected replacement changed state: %v", got)
	}
}