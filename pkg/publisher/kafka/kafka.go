package kafka

import (
	"context"
	"encoding/json"
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
	// TODO: Move this into shared utils function
	encodedData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// TODO: Remove this
	topic = normalizeKafkaTopic(topic)

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

func (k *Publisher) Cleanup() error {
	return k.producer.Close()
}

// normalizeKafkaTopic replaces characters that doesn't support as a kafka topic.
// Quote: Due to limitations in metric names, topics with a period ('.') or underscore
// ('_') could collide. To avoid issues it is best to use either, but not both.
// https://github.com/apache/kafka/blob/c8d61a5cbe9acfc78b7af64b1607a8aa772dad39/tools/src/main/java/org/apache/kafka/tools/TopicCommand.java#L440-L447
// TODO: Should move as a validate function. Let the caller decide the topic name.
func normalizeKafkaTopic(topic string) string {
	topic = strings.ReplaceAll(topic, ":", ".")
	topic = strings.ReplaceAll(topic, "_", "-")
	return topic
}
