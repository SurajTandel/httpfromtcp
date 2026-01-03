package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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
