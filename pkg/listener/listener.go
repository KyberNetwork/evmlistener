package listener

import (
	"context"
	"math/big"
	"syscall"

	"github.com/KyberNetwork/evmlistener/pkg/errors"
	"github.com/KyberNetwork/evmlistener/pkg/evmclient"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	bufLen = 10000
)

// Listener represents a listener service for on-chain events.
type Listener struct {
	evmClient evmclient.IClient
	handler   *Handler
	l         *zap.SugaredLogger
}

// New ...
func New(l *zap.SugaredLogger, evmClient evmclient.IClient, handler *Handler) *Listener {
	return &Listener{
		evmClient: evmClient,
		handler:   handler,
		l:         l,
	}
}

func (l *Listener) handleNewHeader(ctx context.Context, header *types.Header) (types.Block, error) {
	var err error
	var logs []types.Log

	l.l.Debugw("Handle for new head", "hash", header.Hash)

	logs, err = getLogsByBlockHash(ctx, l.evmClient, header.Hash)
	if err != nil {
		l.l.Errorw("Fail to get logs by block hash", "hash", header.Hash, "error", err)

		return types.Block{}, err
	}

	l.l.Debugw("Handle new head success", "hash", header.Hash)

	return headerToBlock(header, logs), nil
}

func (l *Listener) handleOldHeaders(ctx context.Context, blockCh chan<- types.Block) error {
	blockNumber, err := l.evmClient.BlockNumber(ctx)
	if err != nil {
		l.l.Errorw("Fail to get latest block number", "error", err)

		return err
	}

	savedBlock, err := l.handler.blockKeeper.Head()
	if err != nil {
		l.l.Errorw("Fail to get last saved block", "error", err)

		return err
	}

	fromBlock := savedBlock.Number.Uint64()
	if blockNumber <= fromBlock+1 {
		return nil
	}

	l.l.Infow("Synchronize for new headers", "fromBlock", fromBlock, "toBlock", blockNumber)
	for i := fromBlock + 1; i < blockNumber; i++ {
		block, err := getBlockByNumber(ctx, l.evmClient, new(big.Int).SetUint64(i))
		if err != nil {
			l.l.Errorw("Fail to get block by number", "number", i, "error", err)

			return err
		}

		blockCh <- block
	}

	return nil
}

func (l *Listener) subscribeNewBlockHead(ctx context.Context, blockCh chan<- types.Block) error {
	l.l.Info("Start subscribing for new head of the chain")
	headerCh := make(chan *types.Header, 1)
	sub, err := l.evmClient.SubscribeNewHead(ctx, headerCh)
	if err != nil {
		l.l.Errorw("Fail to subscribe new head", "error", err)

		return err
	}

	defer sub.Unsubscribe()

	err = l.handleOldHeaders(ctx, blockCh)
	if err != nil {
		l.l.Errorw("Fail to handle old headers", "error", err)

		return err
	}

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

func (l *Listener) syncBlocks(ctx context.Context, blockCh chan types.Block) error {
	for {
		err := l.subscribeNewBlockHead(ctx, blockCh)
		if err == nil {
			return nil
		}

		l.l.Errorw("Fail to subscribe new block head", "error", err)
		if !websocket.IsCloseError(err, websocket.CloseAbnormalClosure,
			websocket.CloseNormalClosure, websocket.CloseServiceRestart) &&
			!errors.Is(err, syscall.ECONNRESET) {
			return err
		}

		l.l.Infow("Re-subscribe for new block head from node")
	}
}

// Run listens for new block head and handle it.
func (l *Listener) Run(ctx context.Context) error {
	l.l.Info("Start listener service")
	defer l.l.Info("Stop listener service")

	blockCh := make(chan types.Block, bufLen)
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
