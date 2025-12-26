package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type parserState string

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	state       parserState
}

var SEPARATOR = []byte("\r\n")
var ErrorMalformedRequestLine = fmt.Errorf("malformed request line")
var ErrorUnspportedHttpVersion = fmt.Errorf("unsupported HTTP version")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")

const (
	StateInit  parserState = "init"
	StateDone  parserState = "done"
	StateError parserState = "error"
)

func newRequest() *Request {
	return &Request{
		state: StateInit,
	}
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func (r *Request) parse(data []byte) (int, error) {
	readIdx := 0
outer:
	for {
		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState
		case StateInit:
			requestLine, readIdx, err := parseRequestLine(data)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if readIdx == 0 {
				break outer
			}
			r.RequestLine = *requestLine
			r.state = StateDone
			break outer
		default:
			return 0, fmt.Errorf("unknown state: %s", r.state)
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
		// If buffer is full, check if we need to grow it
		if bufLen == len(buf) {
			readIdx, err := request.parse(buf[:bufLen])
			if err != nil {
				return nil, err
			}
			if readIdx == 0 {
				// Need more data but buffer is full, grow it
				newBuf := make([]byte, len(buf)*2)
				copy(newBuf, buf)
				buf = newBuf
			} else {
				// We parsed something, shift the buffer
				copy(buf, buf[readIdx:bufLen])
				bufLen -= readIdx
			}
		}

		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			if errors.Is(err, io.EOF) {
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
