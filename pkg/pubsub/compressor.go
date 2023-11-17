package pubsub

import (
	"encoding/binary"

	"github.com/pierrec/lz4/v4"
)

const lz4PrependSize = 4

// compress data by lz4.
func compress(input []byte, size uint32) ([]byte, error) {
	buf := make([]byte, lz4.CompressBlockBound(int(size)))
	var c lz4.Compressor

	n, err := c.CompressBlock(input, buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

// decompress data by lz4.
func decompress(input []byte, size uint32) ([]byte, error) {
	decompressed := make([]byte, size)
	_, err := lz4.UncompressBlock(input, decompressed)
	if err != nil {
		return nil, err
	}

	return decompressed, err
}

// IntToBigEndianBytes Convert an integer into 4-bytes array using BigEndian encoding.
func intToBigEndianBytes(value uint32) []byte {
	bs := make([]byte, lz4PrependSize)
	binary.BigEndian.PutUint32(bs, value)

	return bs
}

// BigEndianByteToUInt32 Read BigEndian 4-bytes array and convert to uint32.
func bigEndianByteToUInt32(input []byte) uint32 {
	return binary.BigEndian.Uint32(input)
}

// CompressWithSizePrepended prepend the size of compress output in 4 bytes array with Big Endian encoding.
func CompressWithSizePrepended(input []byte) ([]byte, error) {
	length := uint32(len(input))
	arrBytes, err := compress(input, length)
	if err != nil {
		return nil, err
	}
	lengthInBytes := intToBigEndianBytes(length)
	result := append(lengthInBytes, arrBytes...) //nolint:gocritic

	return result, nil
}

// DecompressWithSizePrepended decompress a content with first 4 bytes is a big endian number.
func DecompressWithSizePrepended(input []byte) ([]byte, error) {
	size := bigEndianByteToUInt32(input[:lz4PrependSize])

	return decompress(input[lz4PrependSize:], size)
}
