package pubsub_test

import (
	"encoding/binary"
	"testing"

	"github.com/KyberNetwork/evmlistener/pkg/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestCompressWithSizePrepended(t *testing.T) {
	s := "hello world"
	expectedResult := []byte{0x0, 0x0, 0x0, 0xb, 0xb0, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64}

	compressed, err := pubsub.CompressWithSizePrepended([]byte(s))
	assert.Nil(t, err)

	size := binary.BigEndian.Uint32(compressed[:4])
	assert.Equal(t, uint32(len(s)), size)
	assert.Equal(t, expectedResult, compressed)
}

func TestDecompressWithSizePrepended(t *testing.T) {
	data := []byte{0x0, 0x0, 0x0, 0xb, 0xb0, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64}

	decompressed, err := pubsub.DecompressWithSizePrepended(data)
	assert.Nil(t, err)
	assert.Equal(t, "hello world", string(decompressed))
}
