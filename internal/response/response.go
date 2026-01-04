package response

import (
	"fmt"
	"io"
	"strconv"
	"webserver/internal/headers"
)

type response struct {
	statusCode StatusCode
	headers    *headers.Headers
	body       []byte
}

type Writer struct {
	writer io.Writer
}

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		writer: writer,
	}
}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	h.Set("Content-Length", strconv.Itoa(contentLen))
	return h
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := []byte{}
	switch statusCode {
	case StatusOK:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	}
	_, err := w.writer.Write(statusLine)
	return err
}

func (w *Writer) WriteHeaders(h *headers.Headers) error {
	b := []byte{}
	h.ForEach(func(name, value string) {
		b = fmt.Appendf(b, "%s: %s\r\n", name, value)
	})
	b = append(b, headers.SEPARATOR...)
	_, err := w.writer.Write(b)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	n := len(p)
	w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
	_, err := w.WriteBody(p[:n])
	if err != nil {
		return 0, err
	}
	_, err = w.WriteBody([]byte("\r\n"))
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	_, err := w.WriteBody([]byte("0\r\n"))
	if err != nil {
		return 0, err
	}
	return 0, nil
}

func (w *Writer) WriteTrailers(h *headers.Headers) error {
	b := []byte{}
	h.ForEach(func(name, value string) {
		b = fmt.Appendf(b, "%s: %s\r\n", name, value)
	})
	b = append(b, headers.SEPARATOR...)
	_, err := w.writer.Write(b)
	if err != nil {
		return err
	}
	return nil
}
