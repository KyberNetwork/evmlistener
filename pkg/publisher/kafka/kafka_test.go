package kafka

import (
	"context"
	"testing"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/stretchr/testify/assert"
)

func TestValidateTopicName(t *testing.T) {
	assert.Error(t, ValidateTopicName("ethereum-listener:blocks_stream"))
	assert.Error(t, ValidateTopicName("ethereum-listener.blocks-stream"))
	assert.NoError(t, ValidateTopicName("ethereum_listener.blocks_stream"))
}

func TestPublishMessage(t *testing.T) {
	seedBroker := sarama.NewMockBroker(t, 1)
	leader := sarama.NewMockBroker(t, 2)
	topic := "ethereum_listener.blocks_stream"

	metadataResponse := new(sarama.MetadataResponse)
	metadataResponse.AddBroker(leader.Addr(), leader.BrokerID())
	metadataResponse.AddTopicPartition(topic, 0, leader.BrokerID(), nil, nil, nil, sarama.ErrNoError)
	seedBroker.Returns(metadataResponse)

	prodSuccess := new(sarama.ProduceResponse)
	prodSuccess.AddTopicPartition(topic, 0, sarama.ErrNoError)
	leader.Returns(prodSuccess)

	config := mocks.NewTestConfig()
	config.Producer.Return.Successes = true
	testProducer, err := sarama.NewSyncProducer(
		[]string{seedBroker.Addr()},
		config,
	)
	// Since NewPublisher doesn't expose to customize config,
	// I use direct struct Publisher initialize instead.
	publisher := &Publisher{
		producer: testProducer,
	}

	assert.NoError(t, err)
	assert.NoError(t, publisher.Publish(context.Background(), topic, []byte("hello")))

	seedBroker.Close()
	leader.Close()
}
