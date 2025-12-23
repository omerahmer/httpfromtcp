package request

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"log/slog"
	"strconv"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateError   parserState = "error"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	state       parserState
	Headers     *headers.Headers
	RequestLine RequestLine
	Body        string
}

func getInt(headers *headers.Headers, name string, defaultValue int) int {
	valueStr, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func newRequest() *Request {
	return &Request{
		Headers: headers.NewHeaders(),
		state:   StateInit,
		Body:    "",
	}
}

func (r *RequestLine) ValidHTTP() bool {
	return r.HttpVersion == "HTTP/1.1"
}

var ErrorMalformedRequestLine = fmt.Errorf("malformed request line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")
var SEPARATOR = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}
	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorMalformedRequestLine
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) hasBody() bool {
	length := getInt(r.Headers, "content-length", 0)
	return length > 0
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
dance:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break dance
		}

		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break dance
			}
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break dance
			}
			read += n
			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}
		case StateBody:
			lengthStr := getInt(r.Headers, "content-length", 0)
			if lengthStr == 0 {
				panic("chunked not implemented")
			}
			remaining := min(lengthStr-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			slog.Info("parse#StateBody", "remaining", remaining, "read", read, "body", r.Body)

			if len(r.Body) == lengthStr {
				r.state = StateDone
			}

		case StateDone:
			break dance
		default:
			panic("somehow we have programmed poorly")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufIdx := 0
	for !request.done() {
		numBytesRead, err := reader.Read(buf[bufIdx:])
		if err != nil {
			return nil, err
		}
		bufIdx += numBytesRead
		readN, err := request.parse(buf[:bufIdx])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[readN:bufIdx])
		bufIdx -= readN
	}
	return request, nil
}
