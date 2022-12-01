package app

import (
	"context"

	"github.com/KyberNetwork/evmlistener/pkg/block"
	"github.com/KyberNetwork/evmlistener/pkg/listener"
	"github.com/KyberNetwork/evmlistener/pkg/redis"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// NewApp creates a new cli App instance with common flags pre-loaded.
func NewApp() *cli.App {
	app := cli.NewApp()
	app.Flags = NewFlags()

	return app
}

func redisConfigFromCli(c *cli.Context) redis.Config {
	cfg := redis.Config{
		MasterName:   c.String(redisMasterNameFlag.Name),
		Addrs:        c.StringSlice(redisAddrsFlag.Name),
		DB:           c.Int(redisDBFlag.Name),
		KeyPrefix:    c.String(redisKeyPrefixFlag.Name),
		ReadTimeout:  c.Duration(redisReadTimeoutFlag.Name),
		WriteTimeout: c.Duration(redisWriteTimeoutFlag.Name),
	}

	if cfg.MasterName != "" {
		cfg.SentinelUsername = c.String(redisUsernameFlag.Name)
		cfg.SentinelPassword = c.String(redisPasswordFlag.Name)
	} else {
		cfg.Username = c.String(redisUsernameFlag.Name)
		cfg.Password = c.String(redisPasswordFlag.Name)
	}

	return cfg
}

// NewListener setups and returns listener service.
func NewListener(c *cli.Context) (*listener.Listener, error) {
	l := zap.S()

	rpc := c.String(nodeRPCFlag.Name)
	ethClient, err := ethclient.DialContext(context.Background(), rpc)
	if err != nil {
		l.Errorw("Fail to connect to node", "rpc", rpc, "error", err)

		return nil, err
	}

	redisConfig := redisConfigFromCli(c)
	redisClient, err := redis.New(redisConfig)
	if err != nil {
		l.Errorw("Fail to connect to redis", "cfg", redisConfig, "error", err)

		return nil, err
	}

	maxNumBlocks := c.Int(maxNumBlocksFlag.Name)
	blockExpiration := c.Duration(blockExpirationFlag.Name)
	blockKeeper := block.NewRedisBlockKeeper(redisClient, maxNumBlocks, blockExpiration)

	maxLen := c.Int64(publisherMaxLenFlag.Name)
	redisStream := redis.NewStream(redisClient, maxLen)

	topic := c.String(publisherTopicFlag.Name)
	handler := listener.NewHandler(topic, ethClient, blockKeeper, redisStream)

	return listener.New(ethClient, handler), nil
}
