# Pack Calculator

A Go HTTP service that calculates the optimal pack combination needed to fulfil a customer order.

## Problem Statement

Given a set of configurable pack sizes and an order quantity **N**, find a combination of whole packs such that:

1. **Minimum total items** ≥ N (packs cannot be broken open).
2. **Minimum pack count** among solutions with equal item counts.

## Algorithm

This is a variant of the **unbounded coin-change** problem, solved with bottom-up dynamic programming.

- `dp[i]` = minimum packs needed to reach exactly `i` items.
- We compute `dp[0 … N + max(packSizes)]`, then scan forward from `N` to find the smallest reachable `T ≥ N`.
- The pack breakdown is reconstructed by following parent pointers from `T` back to `0`.

**Complexity:** O(N × |packSizes|) time, O(N) space — handles 500 000 items with non-standard pack sizes (e.g. `[23, 31, 53]`) in milliseconds.
