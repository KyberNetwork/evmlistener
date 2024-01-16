package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/IBM/sarama"
)

type Publisher struct {
	producer sarama.SyncProducer
}

func NewPublisher(config *Config) (*Publisher, error) {
	c := sarama.NewConfig()

	// Producer's MaxMessageBytes is currently compare to the size of un-compress
	// message, while it should compare to compressed size. This causes error
	// when publish some messages with large un-compress size > 1MB.
	// However, this is just client-side safety check, remote Kafka cluster
	// will check it again with correct compressed size.
	// So, set MaxMessageBytes to MaxRequestSize for now.
	// Reference: https://github.com/IBM/sarama/issues/2142
	c.Producer.MaxMessageBytes = int(sarama.MaxRequestSize)
	c.Producer.Compression = sarama.CompressionLZ4
	c.Producer.RequiredAcks = sarama.WaitForAll
	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = true

	// Use SyncProducer since we want to ensure the message is published.
	producer, err := sarama.NewSyncProducer(config.Addresses, c)
	if err != nil {
		return nil, err
	}

	return &Publisher{producer: producer}, nil
}

func (k *Publisher) Publish(ctx context.Context, topic string, data interface{}) error {
	encodedData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(encodedData),
	}

	_, _, err = k.producer.SendMessage(message)
	if err != nil {
		return err
	}

	return nil
}

// ValidationTopicName returns error if the string is invalid as Kafka topic name.
// Due to limitations in metric names, topics with a period ('.') or underscore
// ('_') could collide. To avoid issues it is best to use either, but not both.
func ValidationTopicName(topic string) error {
	legalChars := "[a-zA-Z0-9\\._\\-]"

	_, err := regexp.MatchString(topic, legalChars)
	if err != nil {
		return err
	}

	if strings.Contains(topic, "-") && strings.Contains(topic, ".") {
		return errors.New("collide characters in topic name")
	}

	return nil
}
