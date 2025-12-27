package main

import (
	"fmt"
	"log"
	"net"
	"webserver/internal/request"
)

/*
// getLinesChannel is a helper function that reads lines from a reader and returns them as a channel
func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string)

	go func() {
		defer f.Close()
		defer close(out)

		str := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err == io.EOF {
				break
			}

			data = data[:n]
			if i := bytes.IndexByte(data, '\n'); i >= 0 {
				str += string(data[:i])
				data = data[i+1:]
				out <- str
				str = ""
			}
			str += string(data)
		}

		if len(str) > 0 {
			out <- str
		}
	}()
	return out
}
*/

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error", "error", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error", "error", err)
		}
		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error", "error", err)
		}
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		req.Headers.ForEach(func(name, value string) {
			fmt.Printf("- %s: %s\n", name, value)
		})
	}
}
