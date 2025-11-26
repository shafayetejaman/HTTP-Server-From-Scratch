package response

import (
	"bytes"
	"errors"
	"io"
	"net"
	"strconv"
	"tcpTohttp/internal/headers"
)

type StatusCode int

const CRLF = "\r\n"
const (
	StatusOK            StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var statusPhrase string
	httpVersion := "HTTP/1.1"

	switch statusCode {
	case StatusOK:
		statusPhrase = httpVersion + " 200 OK"
	case BadRequest:
		statusPhrase = httpVersion + " 400 Bad Request"
	case InternalServerError:
		statusPhrase = httpVersion + " 500 Internal Server Error"
	default:
		statusPhrase = httpVersion + " " + strconv.Itoa(int(statusCode)) + " "
	}

	_, err := w.Write([]byte(statusPhrase + CRLF))

	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")
	return *headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	var data string
	for key, val := range headers.Headers {
		data += key + ": " + val + CRLF
	}
	data += CRLF
	_, err := w.Write([]byte(data))

	return err
}

type Writer struct {
	Conn   net.Conn
	Status StatusWrite
}
type StatusWrite int

const (
	StatusWriteStatusLine StatusWrite = iota
	StatusWriteHeaders
	StatusWriteBody
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.Status != StatusWriteStatusLine {
		return errors.New("alwady writen status line")
	}

	buff := bytes.NewBuffer([]byte{})
	err := WriteStatusLine(buff, statusCode)
	if err != nil {
		return err
	}
	w.Conn.Write(buff.Bytes())
	w.Status = StatusWriteHeaders
	return nil
}

func (w *Writer) WriteHeaders(headers *headers.Headers,
	delHeaders []string, trailers *headers.Headers) error {

	if w.Status != StatusWriteHeaders {
		return errors.New("alwady writen Headers")
	}
	buff := bytes.NewBuffer([]byte{})
	defHeaders := GetDefaultHeaders(0)

	if headers != nil {
		for key, val := range headers.Headers {
			defHeaders.Replace(key, val)
		}
	}

	for _, key := range delHeaders {
		defHeaders.Delete(key)
	}

	if trailers != nil {
		for key := range trailers.Headers {
			defHeaders.Set("Trailer", key)
		}
	}

	err := WriteHeaders(buff, defHeaders)
	if err != nil {
		return err
	}
	w.Conn.Write(buff.Bytes())
	w.Status = StatusWriteBody
	return nil

}
func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.Status != StatusWriteBody {
		return 0, errors.New("alwady writen status line")
	}
	return w.Conn.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	lengthLine := strconv.FormatInt(int64(len(p)), 16) + CRLF

	var buf bytes.Buffer
	buf.WriteString(lengthLine)
	buf.Write(p)
	buf.WriteString(CRLF)

	data := buf.Bytes()
	read := 0
	for read < len(data) {
		n, err := w.Conn.Write(data[read:])
		if err != nil {
			return 0, err
		}
		read += n
	}
	return len(p), nil
}
func (w *Writer) WriteChunkedBodyDone() (int, error) {

	data := []byte("0" + CRLF)
	read := 0
	for read < len(data) {
		n, err := w.Conn.Write(data[read:])
		if err != nil {
			return 0, err
		}
		read += n
	}
	return read, nil

}
func (w *Writer) WriteTrailers(h *headers.Headers) error {
	for key, val := range h.Headers {
		_, err := w.Conn.Write([]byte(key + ":" + val + CRLF))
		if err != nil {
			return err
		}
	}
	_, err := w.Conn.Write([]byte(CRLF))
	return err
}
