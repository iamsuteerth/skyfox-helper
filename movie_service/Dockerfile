FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && \
    go mod verify

COPY *.go ./
COPY movies.json ./
COPY aphorisms.txt ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /app/bin/movie-service

FROM alpine:3.19

RUN addgroup -S appgroup && \
    adduser -S -G appgroup appuser && \
    apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /app/bin/movie-service .
COPY --from=builder --chown=appuser:appgroup /app/movies.json .
COPY --from=builder --chown=appuser:appgroup /app/aphorisms.txt .

RUN chmod +x /app/movie-service

ENV GIN_MODE=release \
    PORT=4567

USER appuser

EXPOSE 4567

CMD ["./movie-service"]
