# Build stage
FROM golang:1.24.0-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates build-base libwebp-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/image-optimizer/ ./cmd/image-optimizer/
COPY internal/imageprocessing/ ./internal/imageprocessing

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build \
    -o image-optimizer ./cmd/image-optimizer

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates libwebp

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/image-optimizer .

# Create non-root user
RUN addgroup -g 1000 -S appuser && adduser -u 1000 -S appuser -G appuser

USER appuser

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./image-optimizer"]
