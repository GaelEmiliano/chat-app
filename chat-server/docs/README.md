# Chat Server

Robust multi-client TCP chat server written in Go (Go 1.23.0).
Implements the full JSON-based chat protocol specified by the project, using newline (`\n`) framing, strict validation, and authoritative server-side state management.

This is a **server only** implementation.
All protocol authority (users, statuses, private messaging, rooms, invitations, and permissions) is enforced exclusively by the server.

The implementation follows a clean, single-owner concurrency model and fails closed on all protocol violations, exactly as required by the specification.

___

## Protocol Scope

**This server implements the protocol exactly as specified; no extensions, shortcuts, or assumptions are made.**

Invalid JSON, malformed messages, unexpected fields, invalid state transitions, or protocol misuse result in the required `INVALID` response followed by client disconnection.

## Features

- Concurrent TCP server using Go standard library
- Newline (`\n`) message framing
- Strict JSON parsing and validation (`encoding/json`)
- Deterministic protocol enforcement (fail-closed)
- User identification and status management
- Private and public messaging
- Room lifecycle management:
  - creation, invitation, join, leave
  - automatic room cleanup when empty
- Correct broadcast semantics (`NEW_USER`, `LEFT_ROOM`, `DISCONNECTED`, etc.)
- Graceful and abrupt disconnection handling
- Single-goroutine hub architecture (no data races)
- Dockerized build and runtime
- No protocol deviations or undocumented behavior

## Requirements

### Recommended (Docker)
- Docker ≥ 20.x

See:
- [Docker](https://www.docker.com/get-started/).

### Manual build (not recommended)
- GNU/Linux (or similar UNIX system).
- Go **1.23.0**

See:
- [Go](https://go.dev/doc/go1.23)

No external Go dependencies are used.

## Build and Run

### Docker (Recommended)

#### Build image
From the `chat-server` directory:

```sh
$ docker build -t chat-server:0.1.0 .
```

#### Run server

``` sh
$ docker run --rm -p 8080:8080 chat-server:0.1.0
```

To change the listening port:

``` sh
$ $ docker run --rm \
-e CHAT_SERVER_ADDR=:<port> \
-p <port>:<port> \
chat-server:0.1.0
```

The server listens on all interfaces (`0.0.0.0`) inside the container.

### Manual build (Go toolchain)

Only use this if Docker is not available 
(I don't think that happen if it's a proper Docker installation).

#### Build

``` sh
$ go build -o chat-server ./cmd/chat-server
```

#### Run

``` sh
$ ./chat-server
``` 

Change the listening address or port using an environment variable:

``` sh
$ CHAT_SERVER_ADDR=:9000 ./chat-server
```

## Configuration

The server is configured exclusively via environment variables.

- CHAT_SERVER_ADDR
  Listening address and port.
  Default: :8080
  
Example:

``` sh
$ CHAT_SERVER_ADDR=0.0.0.0:5555
```

## Concurrency model

The server uses a single-owner hub design:

- All shared state (users, rooms, memberships) is owned by one goroutine.
- Network I/O is isolated per client.
- Communication between components happens through typed channels.
- No locks, no shared mutable state, no data races.

This design guarantees correctness under concurrency and simplifies protocol reasoning.

## Notes

The server does not echo events back to the sender unless explicitly required by the protocol. All disconnections (explicit or abrupt) trigger the correct protocol notifications. The server is suitable for local testing, Docker-based deployments, and academic evaluation.

📝 **MIT License**

This project is released under the MIT License.
You are free to use, modify, and redistribute it for academic or personal purposes.

See the `LICENSE` file for full details.

[LICENSE](../../LICENSE).

## Status

📦 ***Stable - v0.1.0**

## Issues and Contact

🐛 Found a bug, unexpected behavior, or protocol issue?
📬 Please report it by opening an issue in the repository or by contacting:

**Email:** emiliano.arreguin@ciencias.unam.mx

When reporting issues, include:
- whether you are running via Docker or manually
- server and client logs if possible
- a short description of what you expected vs what happened

Clear reports help improve the project 🙂
