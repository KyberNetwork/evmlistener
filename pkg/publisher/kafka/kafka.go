package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"google.golang.org/protobuf/proto"
)

type Publisher struct {
	producer           sarama.SyncProducer
	totalMessage       uint32
	jsonEncodeTime     time.Duration
	protobufEncodeTime time.Duration
}

func NewPublisher(config *Config) (*Publisher, error) {
	c := sarama.NewConfig()

	// Sarama Producer's MaxMessageBytes is currently compare to the size of
	// un-compress message, while it should compare to compressed size.
	// This causes error when publish some messages with large un-compress size,
	// while the compressed size is smaller than Broker's config.
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
	k.totalMessage += 1
	evmMessage := data.(types.Message)

	// JSON encode and publish messages
	start := time.Now()
	jsonEncodedData, err := json.Marshal(evmMessage)
	k.jsonEncodeTime += time.Since(start)
	if err != nil {
		return err
	}
	jsonEncodedMessage := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(jsonEncodedData),
	}
	_, _, err = k.producer.SendMessage(jsonEncodedMessage)
	if err != nil {
		return err
	}

	// Protobuf encode and publish messages
	start = time.Now()
	protobufMessage := evmMessage.ToProtobuf()
	protobufEncodedData, err := proto.Marshal(protobufMessage)
	k.protobufEncodeTime += time.Since(start)
	if err != nil {
		return err
	}
	protobufEncodedMessage := &sarama.ProducerMessage{
		Topic: topic + ".protobuf",
		Value: sarama.ByteEncoder(protobufEncodedData),
	}
	_, _, err = k.producer.SendMessage(protobufEncodedMessage)
	if err != nil {
		return err
	}

	fmt.Println("JSON Encode", k.jsonEncodeTime, k.jsonEncodeTime/time.Duration(k.totalMessage))
	fmt.Println("Protobuf Encode", k.protobufEncodeTime, k.protobufEncodeTime/time.Duration(k.totalMessage))

	return nil
}

// ValidateTopicName returns error if the string is invalid as Kafka topic name.
// NOTE: Due to limitations in metric names, topics with a period ('.') or underscore
// ('_') could collide. To avoid issues it is best to use either, but not both.
func ValidateTopicName(topic string) error {
	expression := "^[a-zA-Z0-9\\._\\-]+$"
	matched, err := regexp.MatchString(expression, topic)
	if err != nil {
		return err
	}

	if !matched {
		return errors.New("invalid characters in topic name")
	}

	if strings.Contains(topic, "-") && strings.Contains(topic, ".") {
		return errors.New("collide characters in topic name")
	}

	return nil
}
