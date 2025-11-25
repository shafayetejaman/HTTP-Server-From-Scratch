package request

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	r, err := RequestFromReader(strings.NewReader(
		"GET / HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n",
	))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	r, err = RequestFromReader(strings.NewReader(
		"GET /coffee HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n",
	))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: POST with body
	r, err = RequestFromReader(strings.NewReader(
		"POST /users HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Content-Length: 12\r\n" +
			"Accept: */*\r\n" +
			"\r\n" +
			"{name:akash}",
	))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/users", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Invalid number of parts
	_, err = RequestFromReader(strings.NewReader(
		"/coffee HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n",
	))
	require.Error(t, err)

	// Test: Invalid order
	_, err = RequestFromReader(strings.NewReader(
		"/coffee GET HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n",
	))
	require.Error(t, err)

	// Test: Invalid HTTP version
	_, err = RequestFromReader(strings.NewReader(
		"GET /coffee HTTP/23.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n",
	))
	require.Error(t, err)

	// Chunked tests
	reader := &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)

	// with path
	reader = &chunkReader{
		data: "GET /coffee HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)

	// big chunk read
	reader = &chunkReader{
		data: "GET /coffee HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"\r\n",
		numBytesPerRead: 1012,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)

}

func TestHeaderAndReqLine(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	assert.Equal(t, "localhost:42069", r.Headers.Get("host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("user-agent"))
	assert.Equal(t, "*/*", r.Headers.Get("accept"))

	reader = &chunkReader{
		data: "GET / HTTP\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept : */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	reader = &chunkReader{
		data: "GET /files/coffee HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", r.Headers.Get("Host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("user-agent"))
	assert.Equal(t, "*/*", r.Headers.Get("accept"))
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/files/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestPasingTheBody(t *testing.T) {
	// Test: Standard Body
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	// Test: Body shorter than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 5\r\n" +
			"\r\n" +
			"more content",
		numBytesPerRead: 30,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Content-Length: 5\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}
