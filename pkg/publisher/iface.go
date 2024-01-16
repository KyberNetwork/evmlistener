package publisher

import "context"

// Publisher ...
type Publisher interface {
	Publish(ctx context.Context, topic string, data interface{}) error
}
