# syntax=docker/dockerfile:1

FROM debian:bookworm-slim AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    meson \
    ninja-build \
    pkg-config \
    libjansson-dev \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /src

COPY meson.build ./
COPY src/ ./src/

RUN meson setup build -Dbuildtype=release \
 && meson compile -C build

# ----------------------------

FROM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
    libjansson4 \
 && rm -rf /var/lib/apt/lists/*

RUN useradd -r -u 10001 -g root appuser

WORKDIR /app

COPY --from=builder /src/build/chat-client /app/chat-client

USER appuser

ENTRYPOINT ["/app/chat-client"]
