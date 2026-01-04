package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"webserver/internal/headers"
	"webserver/internal/request"
	"webserver/internal/response"
	"webserver/internal/server"
)

const port = 42069

func respond400() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func respond500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func respond200() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		body := respond200()
		statusCode := response.StatusOK
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			body = respond400()
			statusCode = response.StatusBadRequest
		case "/myproblem":
			body = respond500()
			statusCode = response.StatusInternalServerError
		case "/video":

			f, err := os.ReadFile("assets/vim.mp4")
			if err != nil {
				body = respond500()
				statusCode = response.StatusInternalServerError
			} else {
				h.Replace("Content-Type", "video/mp4")
				h.Replace("Content-Length", strconv.Itoa(len(f)))

				w.WriteStatusLine(response.StatusOK)
				w.WriteHeaders(h)
				w.WriteBody(f)
				return
			}
		default:
			if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
				target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
				resp, err := http.Get("https://httpbin.org/" + target)
				if err != nil {
					body = respond500()
					statusCode = response.StatusInternalServerError
				} else {
					w.WriteStatusLine(response.StatusOK)
					h.Delete("Content-Length")
					h.Set("Transfer-Encoding", "chunked")
					h.Replace("Content-Type", "text/plain")
					h.Set("Trailer", "X-Content-SHA256")
					h.Set("Trailer", "X-Content-Length")
					w.WriteHeaders(h)

					fullBody := []byte{}
					for {
						data := make([]byte, 32)
						n, err := resp.Body.Read(data)
						if err != nil {
							break
						}
						fullBody = append(fullBody, data[:n]...)
						w.WriteChunkedBody(data[:n])
					}
					w.WriteChunkedBodyDone()
					trailers := headers.NewHeaders()
					trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", sha256.Sum256(fullBody)))
					trailers.Set("X-Content-Length", strconv.Itoa(len(fullBody)))
					w.WriteTrailers(trailers)

					return
				}
			}
		}
		h.Replace("Content-Length", strconv.Itoa(len(body)))
		h.Replace("Content-Type", "text/html")
		w.WriteStatusLine(statusCode)
		w.WriteHeaders(h)
		w.WriteBody(body)
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close(s)
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
