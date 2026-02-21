package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("Valid single header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, 23, n)
		assert.False(t, done)
	})

	t.Run("Handles invalid chars", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("broken-💔: broken heart\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.Error(t, err)
		require.False(t, done)
		require.Equal(t, 0, n)
	})

	t.Run("Map does not contain any upper case keys", func(t *testing.T) {
		headers := NewHeaders()
		testStr := "BROKEN: broken heart\r\n\r\n"
		data := []byte(testStr)
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.False(t, done)
		require.Equal(t, len(data)-2, n)
		require.Equal(t, "broken heart", headers["broken"])
	})

	t.Run("Invalid spacing header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("       Host : localhost:42069       \r\n\r\n")
		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Handles appending values to already existing keys", func(t *testing.T) {
		headers := NewHeaders()
		headers["host"] = "localhost:69"
		data := []byte("Host: localhost:42069\r\n\r\n")
		_, _, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "localhost:69,localhost:42069", headers["host"])
	})
}
