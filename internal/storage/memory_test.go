package storage_test

import (
	"sync"
	"testing"

	"github.com/Burgatski/pack-calculator/internal/storage"
)

func TestNewMemory_DefaultSizes(t *testing.T) {
	m := storage.NewMemory()
	sizes := m.GetPackSizes()
	if len(sizes) == 0 {
		t.Fatal("expected non-empty default pack sizes")
	}
	// Defaults must be positive and sorted.
	for i, s := range sizes {
		if s <= 0 {
			t.Errorf("sizes[%d] = %d, must be positive", i, s)
		}
		if i > 0 && sizes[i] <= sizes[i-1] {
			t.Errorf("sizes not sorted: sizes[%d]=%d <= sizes[%d]=%d", i, sizes[i], i-1, sizes[i-1])
		}
	}
}

func TestGetPackSizes_ReturnsCopy(t *testing.T) {
	m := storage.NewMemory()
	a := m.GetPackSizes()
	b := m.GetPackSizes()

	// Mutating the returned slice must not affect subsequent reads.
	a[0] = 999999
	if b[0] == 999999 {
		t.Error("GetPackSizes returned a shared reference, not a copy")
	}
	c := m.GetPackSizes()
	if c[0] == 999999 {
		t.Error("mutating returned slice changed internal state")
	}
}

func TestGetPackSizes_IsSorted(t *testing.T) {
	m := storage.NewMemory()
	if err := m.SetPackSizes([]int{5000, 250, 1000, 500, 2000}); err != nil {
		t.Fatalf("SetPackSizes: %v", err)
	}
	sizes := m.GetPackSizes()
	for i := 1; i < len(sizes); i++ {
		if sizes[i] < sizes[i-1] {
			t.Errorf("sizes not sorted at index %d: %v", i, sizes)
		}
	}
}

func TestSetPackSizes_ValidInput(t *testing.T) {
	m := storage.NewMemory()
	want := []int{10, 50, 100}
	if err := m.SetPackSizes(want); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := m.GetPackSizes()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("sizes[%d] = %d, want %d", i, got[i], v)
		}
	}
}

func TestSetPackSizes_RejectsEmpty(t *testing.T) {
	m := storage.NewMemory()
	if err := m.SetPackSizes([]int{}); err == nil {
		t.Error("expected error for empty sizes, got nil")
	}
}

func TestSetPackSizes_RejectsZero(t *testing.T) {
	m := storage.NewMemory()
	if err := m.SetPackSizes([]int{0, 100}); err == nil {
		t.Error("expected error for zero pack size, got nil")
	}
}

func TestSetPackSizes_RejectsNegative(t *testing.T) {
	m := storage.NewMemory()
	if err := m.SetPackSizes([]int{-5, 100}); err == nil {
		t.Error("expected error for negative pack size, got nil")
	}
}

func TestSetPackSizes_RollsBackOnError(t *testing.T) {
	m := storage.NewMemory()
	before := m.GetPackSizes()

	_ = m.SetPackSizes([]int{})

	after := m.GetPackSizes()
	if len(before) != len(after) {
		t.Errorf("store mutated after failed SetPackSizes: before=%v after=%v", before, after)
	}
}


// TestMemoryConcurrency runs concurrent readers and writers to verify that
func TestMemoryConcurrency(t *testing.T) {
	m := storage.NewMemory()
	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			_ = m.GetPackSizes()
		}(i)
		go func(n int) {
			defer wg.Done()
			sizes := []int{n + 1, n + 2, n + 3}
			_ = m.SetPackSizes(sizes)
		}(i)
	}
	wg.Wait()
}
