# HTTP From TCP

Building an HTTP/1.1 server from scratch using raw TCP sockets. This is course code from Boot.dev, developed in WSL Alpine.

## What's This About

Instead of using Go's `net/http` package like a normal person, this project implements HTTP from the ground up. We're talking raw TCP connections, manually parsing request lines, headers, bodies - the whole thing. It's a great way to actually understand what's happening under the hood when you make an HTTP request.

## Project Structure

```
cmd/
  httpserver/    - The main HTTP server
  tcplistener/   - Simple TCP listener for debugging requests
  udplistener/   - UDP listener (because why not)
  udpsender/     - UDP sender for testing

internal/
  headers/       - HTTP header parsing and management
  request/       - HTTP request parsing (state machine style)
  response/      - HTTP response writing
  server/        - TCP server with connection handling
```

## The HTTP Server

Runs on port 42069 (obviously). Handles a few routes:

- `/` - Returns a success page
- `/yourproblem` - Returns 400 Bad Request (your fault)
- `/myproblem` - Returns 500 Internal Server Error (my fault)
- `/video` - Serves `assets/vim.mp4` if it exists
- `/httpbin/*` - Proxies to httpbin.org with chunked transfer encoding and trailers

The proxy endpoint is the interesting one - it streams responses using chunked transfer encoding and adds SHA256 and content length trailers at the end.

## How It Works

### Request Parsing

The request parser is a state machine that processes incoming bytes:
1. Parse the request line (method, target, HTTP version)
2. Parse headers until we hit the empty line
3. Read the body based on Content-Length

It handles partial reads and buffer management properly, so it works with real TCP connections where data arrives in chunks.

### Response Writing

The response writer can:
- Write status lines (200, 400, 500)
- Write headers
- Write regular bodies
- Write chunked bodies (for streaming)
- Write trailers (metadata after the body)

### Headers

Case-insensitive header storage with support for:
- Getting/setting/replacing/deleting headers
- Parsing raw header bytes
- Iterating over all headers
- Comma-separated values for duplicate headers

## Running It

```bash
# Build and run the HTTP server
go run ./cmd/httpserver

# Test it
curl http://localhost:42069/
curl http://localhost:42069/yourproblem
curl http://localhost:42069/video
```

Or use netcat for raw requests:

```bash
echo -e "GET /httpbin/html HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n" | nc localhost 42069
```

## The Other Tools

**tcplistener** - Just accepts TCP connections and prints the parsed HTTP request. Useful for debugging.

**udplistener/udpsender** - UDP tools for messing around with datagrams. Not really HTTP-related but they were part of the learning process.

## Dependencies

- Go 1.22+
- testify (for tests)

## Notes

This was built as a learning exercise. The error messages in the HTML responses are intentionally cheeky. The port number is also intentionally immature. It's fine.
