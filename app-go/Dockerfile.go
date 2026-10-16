# Stage 1: Build binary
FROM golang:1.22-alpine AS builder

WORKDIR /build

ENV GOSUMDB=off
RUN go env -w GOSUMDB=off

COPY go.mod ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN go build -ldflags="-w -s" -o benchmark-go ./cmd/server/main.go

# Stage 2: Minimal runtime image (Security Hardened Non-Root User)
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

# Create non-root system user and group
RUN addgroup -g 1000 appgroup && adduser -u 1000 -G appgroup -D appuser

WORKDIR /app
COPY --from=builder /build/benchmark-go /app/benchmark-go
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/benchmark-go"]
