package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test: Valid single header
func TestHeaderParse(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\nFooFoo:       barbar       \r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)

	host, ok := headers.Get("Host")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", host)

	fooFoo, ok := headers.Get("FooFoo")
	assert.True(t, ok)
	assert.Equal(t, "barbar", fooFoo)

	_, ok = headers.Get("MissingKey")
	assert.False(t, ok)
	assert.Equal(t, 52, n)
	assert.True(t, true)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
