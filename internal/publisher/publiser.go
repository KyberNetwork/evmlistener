package publisher

import (
	"context"

	"github.com/KyberNetwork/evmlistener/pkg/types"
)

type Client interface {
	Publish(ctx context.Context, cfg Config, data []byte, extra map[string]string) error
}

type Publisher interface {
	Publish(ctx context.Context, msg types.Message) error
}
