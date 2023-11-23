package redis

import (
	"context"

	"github.com/KyberNetwork/evmlistener/internal/publisher"
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

// Publish publishes a message to given topic.
func (s *Stream) Publish(ctx context.Context, cfg publisher.Config, data []byte, extra map[string]string) error {
	values := []string{MessageKey, string(data)}
	if len(extra) > 0 {
		for k, v := range extra {
			values = append(values, k, v)
		}
	}

	return s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: cfg.Topic,
		MaxLen: s.maxLen,
		Approx: true,
		ID:     "*",
		Values: values,
	}).Err()
}
