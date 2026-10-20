# Klyniva backend image
# Build: docker build -t klyniva/backend ./be

# ---- Build stage ----------------------------------------------------------
FROM golang:1.27-alpine AS builder
WORKDIR /src

# Download modules first so they are cached between source changes
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Static binaries (no cgo) so they run on distroless/static
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate \
 && mkdir -p /out/uploads

# ---- Runtime stage --------------------------------------------------------
# distroless: no shell or package manager; runs as uid 65532 (nonroot)
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/server /out/migrate /app/
COPY --from=builder --chown=65532:65532 /out/uploads /app/uploads

ENV PORT=9000 \
    GIN_MODE=release \
    UPLOAD_PATH=/app/uploads
EXPOSE 9000
USER 65532:65532

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD ["/app/server", "-healthcheck"]

ENTRYPOINT ["/app/server"]
