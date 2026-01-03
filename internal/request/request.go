package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"

	"webserver/internal/headers"
)

type parserState string

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        []byte
	state       parserState
}

func getIntHeader(headers *headers.Headers, name string, defaultValue int) int {
	value, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return valueInt
}

var SEPARATOR = []byte("\r\n")
var ErrorMalformedRequestLine = fmt.Errorf("malformed request line")
var ErrorUnspportedHttpVersion = fmt.Errorf("unsupported HTTP version")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func (r *Request) parse(data []byte) (int, error) {
	readIdx := 0
outer:
	for {
		currentData := data[readIdx:]
		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState

		case StateInit:
			requestLine, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			readIdx += n
			r.RequestLine = *requestLine
			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			readIdx += n

			if n == 0 {
				break outer
			}
			if done {
				r.state = StateBody
			}

		case StateBody:
			contentLength := getIntHeader(r.Headers, "Content-Length", 0)
			if contentLength == 0 {
				r.state = StateDone
				break outer
			}
			remaining := min(contentLength-len(r.Body), len(currentData))
			r.Body = append(r.Body, currentData[:remaining]...)
			readIdx += remaining

			if len(r.Body) == contentLength {
				r.state = StateDone
				break outer
			}

		case StateDone:
			break outer

		default:
			return 0, fmt.Errorf("unknown state: %s", r.state)
		}

		// No more data to process
		if len(currentData) == 0 {
			break outer
		}
	}
	return readIdx, nil
}

func parseRequestLine(line []byte) (*RequestLine, int, error) {
	idx := bytes.Index(line, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}
	requestLine := line[:idx]
	readIdx := idx + len(SEPARATOR)

	parts := bytes.Split(requestLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || !bytes.Equal(httpParts[0], []byte("HTTP")) || !bytes.Equal(httpParts[1], []byte("1.1")) {
		return nil, 0, ErrorUnspportedHttpVersion
	}

	return &RequestLine{
		HttpVersion:   string(httpParts[1]),
		RequestTarget: string(parts[1]),
		Method:        string(parts[0]),
	}, readIdx, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 8)
	bufLen := 0
	for !request.done() {
		if bufLen == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.state == StateHeaders {
					return nil, fmt.Errorf("malformed header")
				}
				if request.state == StateBody {
					contentLength := getIntHeader(request.Headers, "Content-Length", 0)
					if len(request.Body) < contentLength {
						return nil, fmt.Errorf("body shorter than Content-Length: got %d, expected %d", len(request.Body), contentLength)
					}
				}
				request.state = StateDone
				break
			}
			return nil, err
		}
		bufLen += n
		readIdx, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readIdx:bufLen])
		bufLen -= readIdx
	}

	return request, nil
}
