package listener

import (
	"context"
	"math/big"

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
	option      *FilterOption
}

// NewHandler ...
func NewHandler(
	l *zap.SugaredLogger, topic string, evmClient evmclient.IClient,
	blockKeeper block.Keeper, publisher pubsub.Publisher, options ...Option,
) *Handler {
	var opts FilterOption
	for _, v := range options {
		v(&opts)
	}

	return &Handler{
		topic:       topic,
		evmClient:   evmClient,
		blockKeeper: blockKeeper,
		publisher:   publisher,
		l:           l,
		option:      &opts,
	}
}

func (h *Handler) getBlockNumber(ctx context.Context) (uint64, error) {
	hash := h.blockKeeper.GetHead()
	if hash != "" {
		h.l.Infow("Get header from block hash", "hash", hash)
		header, err := getHeaderByHash(ctx, h.evmClient, hash)
		if err != nil {
			h.l.Errorw("Fail to get header by hash", "hash", hash, "error", err)

			return 0, err
		}

		return header.Number.Uint64(), nil
	}

	h.l.Infow("Get latest block number from node")

	return h.evmClient.BlockNumber(ctx)
}

// Init ...
func (h *Handler) Init(ctx context.Context) error {
	h.l.Info("Init block keeper")
	err := h.blockKeeper.Init()
	if err != nil {
		h.l.Errorw("Fail to initialize block keeper", "error", err)

		return err
	}

	if h.blockKeeper.Len() > 0 {
		return nil
	}

	h.l.Info("Get saved block number or latest block number")
	toBlock, err := h.getBlockNumber(ctx)
	if err != nil {
		h.l.Errorw("Fail to get block number", "error", err)

		return err
	}

	fromBlock := toBlock - uint64(h.blockKeeper.Cap()) + 1

	h.l.Infow("Get blocks from node", "from", fromBlock, "to", toBlock)
	blocks, err := GetBlocks(ctx, h.evmClient, fromBlock, toBlock, h.option.withLogs,
		h.option.filterContracts, h.option.filterTopics)
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
		h.l.Errorw("Fail to get block from redis", "hash", hash, "error", err)

		return types.Block{}, err
	}

	b, err = getBlockByHash(ctx, h.evmClient, hash, h.option.withLogs, h.option.filterContracts,
		h.option.filterTopics)
	if err != nil {
		h.l.Errorw("Fail to get block from ndoe", "hash", hash, "error", err)

		return types.Block{}, err
	}

	return b, nil
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
	for i := range n / 2 {
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

	blockHead, err := h.blockKeeper.Head()
	if err == nil {
		blockDiff := new(big.Int).Sub(blockHead.Number, b.Number).Int64()
		if blockDiff > int64(h.blockKeeper.Cap()) {
			log.Warnw("Ignore block that too old",
				"blockNumber", b.Number,
				"blockHeadNumber", blockHead.Number,
				"blockDiff", blockDiff,
			)

			return nil
		}
	}

	return h.handleNewBlock(ctx, b)
}
