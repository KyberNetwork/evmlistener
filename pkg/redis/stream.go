package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
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
func (s *Stream) Publish(ctx context.Context, topic string, data []byte) error {
	return s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: topic,
		MaxLen: s.maxLen,
		Approx: true,
		ID:     "*",
		Values: []string{MessageKey, string(data)},
	}).Err()
}
