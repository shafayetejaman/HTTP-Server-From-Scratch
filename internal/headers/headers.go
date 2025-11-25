package headers

import (
	"bytes"
	"errors"
	"strings"
	"unicode"
)

type Headers struct {
	Headers map[string]string
}

func (h *Headers) Get(key string) string {
	key = strings.ToLower(key)

	return h.Headers[key]
}

func (h *Headers) Set(key, val string) {
	key = strings.ToLower(key)

	// if oldVal, ok := h.Headers[key]; ok {
	// 	val = oldVal + "," + val
	// }
	h.Headers[key] = val
}

const CRLF = "\r\n"

func isTChar(s string) bool {
	if len(s) < 1 {
		return false
	}

	valid := "!#$%&'*+-.^_`|~"

	for _, c := range s {
		if unicode.IsLetter(c) ||
			unicode.IsDigit(c) ||
			strings.ContainsRune(valid, c) {
			continue
		}
		return false
	}
	return true
}

func ParseHeader(headerLine string) (string, string, error) {
	firstColonIdx := strings.Index(headerLine, ":")
	// println(headerLine)k
	if firstColonIdx == -1 {
		return "", "", errors.New("missing : in header")
	}
	key := headerLine[:firstColonIdx]
	val := strings.TrimPrefix(headerLine[firstColonIdx+1:], " ")

	if !isTChar(key) {
		return "", "", errors.New("invalid key for header")
	}

	return key, val, nil
}

func (h *Headers) Parse(data []byte) (n int, done bool, err error) {

	read := 0

	for {
		var headerLine string

		if idx := bytes.Index(data[read:], []byte(CRLF)); idx == -1 {
			return read, false, nil
		} else {
			headerLine = string(bytes.Trim(data[read:read+idx], " "))
			read += idx + len(CRLF)
			if len(headerLine) == 0 {
				break
			}

		}

		key, val, err := ParseHeader(headerLine)

		if err != nil {
			return 0, false, err
		}

		h.Set(key, val)
	}

	return read, true, nil

}

func NewHeaders() *Headers {
	return &Headers{Headers: make(map[string]string)}
}
