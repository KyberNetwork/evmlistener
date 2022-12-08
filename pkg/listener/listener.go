package listener

import (
	"context"
	"math/big"

	ltypes "github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const bufLen = 10000

// EVMClient is an client for evm used by listener.
type EVMClient interface {
	BlockNumber(context.Context) (uint64, error)
	SubscribeNewHead(context.Context, chan<- *types.Header) (ethereum.Subscription, error)
	FilterLogs(context.Context, ethereum.FilterQuery) ([]types.Log, error)
	HeaderByHash(context.Context, common.Hash) (*types.Header, error)
	HeaderByNumber(context.Context, *big.Int) (*types.Header, error)
}

// Listener represents a listener service for on-chain events.
type Listener struct {
	evmClient EVMClient // nolint: unused
	handler   *Handler  // nolint: unused
	l         *zap.SugaredLogger
}

// New ...
func New(evmClient EVMClient, handler *Handler) *Listener {
	return &Listener{
		evmClient: evmClient,
		handler:   handler,
		l:         zap.S(),
	}
}

func (l *Listener) handleNewHeader(ctx context.Context, header *types.Header) (ltypes.Block, error) {
	logs, err := getLogsByBlockHash(ctx, l.evmClient, header.Hash())
	if err != nil {
		l.l.Errorw("Fail to get logs by block hash", "hash", header.Hash(), "error", err)

		return ltypes.Block{}, err
	}

	return headerToBlock(header, logs), nil
}

func (l *Listener) subscribeNewBlockHead(ctx context.Context, blockCh chan<- ltypes.Block) error {
	l.l.Info("Start subscribing for new head of the chain")
	headerCh := make(chan *types.Header, 1)
	sub, err := l.evmClient.SubscribeNewHead(ctx, headerCh)
	if err != nil {
		l.l.Errorw("Fail to subscribe new head", "error", err)

		return err
	}

	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			l.l.Infow("Stop subscribing for new head")

			return nil
		case err = <-sub.Err():
			l.l.Errorw("Error while subscribing new head", "error", err)

			return err
		case header := <-headerCh:
			l.l.Debugw("Receive new head of the chain", "header", header)
			b, err := l.handleNewHeader(ctx, header)
			if err != nil {
				l.l.Errorw("Fail to handle new head", "header", header, "error", err)
			} else {
				blockCh <- b
			}
		}
	}
}

func (l *Listener) syncBlocks(ctx context.Context, blockCh chan ltypes.Block) error {
	for {
		err := l.subscribeNewBlockHead(ctx, blockCh)
		if err == nil {
			return nil
		}

		l.l.Errorw("Fail to subscribe new block head", "error", err)
		if !websocket.IsCloseError(err, websocket.CloseAbnormalClosure,
			websocket.CloseNormalClosure, websocket.CloseServiceRestart,
		) {
			return err
		}

		l.l.Infow("Re-subscribe for new block head from node")
	}
}

// Run listens for new block head and handle it.
func (l *Listener) Run(ctx context.Context) error {
	l.l.Info("Start listener service")
	defer l.l.Info("Stop listener service")

	blockCh := make(chan ltypes.Block, bufLen)
	go func() {
		err := l.syncBlocks(ctx, blockCh)
		if err != nil {
			l.l.Fatalw("Fail to sync blocks", "error", err)
		}

		close(blockCh)
	}()

	l.l.Info("Init handler")
	err := l.handler.Init(ctx)
	if err != nil {
		l.l.Errorw("Fail to init handler", "error", err)

		return err
	}

	l.l.Info("Start handling for new blocks")
	for b := range blockCh {
		l.l.Debugw("Receive new block",
			"hash", b.Hash, "parent", b.ParentHash, "numLogs", len(b.Logs))
		err = l.handler.Handle(ctx, b)
		if err != nil {
			l.l.Errorw("Fail to handle new block", "hash", b.Hash, "error", err)

			return err
		}
	}

	return nil
}
