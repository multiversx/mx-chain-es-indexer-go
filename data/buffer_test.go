package data

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBufferSlice_PutDataShouldWork(t *testing.T) {
	buffSlice := NewBufferSlice(DefaultMaxBulkSize)

	meta := generateRandomBytes(100)
	serializedData := generateRandomBytes(100)

	err := buffSlice.PutData(meta, serializedData)
	require.Nil(t, err)

	serializedData = generateRandomBytes(DefaultMaxBulkSize)
	err = buffSlice.PutData(meta, serializedData)
	require.Nil(t, err)

	returnedBuffSlice := buffSlice.Buffers()
	require.Equal(t, 2, len(returnedBuffSlice))
}

func TestBufferSlice_PutDataShouldWorkNilSerializedData(t *testing.T) {
	buffSlice := NewBufferSlice(DefaultMaxBulkSize)

	meta := []byte("my data")

	err := buffSlice.PutData(meta, nil)
	require.Nil(t, err)

	returnedBuffSlice := buffSlice.Buffers()
	require.Equal(t, 1, len(returnedBuffSlice))
}

func TestBufferSlice_PutDataShouldWorkNilSerializedDataSize1(t *testing.T) {
	buffSlice := NewBufferSlice(1)

	meta := []byte("my data")

	err := buffSlice.PutData(meta, []byte("serialized"))
	require.Nil(t, err)

	returnedBuffSlice := buffSlice.Buffers()
	require.Equal(t, 1, len(returnedBuffSlice))
	require.Equal(t, "my dataserialized\n", returnedBuffSlice[0].String())
}

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)

	return b
}
