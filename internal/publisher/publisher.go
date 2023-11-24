package publisher

import (
	"context"

	"github.com/KyberNetwork/evmlistener/pkg/types"
)

type Publisher interface {
	Publish(ctx context.Context, msg types.Message) error
}
