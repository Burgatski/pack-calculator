package storage

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

var defaultPackSizes = []int{250, 500, 1000, 2000, 5000}

type Memory struct {
	mu    sync.RWMutex
	sizes []int
}

func NewMemory() *Memory {
	sizes := make([]int, len(defaultPackSizes))
	copy(sizes, defaultPackSizes)
	return &Memory{sizes: sizes}
}

// Returns a sorted copy, safe to mutate by the caller.
func (m *Memory) GetPackSizes() []int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]int, len(m.sizes))
	copy(out, m.sizes)
	return out
}

// Validates and replaces all pack sizes atomically.
func (m *Memory) SetPackSizes(sizes []int) error {
	if len(sizes) == 0 {
		return errors.New("at least one pack size is required")
	}
	for _, s := range sizes {
		if s <= 0 {
			return fmt.Errorf("pack sizes must be positive, got %d", s)
		}
	}

	sorted := make([]int, len(sizes))
	copy(sorted, sizes)
	sort.Ints(sorted)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.sizes = sorted
	return nil
}