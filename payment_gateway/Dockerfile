FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && \
    go mod verify

COPY server/*.go ./server/
COPY types/*.go ./types/
COPY validator/*.go ./validator/
COPY processor/*.go ./processor/ 

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /app/bin/payment-gateway \
    ./server

FROM alpine:3.21

RUN addgroup -S appgroup && \
    adduser -S -G appgroup appuser && \
    apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /app/bin/payment-gateway .

RUN chmod +x /app/payment-gateway

ENV PORT=8082 \
    LOG_LEVEL=info \
    GIN_MODE=release \
    APP_VERSION=prod \
    API_KEY=""

USER appuser

EXPOSE 8082

HEALTHCHECK --interval=60s --timeout=3s --start-period=5s --retries=3 CMD wget -qO- http://localhost:8082/pshealth || exit 1

CMD ["./payment-gateway"]
