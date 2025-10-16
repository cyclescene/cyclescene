# 1. BUILD STAGE
FROM golang:latest AS builder

# Set up the working directory
WORKDIR /app

# Copy go.mod and go.sum (from the root) to handle dependencies first
COPY go.mod go.sum ./

# Download dependencies (cached layer if go.mod doesn't change)
RUN go mod download

# Copy only the code this service needs
COPY cmd/db-backups ./cmd/db-backups
# Add any internal packages if db-backups uses them
# COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -v -o backup ./cmd/db-backups

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN apk --no-cache add tzdata

WORKDIR /app

COPY --from=builder /app/backup .

RUN adduser -D appuser
RUN chown appuser:appuser ./backup

USER appuser

ENTRYPOINT ["./backup"]
