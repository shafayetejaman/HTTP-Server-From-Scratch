# HTTP Server From Scratch

A minimal HTTP server implemented in Go with low-level handling of HTTP requests and responses — no `net/http` convenience helpers used. This project is a hands-on exploration of how HTTP works under the hood: parsing raw request bytes, extracting headers, building response bytes, and wiring everything to a TCP listener.

**Why this project matters**
- **Low-level HTTP parsing:** Implemented custom parsing and serialization for HTTP requests and responses instead of relying on high-level libraries.
- **Hands-on networking:** Uses a custom TCP listener (`cmd/tcplistener`) and a simple HTTP server (`cmd/httpserver`) to demonstrate the full request → response lifecycle.
- **Tested components:** Unit tests for header parsing and request handling live alongside the implementation.

**Highlights**
- **Manual request parsing:** Read raw bytes from a TCP connection and parse HTTP method, path, version, headers, and body into structured types.
- **Custom response serializer:** Constructed HTTP response bytes (status line, headers, body) manually to control formatting and content-length handling.
- **Separation of concerns:** Clear package layout under `internal/` with `request`, `response`, and `headers` subpackages that encapsulate parsing and serialization logic.
- **Realistic runner:** Example command-line entry points under `cmd/` to run the server and related utilities.

**Project structure**
- `cmd/httpserver` — Example HTTP server using the internal packages.
- `cmd/tcplistener` — TCP listener used to accept raw connections.
- `cmd/udpsender` — A small UDP tool included for experimentation.
- `internal/` — Core implementation
  - `request/` — Request parsing and tests (`request.go`, `request_test.go`)
  - `response/` — Response building and writing (`response.go`)
  - `headers/` — Header parsing helpers and tests (`headers.go`, `headers_test.go`)

**Quick start**
Run the server from the repository root:

```bash
go run ./cmd/httpserver
```

To run the TCP listener directly:

```bash
go run ./cmd/tcplistener
```

Run unit tests:

```bash
go test ./...
```

**Example: what the code does (short)**
- Accepts a raw TCP connection.
- Reads request bytes and splits into request-line, headers, and body.
- Parses headers into a structured map and exposes typed request fields (method, path, version).
- Builds response bytes by composing a status line, header block, and body and writes them back to the connection.

**Resume-ready bullets**
- Implemented a low-level HTTP request parser and response serializer in Go, handling raw TCP connections and manual header/body processing without using the `net/http` high-level APIs.
- Designed and implemented modular packages for `request`, `response`, and `headers`, with unit tests verifying parsing edge cases and header behavior.
- Built a small CLI server and TCP listener to validate the end-to-end request → response flow and demonstrate correct HTTP formatting.

**What I learned / takeaways**
- How HTTP message framing works (request line, headers, CRLF separators, body length handling).
- Practical experience with Go networking primitives and reading/writing raw sockets.
- Importance of careful parsing and defensive handling for real-world protocol implementations.
 - Practical experience with Go networking primitives and reading/writing raw sockets.
 - Importance of careful parsing and defensive handling for real-world protocol implementations.
