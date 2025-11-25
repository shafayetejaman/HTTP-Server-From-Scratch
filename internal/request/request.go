package request

import (
	"bytes"
	"errors"
	"io"
	"log"
	"slices"
	"strconv"
	"strings"
	"tcpTohttp/internal/headers"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	State       State
	Body        []byte
}

type State int

const (
	StateInitialized State = iota
	StateParsingHeaders
	StateParseBody
	StateDone
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const CRLF = "\r\n"

func (r *Request) getContentLen() int {
	contentLen := r.Headers.Get("Content-Length")

	if contentLen == "" {
		return 0
	}

	n, err := strconv.Atoi(contentLen)
	if err != nil {
		log.Fatal(err)
	}
	if n < 0 {
		return 0
	}
	return n
}

func (r *Request) parse(data []byte) (int, error) {

	read := 0
	for {
		switch r.State {
		case StateDone:
			return read, nil

		case StateInitialized:
			reqLine, n, err := parseRequestLine(data[read:])
			if err != nil {
				return 0, err
			}
			if n == 0 {
				return 0, nil
			}
			r.RequestLine = reqLine
			r.State = StateParsingHeaders
			read += n

		case StateParsingHeaders:
			// slog.Info("header", "data", data[read:], "read", read)
			n, done, err := r.Headers.Parse(data[read:])

			if err != nil {
				return 0, err
			}
			read += n
			if !done {
				return read, nil
			}

			if r.getContentLen() > 0 {
				r.State = StateParseBody

			} else {
				r.State = StateDone
			}

		case StateParseBody:
			currData := data[read:]
			// too much data
			if len(r.Body)+len(currData) > r.getContentLen() {
				return 0, errors.New("too much data data")

			}
			// body missing
			if len(currData) == 0 && len(data) == 0 {
				return 0, errors.New("not enough data")
			}
			// more data need
			if len(currData) == 0 {
				return read, nil
			}

			n := min(r.getContentLen()-len(r.Body), len(currData))
			r.Body = append(r.Body, currData[:n]...)
			read += n

			if len(r.Body) == r.getContentLen() {
				r.State = StateDone
			}

		default:
			log.Fatal("state dose not match")
		}
	}
}

func (r Request) done() bool {
	return r.State == StateDone
}

func parseRequestLine(data []byte) (RequestLine, int, error) {

	var reqLine []byte
	req := RequestLine{}
	read := 0
	if i := bytes.Index(data, []byte(CRLF)); i == -1 {
		return req, 0, nil
	} else {
		reqLine = data[:i]
		read = i + len(CRLF)
	}

	reqPart := strings.Split(string(reqLine), " ")

	if len(reqPart) != 3 {
		return req, 0, errors.New("missing or to mangy arguments")
	}
	// GET /coffee HTTP/1.1
	// validate method
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	if !slices.Contains(methods, reqPart[0]) {
		return req, 0, errors.New("invalid method")
	}
	req.Method = reqPart[0]

	// validate http vertion
	httpv := strings.Split(reqPart[2], "/")
	if !(len(httpv) == 2 && httpv[0] == "HTTP" && httpv[1] == "1.1") {
		return req, 0, errors.New("invalid http vertions")
	}
	req.HttpVersion = httpv[1]

	// validate path
	path := reqPart[1]
	invalidChar := []string{"\\", " ", "+", "\n", "<", ">",
		"|", "\"", "\\'", "{", "}", "^"}

	if !strings.HasPrefix(path, "/") ||
		slices.Contains(invalidChar, path) ||
		!isAssci(path) {
		return req, 0, errors.New("invalid path")
	}
	req.RequestTarget = path

	return req, read, nil
}
func isAssci(s string) bool {
	for _, c := range s {
		if c > unicode.MaxASCII {
			return false
		}
	}
	return true
}
func RequestFromReader(reader io.Reader) (*Request, error) {
	req := Request{State: StateInitialized,
		Headers: *headers.NewHeaders()}

	buffer := make([]byte, 1024)
	read := 0

	for !req.done() {
		var n int
		var err error
		if read == len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		n, err = reader.Read(buffer[read:])
		if err != nil {
			return nil, err
		}
		read += n
		n, err = req.parse(buffer[:read])
		if err != nil {
			return nil, err
		}
		copy(buffer, buffer[n:read])
		read -= n

	}
	return &req, nil
}
