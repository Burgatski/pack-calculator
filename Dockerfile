# ── Stage 1: Build ───────────────────────────────────────────────────────────
# Uses the full Go toolchain to compile a static, self-contained binary.
# Separating build and runtime stages keeps the final image under 15 MB.
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Download dependencies before copying source so this layer is cached
# as long as go.mod / go.sum are unchanged.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the full source tree and build.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o pack-calculator ./cmd/server

# ── Stage 2: Runtime ─────────────────────────────────────────────────────────
# Minimal Alpine with TLS certificates only; no Go toolchain in the image.
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/pack-calculator .

# PORT is the only runtime configuration the server reads.
ENV PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:${PORT}/health || exit 1

CMD ["./pack-calculator"]
