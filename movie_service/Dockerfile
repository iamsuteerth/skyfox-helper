FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && \
    go mod verify

COPY server/*.go ./server/
COPY internal/models/*.go ./internal/models/
COPY internal/services/*.go ./internal/services/
COPY data/*.json ./data/

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /app/bin/movie-service \
    ./server

FROM alpine:3.21

RUN addgroup -S appgroup && \
    adduser -S -G appgroup appuser && \
    apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /app/bin/movie-service .
COPY --from=builder --chown=appuser:appgroup /app/data ./data

RUN chmod +x /app/movie-service

ENV PORT=4567 \
    MOVIES_DATA_PATH=/app/data/movies.json \
    LOG_LEVEL=info \
    GIN_MODE=release \
    APP_VERSION=prod \
    API_KEY=""

USER appuser

EXPOSE 4567

HEALTHCHECK --interval=60s --timeout=3s --start-period=5s --retries=3 CMD wget -qO- http://localhost:4567/mshealth || exit 1

CMD ["./movie-service"]
