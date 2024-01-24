package encoder

import (
	"math/big"
	"testing"

	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/KyberNetwork/evmlistener/protobuf/pb"
	"github.com/stretchr/testify/assert"
)

func TestProtobufEncoder(t *testing.T) {
	e := NewProtobufEncoder()

	b := types.Block{Number: big.NewInt(10)}
	m := types.Message{
		RevertedBlocks: []types.Block{},
		NewBlocks:      []types.Block{b},
	}

	data, err := e.Encode(m)
	assert.NoError(t, err)

	var decoded pb.Message
	err = e.Decode(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(decoded.GetRevertedBlocks()))
	assert.Equal(t, 1, len(decoded.GetNewBlocks()))
}
