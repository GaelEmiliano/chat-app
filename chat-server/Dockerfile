# syntax=docker/dockerfile:1

FROM golang:1.23.0-bookworm AS builder

WORKDIR /src

# Leverage Docker layer caching for dependencies.
COPY go.mod ./
RUN go mod download

# Copy the rest of the source.
COPY . ./

# Build a small, reproducible binary.
# CGO is disabled to avoid runtime libc dependencies.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/chat-server ./cmd/chat-server

# ----------------------------

FROM debian:bookworm-slim AS runtime

RUN useradd --create-home --shell /usr/sbin/nologin --uid 10001 chatserver

WORKDIR /app

COPY --from=builder /out/chat-server /app/chat-server

USER chatserver

# Default: listen on all interfaces port 8080 (override via env).
ENV CHAT_SERVER_ADDR=:8080

EXPOSE 8080

ENTRYPOINT ["/app/chat-server"]
