package block

import (
	"context"
	"fmt"
	"time"

	"github.com/KyberNetwork/evmlistener/pkg/errors"
	"github.com/KyberNetwork/evmlistener/pkg/redis"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

const (
	blockHeadKey = "block-head"
)

// RedisBlockKeeper ...
type RedisBlockKeeper struct {
	expiration time.Duration

	redisClient *redis.Client
	l           *zap.SugaredLogger

	*BaseBlockKeeper
}

// NewRedisBlockKeeper ...
func NewRedisBlockKeeper(
	l *zap.SugaredLogger, client *redis.Client, maxNumBlocks int, expiration time.Duration,
) *RedisBlockKeeper {
	return &RedisBlockKeeper{
		expiration:      expiration,
		redisClient:     client,
		l:               l,
		BaseBlockKeeper: NewBaseBlockKeeper(maxNumBlocks),
	}
}

// Init ...
func (k *RedisBlockKeeper) Init() error {
	// Get blockchain head from redis.
	var hash common.Hash
	err := k.redisClient.Get(context.Background(), blockHeadKey, &hash)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil
		}

		k.l.Errorw("Fail to get blockchain head", "error", err)

		return err
	}

	// Get recent blocks from redis.
	n := k.BaseBlockKeeper.Cap()
	blocks := make([]types.Block, 0, n)
	for i := 0; i < n; i++ {
		var block types.Block
		err = k.redisClient.Get(context.Background(), hash.String(), &block)
		if err != nil {
			if errors.Is(err, errors.ErrNotFound) {
				break
			}

			k.l.Errorw("Fail to get block", "hash", hash, "error", err)

			return err
		}

		hash = block.ParentHash
		blocks = append(blocks, block)
	}

	// Store blocks on memory.
	err = k.BaseBlockKeeper.Init()
	if err != nil {
		k.l.Errorw("Fail to initialize keeper", "error", err)

		return err
	}

	for i := len(blocks) - 1; i >= 0; i-- {
		block := blocks[i]
		err = k.BaseBlockKeeper.Add(block)
		if err != nil {
			k.l.Errorw("Fail to store block on memory",
				"hash", block.Hash, "error", err)

			return err
		}
	}

	return nil
}

// Add adds new block to the keeper and store it into redis.
func (k *RedisBlockKeeper) Add(block types.Block) error {
	// Check if block was already stored in the keeper.
	exists, err := k.BaseBlockKeeper.Exists(block.Hash)
	if err != nil {
		k.l.Errorw("Fail to check block exists", "hash", block.Hash, "error", err)

		return err
	}

	if exists {
		return fmt.Errorf("block %v: %w", block.Hash, errors.ErrAlreadyExists)
	}

	// Store new block and new head into redis.
	err = k.redisClient.Set(context.Background(), block.Hash.String(), block, k.expiration)
	if err != nil {
		k.l.Errorw("Fail to store block into redis", "hash", block.Hash, "error", err)

		return err
	}

	err = k.redisClient.Set(context.Background(), blockHeadKey, block.Hash, k.expiration)
	if err != nil {
		k.l.Errorw("Fail to store block head into redis", "hash", block.Hash, "error", err)

		return err
	}

	return k.BaseBlockKeeper.Add(block)
}

// Get ...
func (k *RedisBlockKeeper) Get(hash common.Hash) (b types.Block, err error) {
	b, err = k.BaseBlockKeeper.Get(hash)
	if !errors.Is(err, errors.ErrNotFound) {
		return b, err
	}

	k.l.Debugw("Look up block from redis", "hash", hash)
	err = k.redisClient.Get(context.Background(), hash.String(), &b)

	return b, err
}
