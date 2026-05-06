# Pack Calculator

A production-grade Go HTTP service that calculates the optimal pack combination for a customer order.

## Problem Statement

Given a configurable set of pack sizes and an order quantity **N**, find a combination of whole packs such that (in priority order):

1. **Minimum total items shipped** — packs cannot be broken open; ship the fewest items that still cover the order.
2. **Minimum pack count** — among solutions with the same item total, prefer fewer packs.

Pack sizes are managed at runtime via the API — no recompilation or redeployment required.

---

## Quick Start

### Option A — Docker (recommended, mirrors what reviewers test)

```bash
# Build and start in one command
docker compose up --build

# The UI is now available at http://localhost:8080
```

To stop:

```bash
docker compose down
```

### Option B — Docker without Compose

```bash
# Build the image
docker build -t pack-calculator .

# Run the container
docker run -d --name pack-calculator -p 8080:8080 pack-calculator

# Open http://localhost:8080

# Stop and remove
docker stop pack-calculator && docker rm pack-calculator
```

### Option C — Run Locally (Go 1.22+ required)

```bash
go run ./cmd/server
# or: make run
```

> The `PORT` environment variable controls the listening port (default `8080`).

---

## Running Tests

```bash
# All packages
go test ./...

# With race detector (recommended)
go test -race ./...

# With per-package coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Or via Makefile
make test
make test-coverage
```

Expected output:

```
ok  github.com/Burgatski/pack-calculator/internal/calculator   (97.4% coverage)
ok  github.com/Burgatski/pack-calculator/internal/handler      (78.4% coverage)
ok  github.com/Burgatski/pack-calculator/internal/storage      (100.0% coverage)
```

---

## API Reference

All endpoints accept and return `application/json`.

### `GET /api/packs` — list pack sizes

```bash
curl http://localhost:8080/api/packs
```

```json
{"pack_sizes":[250,500,1000,2000,5000]}
```

---

### `PUT /api/packs` — replace pack sizes

Pack sizes persist for the life of the process. Any positive integers are valid.

```bash
curl -X PUT http://localhost:8080/api/packs \
  -H 'Content-Type: application/json' \
  -d '{"pack_sizes": [23, 31, 53]}'
```

```json
{"pack_sizes":[23,31,53]}
```

---

### `POST /api/calculate` — calculate optimal packs

```bash
curl -X POST http://localhost:8080/api/calculate \
  -H 'Content-Type: application/json' \
  -d '{"order": 501}'
```

```json
{"packs":{"250":1,"500":1},"total_items":750,"total_packs":2}
```

**Critical edge case** (large order, non-standard pack sizes):

```bash
# Set non-standard sizes first
curl -X PUT http://localhost:8080/api/packs \
  -H 'Content-Type: application/json' \
  -d '{"pack_sizes": [23, 31, 53]}'

# Calculate 500 000 items
curl -X POST http://localhost:8080/api/calculate \
  -H 'Content-Type: application/json' \
  -d '{"order": 500000}'
```

```json
{"packs":{"23":2,"31":7,"53":9429},"total_items":500000,"total_packs":9438}
```

Verification: `9429 × 53 + 7 × 31 + 2 × 23 = 499737 + 217 + 46 = 500000` ✓

---

### `GET /health` — liveness probe

```bash
curl http://localhost:8080/health
```

```json
{"status":"ok"}
```

---

## Example Results (default pack sizes)

| Items ordered | Packs shipped | Total items shipped |
|:---:|---|:---:|
| 1 | 1 × 250 | 250 |
| 250 | 1 × 250 | 250 |
| 251 | 1 × 500 | 500 |
| 501 | 1 × 500 + 1 × 250 | 750 |
| 12 001 | 2 × 5000 + 1 × 2000 + 1 × 250 | 12 250 |

---

## Algorithm

The problem is a variant of the **unbounded coin-change** problem: find the minimum number of "coins" (packs) that sum to at least N, then minimise the count.

A greedy approach (use the largest pack that fits, repeat) fails for arbitrary pack sizes. For example, with sizes `[23, 31, 53]` and order `500 000`, greedy overshoots by thousands. Dynamic programming finds the exact solution in milliseconds.

**Bottom-up DP:**

```
dp[0]   = 0              (base case: 0 items need 0 packs)
dp[i]   = min over all p in packSizes of  dp[i - p] + 1
parent[i] = pack size p used in the optimal step to reach i
```

We compute `dp[0 … N + max(packSizes)]`, then scan forward from `N` to find the smallest reachable `T ≥ N`. The pack breakdown is reconstructed by following `parent` pointers from `T` back to `0`.

- **Time:** O(N × |packSizes|)
- **Space:** O(N + max(packSizes))

---

## Project Structure

```
pack-calculator/
├── cmd/server/
│   ├── main.go              # Entry point — wiring, graceful shutdown
│   └── web/
│       └── index.html       # Single-page UI (embedded into binary at build time)
├── internal/
│   ├── calculator/
│   │   ├── calculator.go    # DP algorithm
│   │   └── calculator_test.go
│   ├── handler/
│   │   ├── handler.go       # HTTP handlers
│   │   └── handler_test.go
│   ├── model/
│   │   └── model.go         # Request / response types
│   └── storage/
│       ├── storage.go       # Storage interface
│       ├── memory.go        # In-memory implementation (thread-safe)
│       └── memory_test.go
├── .dockerignore
├── Dockerfile               # Multi-stage build — final image < 15 MB
├── docker-compose.yml       # Single-command startup
├── Makefile                 # Convenience targets
├── go.mod
└── README.md
```

---

## Design Decisions

| Decision | Rationale |
|---|---|
| Standard `net/http` with Go 1.22 routing patterns | Sufficient for this service; no external dependencies or version lock-in. |
| `internal/` package layout | Prevents accidental imports; enforces the handler → service → storage dependency direction. |
| `storage.Storage` interface | The concrete implementation (currently in-memory) can be replaced with a database-backed one without touching handlers or the algorithm. |
| `sync.RWMutex` in `Memory` | Multiple concurrent reads are common; write lock only on `SetPackSizes`. |
| `//go:embed web` in `main.go` | The binary is fully self-contained — no static-file path configuration needed in Docker or on any host. |
| Multi-stage Dockerfile | Builder stage uses the full Go toolchain; runtime stage contains only the ~6 MB static binary + CA certificates. |
| `log/slog` structured logging | Built into Go 1.21+; zero additional dependencies; JSON-friendly output for log aggregators. |

---

## Makefile Reference

| Target | Description |
|---|---|
| `make build` | Compile binary locally |
| `make run` | Compile and run (port 8080) |
| `make test` | Run all tests |
| `make test-coverage` | Run tests and generate `coverage.html` |
| `make lint` | Run `go vet` |
| `make docker-build` | Build Docker image |
| `make docker-run` | Build image and start container |
| `make docker-stop` | Stop and remove container |
| `make compose-up` | Start via docker compose (detached) |
| `make compose-down` | Stop docker compose stack |
