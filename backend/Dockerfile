# syntax=docker/dockerfile:1.7
# ----------------------------------------------------------------------
# LogiTrack backend — multi-stage image
# Stage 1: build a statically linked binary with Go 1.26 on Alpine.
# Stage 2: copy the binary into a distroless base that ships no shell
# and no package manager, dramatically reducing attack surface.
# ----------------------------------------------------------------------

FROM golang:1.26-alpine AS build
ARG VERSION=dev
WORKDIR /src

RUN apk add --no-cache git ca-certificates tzdata

# Leverage the layer cache: copy modules first, download, then source.
COPY go.mod go.sum ./
RUN go mod download || true

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /out/logitrack-server \
    ./cmd/server

# ----------------------------------------------------------------------
# Runtime image. Distroless/static includes ca-certs + tzdata; non-root
# UID 65532 is predefined ("nonroot"). No shell → HEALTHCHECK uses wget
# via `busybox:nonroot`-style helper is not available, so we exec the
# binary itself with a flag in future versions. For now we rely on
# Kubernetes liveness probes hitting /api/health.
# ----------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=build /out/logitrack-server /app/logitrack-server

USER nonroot:nonroot
EXPOSE 8080

# Distroless has no shell, so HEALTHCHECK is commented — Compose and
# Kubernetes check the /api/health endpoint over HTTP directly.
# HEALTHCHECK --interval=30s --timeout=5s --retries=3 CMD ["/app/logitrack-server", "--healthcheck"]

ENTRYPOINT ["/app/logitrack-server"]
