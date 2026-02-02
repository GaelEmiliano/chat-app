# Chat App

# Chat App

⚠ This project implements a strict academic protocol. Any protocol violation results in immediate client disconnection.

**Chat App** is a simple, robust TCP chat application built as a small client/server system. The server is written in Go (Go 1.23.0) and enforces a strict JSON protocol over newline (`\n`) framing. The client is written in C and is designed for interactive terminal use. The system supports multiple concurrent clients, user identification, status updates, private messaging, public broadcast messaging, and room-based messaging with invitations and membership rules.

This repository is intentionally boring in the best way: predictable behavior, explicit errors, strict input validation, and protocol correctness over “nice-to-have” features. If the client sends malformed JSON, incomplete messages, invalid values, or violates protocol state (for example sending messages before identifying), the server fails closed exactly as specified: it responds with an INVALID response and disconnects the client. The server’s shared state is owned by a single hub goroutine to avoid data races and keep concurrency reasoning simple and reliable.

The easiest way to run everything is with Docker and docker compose. Manual builds are also possible (Go 1.23.0 for the server, a Debian-like build environment for the client), but Docker is the recommended path for a clean, reproducible setup.

---

## Get the source

Git is the normal way to work with this project.

```sh
$ git clone <YOUR_GITHUB_REPO_URL> chat-app
$ cd chat-app
```

If you prefer to download a ZIP archive from GitHub, you can, but you lose history and it is harder to update cleanly. Use git unless you have a reason not to.

---

## Run with Docker (recommended)

✅ The recommended way to run this project is using [Docker](https://docs.docker.com/) and [docker compose](https://docs.docker.com/compose/).

This runs the server as a long-lived container and lets you launch as many interactive clients as you want (one client per terminal). The compose setup creates an isolated network; the server is reachable by the hostname `server` from client containers.

From `chat-app` directory run:

```sh
$ docker compose build
```

Start the server in the background.

```sh
$ docker compose up -d server
``` 

Follow server logs.

```sh
$ docker compose logs -f server
```

Run a client interactively (one per terminal).

```sh
$ docker compose run --rm client server 8080
``` 

Run a second client (in another terminal).

```sh
$ docker compose run --rm client server 8080
```

Stop the server.

```sh
$ docker compose down
```

Notes:
- The server listens on port 8080 by default and is published to the host via compose. If you want a different port, change the mapping and `CHAT_SERVER_ADDR` in `docker-compose.yml` consistently.
- The client connects to `server:8080` when started via compose because `server` is the compose service name.

---

## Client commands (interactive usage)

The client expects the server host and port on startup, then you interact with it via commands. The exact user experience is intentionally minimal and terminal-friendly.

Start the client (local machine, server running on localhost:8080).

```sh
$ ./chat-client 127.0.0.1 8080
```

Start the client (docker compose, server service on 8080).

 ```sh
$ docker compose run --rm client server 8080
```

Typical command flow inside the client:
- identify first (required), then use features like status, users, messaging, and rooms.

For the full list of supported client commands and their exact syntax, see:

- [Client Guide](chat-client/docs/README.md).

---

## Server behavior and protocol correctness

The server is strict by design. If a message is incomplete (missing required fields), has invalid values (for example invalid status), or cannot be recognized (not a JSON object with a string `type` field), the server replies with:

{ "type": "RESPONSE", "operation": "INVALID", "result": "INVALID" }

and disconnects the client.

This is not “unfriendly”; it is deterministic protocol enforcement. It keeps state consistent and makes debugging easier because incorrect behavior is rejected immediately rather than being tolerated and turning into undefined behavior later.

For server-specific configuration, build/run details, and implementation notes, see:

- [Server Guide](chat-server/docs/README.md).

---

## Project layout

- chat-server/
  Go server (Go 1.23.0). Strict newline-framed JSON protocol enforcement.

- chat-client/
  C client (Meson/Ninja build). Interactive "pretty" terminal client.

- docker-compose.yml
  Orchestrates server and client containers for local testing and reproducible runs.

---

If you want details, read the subproject documentation. If you want to run it, use docker compose. That’s it.

---

## Version

📦 **Chat App v1.0.0**

This version implements the complete project protocol as specified, with no extensions and strict validation.
Future changes, if any, will focus on hardening, testing, and documentation—not protocol changes.

---

## License

📝 **MIT License**

This project is released under the MIT License.
You are free to use, modify, and redistribute it for academic or personal purposes.

See the `LICENSE` file for full details.

- [LICENSE](LICENSE)

---

## Issues and Contact

🐛 Found a bug, unexpected behavior, or protocol issue?
📬 Please report it by opening an issue in the repository or by contacting:

**Email:** emiliano.arreguin@ciencias.unam.mx

When reporting issues, include:
- whether you are running via Docker or manually
- server and client logs if possible
- a short description of what you expected vs what happened

Clear reports help improve the project for everyone 🙂
