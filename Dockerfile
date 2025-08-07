# Build stage
FROM golang:1.25rc2-alpine3.22 AS builder

# Install git and ca-certificates (required for fetching dependencies)
RUN apk update && apk add --no-cache git ca-certificates tzdata

# Create appuser for non-root execution
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy go mod files first for better Docker layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o server cmd/server/main.go

# Final stage - minimal scratch image
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copy binary from builder
COPY --from=builder /app/server /server

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/server", "-health-check"] || exit 1

# Run binary
ENTRYPOINT ["/server"]