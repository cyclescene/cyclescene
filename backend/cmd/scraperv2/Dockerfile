# 1. BUILD STAGE
FROM golang:latest AS builder

# Set up the working directory
WORKDIR /app

# Copy go.mod and go.sum (from the root) to handle dependencies first
COPY go.mod go.sum ./

# Download dependencies (cached layer if go.mod doesn't change)
RUN go mod download

# Copy only the code this service needs
COPY cmd/scraperv2 ./cmd/scraperv2
COPY internal/scraper ./internal/scraper
COPY internal/routes ./internal/routes

RUN CGO_ENABLED=0 GOOS=linux go build -v -o scraper ./cmd/scraperv2

# --- 2. FINAL IMAGE STAGE (Minimal runtime) ---
FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN apk --no-cache add tzdata

WORKDIR /app

COPY --from=builder /app/scraper .

RUN adduser -D appuser
RUN chown appuser:appuser ./scraper

USER appuser

ENTRYPOINT ["./scraper"]
