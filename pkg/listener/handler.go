package listener

import (
	"context"

	"github.com/KyberNetwork/evmlistener/pkg/block"
	"github.com/KyberNetwork/evmlistener/pkg/pubsub"
)

// Handler ...
type Handler struct {
	ethClient EVMClient        // nolint: unused
	bs        block.Keeper     // nolint: unused
	publisher pubsub.Publisher // nolint: unused
}

// Handle ...
func (h *Handler) Handle(ctx context.Context, b block.Block) error {
	return nil
}
