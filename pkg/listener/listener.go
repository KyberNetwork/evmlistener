package listener

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

// EVMClient is an client for evm used by listener.
type EVMClient interface {
	BlockNumber(context.Context) (uint64, error)
	SubscribeNewHead(context.Context, chan<- *types.Header) (ethereum.Subscription, error)
	FilterLogs(context.Context, ethereum.FilterQuery) ([]types.Log, error)
	HeaderByHash(context.Context, common.Hash) (*types.Header, error)
}

// Listener represents a listener service for on-chain events.
type Listener struct {
	evmClient EVMClient // nolint: unused
	handler   *Handler  // nolint: unused
	l         *zap.SugaredLogger
}

// Run listens for new block head and handle it.
func (l *Listener) Run(ctx context.Context) error {
	l.l.Info("Start listener service")
	defer l.l.Info("Stop listener service")

	return nil
}
