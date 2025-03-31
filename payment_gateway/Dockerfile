FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && \
    go mod verify

COPY *.go ./
COPY server/*.go ./server/
COPY database/*.go ./database/
COPY types/*.go ./types/
COPY validator/*.go ./validator/

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
    AWS_REGION=ap-south-1 \
    DYNAMODB_TABLE=PendingTransactions \
    LOG_LEVEL=info \
    GIN_MODE=release \
    APP_VERSION=prod

USER appuser

EXPOSE 8082

CMD ["./payment-gateway"]
