package block

import (
	"context"
	"fmt"
	"time"

	"github.com/KyberNetwork/evmlistener/pkg/errors"
	"github.com/KyberNetwork/evmlistener/pkg/redis"
	"github.com/KyberNetwork/evmlistener/pkg/types"
	"go.uber.org/zap"
)

const (
	blockHeadKey = "block-head"

	minExpirationTime = time.Second
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
	var head string
	err := k.redisClient.Get(context.Background(), blockHeadKey, &head)
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
	hash := head
	for range n {
		var block types.Block
		err = k.redisClient.Get(context.Background(), hash, &block)
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

	if len(blocks) == 0 {
		k.BaseBlockKeeper.SetHead(head)

		return nil
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
	expiration := k.getExpiration(int64(block.Timestamp))
	err = k.redisClient.Set(context.Background(), block.Hash, block, expiration)
	if err != nil {
		k.l.Errorw("Fail to store block into redis", "hash", block.Hash, "error", err)

		return err
	}

	err = k.redisClient.Set(context.Background(), blockHeadKey, block.Hash, 0)
	if err != nil {
		k.l.Errorw("Fail to store block head into redis", "hash", block.Hash, "error", err)

		return err
	}

	return k.BaseBlockKeeper.Add(block)
}

// Get ...
func (k *RedisBlockKeeper) Get(hash string) (b types.Block, err error) {
	b, err = k.BaseBlockKeeper.Get(hash)
	if !errors.Is(err, errors.ErrNotFound) {
		return b, err
	}

	k.l.Debugw("Look up block from redis", "hash", hash)
	err = k.redisClient.Get(context.Background(), hash, &b)

	return b, err
}

func (k *RedisBlockKeeper) getExpiration(blockTS int64) time.Duration {
	blockTime := time.Unix(blockTS, 0)
	expiration := k.expiration - time.Since(blockTime)
	if expiration <= 0 {
		expiration = minExpirationTime
	}

	return expiration
}
