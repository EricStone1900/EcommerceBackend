# Stage 1: Build
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /build/server ./cmd/server

# Stage 2: Run
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /build/server .
COPY --from=builder /build/configs ./configs
COPY --from=builder /build/migrations ./migrations

USER appuser

EXPOSE 8080

ENTRYPOINT ["./server"]
CMD ["serve"]
