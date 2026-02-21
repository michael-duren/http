package request

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	t.Run("Good GET Request line", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	})

	t.Run("Good GET Request line with path", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 1,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	})

	t.Run("Invalid number of parts", func(t *testing.T) {
		s := "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"
		_, err := RequestFromReader(
			&chunkReader{
				data:            s,
				numBytesPerRead: len(s),
			},
		)
		require.Error(t, err)
	})

	t.Run("Only supports HTTP 1.1", func(t *testing.T) {
		_, err := RequestFromReader(strings.NewReader("GET /coffee HTTP/2.0\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
		require.Error(t, err)
	})

	t.Run("Invalid number of items in request line", func(t *testing.T) {
		_, err := RequestFromReader(strings.NewReader("HTTP GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
		require.Error(t, err)
	})

	t.Run("Invalid HTTP method", func(t *testing.T) {
		_, err := RequestFromReader(strings.NewReader("GETT /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
		require.Error(t, err)
	})

	t.Run("Doesn't allow lowercase method name", func(t *testing.T) {
		_, err := RequestFromReader(strings.NewReader("get /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
		require.Error(t, err)
	})

	t.Run("EOF with last bytes", func(t *testing.T) {
		reader := &eofReader{
			data: "GET / HTTP/1.1\r\n\r\n",
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	})
}

func TestHeaderParse(t *testing.T) {
	t.Run("Standard Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "localhost:42069", r.Headers["host"])
		assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
		assert.Equal(t, "*/*", r.Headers["accept"])
	})
	t.Run("Malformed Header", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
			numBytesPerRead: 3,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})
	t.Run("Empty Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.Equal(t, len(r.Headers), 0)
	})
	t.Run("Duplicate headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nLocation: 123\r\nLocation: 456\r\n\r\n",
			numBytesPerRead: 30,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.Equal(t, "123,456", r.Headers["location"])
	})

	t.Run("Case insensitive", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nLOCATION: 123\r\nlocation: 456\r\n\r\n",
			numBytesPerRead: 12,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.Equal(t, "123,456", r.Headers["location"])
	})

	t.Run("Multiple headers in one line", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nLOCATION: 123\r\nlocation: 456\r\nlocation: 456\r\nlocation: 456\r\nlocation: 456\r\n\r\n",
			numBytesPerRead: 80,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.Equal(t, "123,456,456,456,456", r.Headers["location"])
	})
}

type eofReader struct {
	data string
	pos  int
}

func (r *eofReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	if r.pos >= len(r.data) {
		return n, io.EOF
	}
	return n, nil
}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}

	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))

	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}
