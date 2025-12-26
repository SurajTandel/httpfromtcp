package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

var SEPARATOR = []byte("\r\n")
var ErrorMalformedHeader = fmt.Errorf("malformed header")
var ErrorMalformedHeaderKey = fmt.Errorf("malformed header key")
var ErrorMalformedHeaderValue = fmt.Errorf("malformed header value")

func NewHeaders() Headers {
	return map[string]string{}
}

func parseHeader(data []byte) (string, string, error) {
	parts := bytes.SplitN(data, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", ErrorMalformedHeader
	}
	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", ErrorMalformedHeaderKey
	}

	return string(name), string(value), nil
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], SEPARATOR)
		if idx == -1 {
			return 0, false, ErrorMalformedHeader
		}
		if idx == 0 {
			done = true
			read += len(SEPARATOR)
			break
		}

		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}
		h[name] = value
		read += idx + len(SEPARATOR)
	}
	return read, done, nil
}
