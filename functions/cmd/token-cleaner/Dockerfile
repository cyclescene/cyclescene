# 1. BUILD STAGE
FROM golang:latest AS builder

# Set up the working directory
WORKDIR /app

# Copy go.mod and go.sum (from the root) to handle dependencies first
COPY go.mod go.sum ./

# Download dependencies (cached layer if go.mod doesn't change)
RUN go mod download

# Copy only the code this service needs
COPY cmd/token-cleaner ./cmd/token-cleaner
# Add any internal packages if token-cleaner uses them
# COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -v -o cleaner ./cmd/token-cleaner

# --- 2. FINAL IMAGE STAGE (Minimal runtime) ---
FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN apk --no-cache add tzdata

WORKDIR /app

COPY --from=builder /app/cleaner .

RUN adduser -D appuser
RUN chown appuser:appuser ./cleaner

USER appuser

ENTRYPOINT ["./cleaner"]
