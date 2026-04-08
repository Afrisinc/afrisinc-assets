# ╔══════════════════════════════════════════════════════════════╗
# ║  Stage 1 — builder                                          ║
# ║  Full Go toolchain, compiles a stripped static binary        ║
# ╚══════════════════════════════════════════════════════════════╝
FROM golang:1.22-alpine AS builder

# Install only what the build needs (git for go mod download, ca-certs for TLS)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy dependency manifests first so Docker caches the module layer.
# This layer only re-downloads when go.mod or go.sum change.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the source
COPY . .

# Build a fully static binary:
#   CGO_ENABLED=0  → no libc dependency, runs in scratch
#   -ldflags        → strip debug symbols and DWARF (-s -w)
#   -trimpath       → remove local build paths from binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -trimpath \
    -o /build/bin/server \
    ./cmd/server


# ╔══════════════════════════════════════════════════════════════╗
# ║  Stage 2 — runtime                                          ║
# ║  Scratch image: zero OS, minimal attack surface             ║
# ╚══════════════════════════════════════════════════════════════╝
FROM scratch AS production

# Pull in certs and timezone data from the builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the compiled binary
COPY --from=builder /build/bin/server /server

# Non-root user (scratch has no useradd, so we set UID numerically)
USER 1001:1001

# Uploads volume mount point
VOLUME ["/data/uploads"]

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/server", "-health"] 

ENTRYPOINT ["/server"]


# ╔══════════════════════════════════════════════════════════════╗
# ║  Stage 3 — development (hot-reload with Air)                ║
# ╚══════════════════════════════════════════════════════════════╝
FROM golang:1.22-alpine AS development

RUN apk add --no-cache git curl tzdata ca-certificates && \
    go install github.com/air-verse/air@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Source is mounted at runtime via docker-compose volume
EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
