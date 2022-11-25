package block

import (
	"time"

	"github.com/KyberNetwork/evmlistener/pkg/redis"
	"go.uber.org/zap"
)

const (
	blockHeadKey = "block-head" // nolint
)

// RedisBlockKeeper ...
type RedisBlockKeeper struct {
	maxNumBlocks int
	expiration   time.Duration

	client *redis.Client
	l      *zap.SugaredLogger
}

// NewRedisBlockKeeper ...
func NewRedisBlockKeeper(client *redis.Client, maxNumBlocks int, expiration time.Duration) *RedisBlockKeeper {
	return &RedisBlockKeeper{
		maxNumBlocks: maxNumBlocks,
		expiration:   expiration,
		client:       client,
		l:            zap.S(),
	}
}
