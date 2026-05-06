// Package calculator finds the minimum-items, minimum-packs combination
// of whole packs needed to cover an order.
//
// Greedy doesn't work for arbitrary pack sizes (e.g. [23, 31, 53] with
// order=500000 gives the wrong answer), so we use bottom-up DP — essentially
// the unbounded coin-change problem:
//
//	dp[i]     = fewest packs to hit exactly i items
//	parent[i] = which pack size was used on the last step to reach i
//
// We fill dp[0..N+max(size)], find the smallest reachable T >= N, then
// backtrack through parent[] to recover the actual pack quantities.
//
// Time: O(N × |sizes|)   Space: O(N)
package calculator

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

const unreachable = math.MaxInt32

// Result is the output of Calculate.
type Result struct {
	Packs      map[int]int // pack size → quantity
	TotalItems int         // items actually shipped (>= order)
	TotalPacks int         // number of individual packs
}

// Calculate returns the optimal pack breakdown for the given order.
// It minimises total items shipped first, then total pack count.
func Calculate(order int, packSizes []int) (Result, error) {
	if order <= 0 {
		return Result{}, fmt.Errorf("order must be positive, got %d", order)
	}
	if len(packSizes) == 0 {
		return Result{}, errors.New("pack sizes list is empty")
	}
	for _, p := range packSizes {
		if p <= 0 {
			return Result{}, fmt.Errorf("pack sizes must be positive, got %d", p)
		}
	}

	sizes := make([]int, len(packSizes))
	copy(sizes, packSizes)
	sort.Ints(sizes)
	maxPack := sizes[len(sizes)-1]

	// Worst-case overshoot is min(size)-1, which is always < max(size),
	// so N+max(size) is a safe upper bound for the DP array.
	limit := order + maxPack

	dp := make([]int, limit+1)
	for i := 1; i <= limit; i++ {
		dp[i] = unreachable
	}
	parent := make([]int, limit+1)

	for i := 1; i <= limit; i++ {
		for _, p := range sizes {
			if p > i {
				break // sorted, so nothing larger will fit either
			}
			if dp[i-p] == unreachable {
				continue
			}
			if dp[i-p]+1 < dp[i] {
				dp[i] = dp[i-p] + 1
				parent[i] = p
			}
		}
	}

	// First reachable amount >= order is our target.
	target := -1
	for t := order; t <= limit; t++ {
		if dp[t] != unreachable {
			target = t
			break
		}
	}
	if target == -1 {
		// Can't happen with a non-empty size list, but guard just in case.
		return Result{}, errors.New("no solution found for the given pack sizes")
	}

	// Backtrack through parent pointers to get pack counts.
	packs := make(map[int]int)
	for rem := target; rem > 0; {
		p := parent[rem]
		packs[p]++
		rem -= p
	}

	return Result{
		Packs:      packs,
		TotalItems: target,
		TotalPacks: dp[target],
	}, nil
}
