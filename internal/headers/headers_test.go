package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	// println("data len: ", len(data))
	n, done, err := headers.Parse(data)

	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 25, n)
	assert.True(t, done)

	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n")
	// println("data len: ", len(data))
	n, done, err = headers.Parse(data)

	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 23, n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\napi_key: *******\r\napi_key: 1234\r\n\r\n")
	n, done, err = headers.Parse(data)

	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "*******,1234", headers.Get("api_key"))
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)

	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)

	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte(" \r\n")
	n, done, err = headers.Parse(data)

	require.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.True(t, done)
}
