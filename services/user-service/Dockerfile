# User Service Dockerfile
FROM golang:1.22-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
COPY shared/go.mod shared/go.sum ./shared/
COPY services/user-service/go.mod services/user-service/go.sum ./services/user-service/

# Download dependencies
RUN go mod download

# Copy source code
COPY shared/ ./shared/
COPY services/user-service/ ./services/user-service/

# Build the application
WORKDIR /app/services/user-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o user-service \
    ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/services/user-service/user-service .

# Copy configuration files
COPY --from=builder /app/configs/development/ ./configs/

# Change ownership to appuser
RUN chown -R appuser:appuser /app
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# Expose port
EXPOSE 8082

# Run the application
CMD ["./user-service"]