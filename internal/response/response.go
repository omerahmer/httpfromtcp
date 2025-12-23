package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type Response struct {
}

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

var stateName = map[StatusCode]string{
	StatusOK:                  "HTTP/1.1 200 OK\r\n",
	StatusBadRequest:          "HTTP/1.1 400 Bad Request\r\n",
	StatusInternalServerError: "HTTP/1.1 500 Internal Server Error\r\n",
}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

type Writer struct {
	writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer: writer}
}

// func (w *Writer) WriteChunkedBody(p []byte) (int, error) {

// }

// func (w *Writer) WriteChunkedBodyDone() (int, error) {

// }

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	line, ok := stateName[statusCode]
	if !ok {
		return fmt.Errorf("unrecognized error code")
	}
	_, err := w.writer.Write([]byte(line))
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	b := []byte{}
	h.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
	})
	b = fmt.Append(b, "\r\n")
	_, err := w.writer.Write(b)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)

	return n, err
}
