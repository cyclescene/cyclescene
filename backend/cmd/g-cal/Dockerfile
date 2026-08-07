FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/g-cal ./cmd/g-cal
COPY internal/scraper ./internal/scraper

RUN CGO_ENABLED=0 GOOS=linux go build -o g-cal ./cmd/g-cal

FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/g-cal ./g-cal

RUN adduser -D appuser && chown appuser:appuser ./g-cal
USER appuser

ENTRYPOINT ["./g-cal"]
