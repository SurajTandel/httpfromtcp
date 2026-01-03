package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers struct {
	headers map[string]string
}

var SEPARATOR = []byte("\r\n")
var ErrorMalformedHeader = fmt.Errorf("malformed header")
var ErrorMalformedHeaderKey = fmt.Errorf("malformed header key")
var ErrorMalformedHeaderValue = fmt.Errorf("malformed header value")

func parseHeader(data []byte) (string, string, error) {
	parts := bytes.SplitN(data, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", ErrorMalformedHeader
	}
	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	// Check for trailing space before colon (invalid per HTTP spec)
	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", ErrorMalformedHeaderKey
	}

	// Trim leading spaces and validate
	name = bytes.TrimSpace(name)
	if len(name) == 0 {
		return "", "", ErrorMalformedHeaderKey
	}

	return string(name), string(value), nil
}

var validTokenChars = map[byte]bool{
	'!': true, '#': true, '$': true, '%': true, '&': true,
	'\'': true, '*': true, '+': true, '-': true, '.': true,
	'^': true, '_': true, '`': true, '|': true, '~': true,
}

func isToken(str []byte) bool {
	if len(str) == 0 {
		return false
	}
	for _, c := range str {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || validTokenChars[c]) {
			return false
		}
	}
	return true
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Get(name string) (string, bool) {
	value, exists := h.headers[strings.ToLower(name)]
	return value, exists
}

func (h *Headers) Set(name, value string) {
	if _, exists := h.headers[strings.ToLower(name)]; exists {
		h.headers[strings.ToLower(name)] += "," + value
		return
	}
	h.headers[strings.ToLower(name)] = value
}

func (h *Headers) Replace(name, value string) {
	h.headers[strings.ToLower(name)] = value
}

func (h *Headers) ForEach(callback func(name, value string)) {
	for name, value := range h.headers {
		callback(name, value)
	}
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], SEPARATOR)
		if idx == -1 {
			return read, false, nil
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
		if !isToken([]byte(name)) {
			return 0, false, ErrorMalformedHeaderKey
		}
		h.Set(name, value)
		read += idx + len(SEPARATOR)
	}
	return read, done, nil
}
