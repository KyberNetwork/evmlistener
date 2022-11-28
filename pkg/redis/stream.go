package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

const (
	MessageKey = "message"
)

// Stream represents a redis stream.
type Stream struct {
	maxLen int64

	client *Client
}

// NewStream returns a new Stream object.
func NewStream(client *Client, maxLen int64) *Stream {
	return &Stream{
		maxLen: maxLen,
		client: client,
	}
}

// Publish publishs a message to given topic.
func (s *Stream) Publish(ctx context.Context, topic string, msg interface{}) error {
	data, err := Encode(msg)
	if err != nil {
		return err
	}

	return s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: topic,
		MaxLen: s.maxLen,
		Approx: true,
		ID:     "*",
		Values: []string{MessageKey, string(data)},
	}).Err()
}
