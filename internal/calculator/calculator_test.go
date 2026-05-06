package calculator_test

import (
	"testing"

	"github.com/Burgatski/pack-calculator/internal/calculator"
)

var defaultSizes = []int{250, 500, 1000, 2000, 5000}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name       string
		order      int
		sizes      []int
		wantPacks  map[int]int
		wantItems  int
		wantErr    bool
	}{
		{
			name:      "1_item_returns_1x250",
			order:     1,
			sizes:     defaultSizes,
			wantPacks: map[int]int{250: 1},
			wantItems: 250,
		},
		{
			name:      "250_items_returns_1x250",
			order:     250,
			sizes:     defaultSizes,
			wantPacks: map[int]int{250: 1},
			wantItems: 250,
		},
		{
			name:      "251_items_returns_1x500_not_2x250",
			order:     251,
			sizes:     defaultSizes,
			wantPacks: map[int]int{500: 1},
			wantItems: 500,
		},
		{
			name:      "501_items_returns_1x500_1x250",
			order:     501,
			sizes:     defaultSizes,
			wantPacks: map[int]int{500: 1, 250: 1},
			wantItems: 750,
		},
		{
			name:      "12001_items_returns_2x5000_1x2000_1x250",
			order:     12001,
			sizes:     defaultSizes,
			wantPacks: map[int]int{5000: 2, 2000: 1, 250: 1},
			wantItems: 12250,
		},
		{
			name:      "500000_items_with_sizes_23_31_53",
			order:     500000,
			sizes:     []int{23, 31, 53},
			wantPacks: map[int]int{53: 9429, 31: 7, 23: 2},
			wantItems: 500000,
		},
		{
			name:      "500_exact",
			order:     500,
			sizes:     defaultSizes,
			wantPacks: map[int]int{500: 1},
			wantItems: 500,
		},
		{
			name:      "5000_exact",
			order:     5000,
			sizes:     defaultSizes,
			wantPacks: map[int]int{5000: 1},
			wantItems: 5000,
		},
		{
			name:      "10000_exact",
			order:     10000,
			sizes:     defaultSizes,
			wantPacks: map[int]int{5000: 2},
			wantItems: 10000,
		},
		{
			name:      "single_size_exact_match",
			order:     100,
			sizes:     []int{100},
			wantPacks: map[int]int{100: 1},
			wantItems: 100,
		},
		{
			name:      "single_size_round_up",
			order:     101,
			sizes:     []int{100},
			wantPacks: map[int]int{100: 2},
			wantItems: 200,
		},
		{
			name:    "zero_order_returns_error",
			order:   0,
			sizes:   defaultSizes,
			wantErr: true,
		},
		{
			name:    "negative_order_returns_error",
			order:   -5,
			sizes:   defaultSizes,
			wantErr: true,
		},
		{
			name:    "empty_sizes_returns_error",
			order:   100,
			sizes:   []int{},
			wantErr: true,
		},
		{
			name:    "negative_pack_size_returns_error",
			order:   100,
			sizes:   []int{-10, 500},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := calculator.Calculate(tc.order, tc.sizes)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.TotalItems != tc.wantItems {
				t.Errorf("TotalItems = %d, want %d", result.TotalItems, tc.wantItems)
			}
			if len(result.Packs) != len(tc.wantPacks) {
				t.Errorf("Packs = %v, want %v", result.Packs, tc.wantPacks)
				return
			}
			for size, qty := range tc.wantPacks {
				if result.Packs[size] != qty {
					t.Errorf("Packs[%d] = %d, want %d (full result: %v)", size, result.Packs[size], qty, result.Packs)
				}
			}
		})
	}
}

// TestCalculateTotalPacksConsistency verifies that TotalPacks equals the sum
// of all quantities in the Packs map.
func TestCalculateTotalPacksConsistency(t *testing.T) {
	result, err := calculator.Calculate(12001, defaultSizes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sum := 0
	for _, qty := range result.Packs {
		sum += qty
	}
	if sum != result.TotalPacks {
		t.Errorf("sum of pack quantities (%d) != TotalPacks (%d)", sum, result.TotalPacks)
	}
}

// TestCalculateTotalItemsConsistency verifies that TotalItems equals the
// sum of (size × quantity) across all packs.
func TestCalculateTotalItemsConsistency(t *testing.T) {
	result, err := calculator.Calculate(501, defaultSizes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	total := 0
	for size, qty := range result.Packs {
		total += size * qty
	}
	if total != result.TotalItems {
		t.Errorf("computed items (%d) != TotalItems (%d)", total, result.TotalItems)
	}
}
