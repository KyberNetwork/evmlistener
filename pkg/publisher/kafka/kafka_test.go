package kafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTopicName(t *testing.T) {
	assert.Error(t, ValidateTopicName("ethereum-listener:blocks_stream"))
	assert.Error(t, ValidateTopicName("ethereum-listener.blocks-stream"))
	assert.NoError(t, ValidateTopicName("ethereum_listener.blocks_stream"))
}
