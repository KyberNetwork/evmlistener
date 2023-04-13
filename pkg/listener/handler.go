package listener

import (
	"context"

	"github.com/KyberNetwork/evmlistener/pkg/block"
	"github.com/KyberNetwork/evmlistener/pkg/errors"
	"github.com/KyberNetwork/evmlistener/pkg/evmclient"
	"github.com/KyberNetwork/evmlistener/pkg/pubsub"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"go.uber.org/zap"
)

// Handler ...
type Handler struct {
	topic string

	evmClient   evmclient.IClient
	blockKeeper block.Keeper
	publisher   pubsub.Publisher
	l           *zap.SugaredLogger
}

// NewHandler ...
func NewHandler(
	l *zap.SugaredLogger, topic string, evmClient evmclient.IClient,
	blockKeeper block.Keeper, publisher pubsub.Publisher,
) *Handler {
	return &Handler{
		topic:       topic,
		evmClient:   evmClient,
		blockKeeper: blockKeeper,
		publisher:   publisher,
		l:           l,
	}
}

// Init ...
func (h *Handler) Init(ctx context.Context) error {
	h.l.Info("Get latest block number")
	toBlock := uint64(17034440) //nolint

	fromBlock := toBlock - uint64(h.blockKeeper.Cap()) + 1

	h.l.Infow("Get blocks from node", "from", fromBlock, "to", toBlock)
	blocks, err := getBlocks(ctx, h.evmClient, fromBlock, toBlock)
	if err != nil {
		h.l.Errorw("Fail to get blocks", "from", fromBlock, "to", toBlock, "error", err)

		return err
	}

	h.l.Infow("Add new blocks to block storage", "len", len(blocks))
	for _, b := range blocks {
		err = h.blockKeeper.Add(b)
		if err != nil {
			h.l.Errorw("Fail to store block", "block", b, "error", err)

			return err
		}
	}

	return nil
}

// getBlock returns block from block keeper or fetch from evm client.
func (h *Handler) getBlock(ctx context.Context, hash string) (types.Block, error) {
	b, err := h.blockKeeper.Get(hash)
	if err == nil {
		return b, nil
	}

	if !errors.Is(err, errors.ErrNotFound) {
		return types.Block{}, err
	}

	return getBlockByHash(ctx, h.evmClient, hash)
}

func (h *Handler) findReorgBlocks(
	ctx context.Context, storedBlock, newBlock types.Block,
) ([]types.Block, []types.Block, error) {
	h.l.Debugw("Find re-organization blocks",
		"oldHash", storedBlock.Hash,
		"newHash", newBlock.Hash,
		"oldParentHash", storedBlock.ParentHash,
		"newParentHash", newBlock.ParentHash,
		"oldBlockNumber", storedBlock.Number,
		"newBlockNumber", newBlock.Number)

	var err error
	var reorgBlocks, newBlocks []types.Block
	storedNumber := storedBlock.Number.Uint64()
	newNumber := newBlock.Number.Uint64()

	for {
		if storedNumber >= newNumber {
			reorgBlocks = append(reorgBlocks, storedBlock)
			storedBlock, err = h.blockKeeper.Get(storedBlock.ParentHash)
			if err != nil {
				h.l.Errorw("Fail to get stored block",
					"number", storedNumber, "error", err)

				return nil, nil, err
			}

			storedNumber--
		}

		if newNumber > storedNumber {
			newBlocks = append(newBlocks, newBlock)
			newBlock, err = h.getBlock(ctx, newBlock.ParentHash)
			if err != nil {
				h.l.Errorw("Fail to get new block",
					"number", newNumber, "error", err)

				return nil, nil, err
			}

			newNumber--
		}

		if storedBlock.Hash == newBlock.Hash {
			break
		}
	}

	n := len(newBlocks)
	for i := 0; i < n/2; i++ {
		newBlocks[i], newBlocks[n-i-1] = newBlocks[n-i-1], newBlocks[i]
	}

	return reorgBlocks, newBlocks, nil
}

func (h *Handler) handleReorgBlock(
	ctx context.Context, b types.Block,
) (revertedBlocks []types.Block, newBlocks []types.Block, err error) {
	head, err := h.blockKeeper.Head()
	if err != nil {
		h.l.Errorw("Fail to get stored block head", "error", err)

		return nil, nil, err
	}

	return h.findReorgBlocks(ctx, head, b)
}

func (h *Handler) handleNewBlock(ctx context.Context, b types.Block) error {
	log := h.l.With(
		"blockNumber", b.Number, "blockHash", b.Hash,
		"parentHash", b.ParentHash, "numLogs", len(b.Logs),
	)

	log.Infow("Handling new block")

	isReorg, err := h.blockKeeper.IsReorg(b)
	if err != nil {
		log.Errorw("Fail to check for re-organization", "error", err)

		return err
	}

	var revertedBlocks, newBlocks []types.Block
	if isReorg {
		log.Infow("Handle re-organization block")
		revertedBlocks, newBlocks, err = h.handleReorgBlock(ctx, b)
		if err != nil {
			log.Errorw("Fail to handle re-organization block", "error", err)

			return err
		}
	} else {
		newBlocks = []types.Block{b}
	}

	log.Infow("Publish message to queue",
		"topic", h.topic,
		"numRevertedBlocks", len(revertedBlocks),
		"numNewBlocks", len(newBlocks))
	msg := types.Message{
		RevertedBlocks: revertedBlocks,
		NewBlocks:      newBlocks,
	}
	err = h.publisher.Publish(ctx, h.topic, msg)
	if err != nil {
		log.Errorw("Fail to publish message", "error", err)

		return err
	}

	// Add new blocks into block keeper.
	for _, b := range newBlocks {
		err = h.blockKeeper.Add(b)
		if err != nil && !errors.Is(err, errors.ErrAlreadyExists) {
			h.l.Errorw("Fail to add block", "hash", b.Hash, "error", err)

			return err
		}
	}

	return nil
}

// Handle ...
func (h *Handler) Handle(ctx context.Context, b types.Block) error {
	log := h.l.With(
		"blockNumber", b.Number, "blockHash", b.Hash,
		"parentHash", b.ParentHash, "numLogs", len(b.Logs),
	)

	exists, err := h.blockKeeper.Exists(b.Hash)
	if err != nil {
		log.Errorw("Fail to check exists for block", "hash", b.Hash, "error", err)

		return err
	}

	if exists {
		log.Debugw("Ignore already handled block", "hash", b.Hash)

		return nil
	}

	return h.handleNewBlock(ctx, b)
}
