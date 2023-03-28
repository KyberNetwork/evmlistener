package app

import (
	"context"
	"strconv"

	"github.com/KyberNetwork/evmlistener/pkg/block"
	"github.com/KyberNetwork/evmlistener/pkg/evmclient"
	"github.com/KyberNetwork/evmlistener/pkg/listener"
	"github.com/KyberNetwork/evmlistener/pkg/redis"
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
	ethClient, err := evmclient.DialContext(context.Background(), rpc)
	if err != nil {
		l.Errorw("Fail to connect to node", "rpc", rpc, "error", err)

		return nil, err
	}

	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		l.Errorw("Fail to get chainID", "error", err)

		return nil, err
	}

	l = l.With("chainName", chainIDToName(chainID.Int64()))

	redisConfig := redisConfigFromCli(c)
	redisClient, err := redis.New(redisConfig)
	if err != nil {
		l.Errorw("Fail to connect to redis", "cfg", redisConfig, "error", err)

		return nil, err
	}

	maxNumBlocks := c.Int(maxNumBlocksFlag.Name)
	blockExpiration := c.Duration(blockExpirationFlag.Name)
	blockKeeper := block.NewRedisBlockKeeper(l, redisClient, maxNumBlocks, blockExpiration)

	maxLen := c.Int64(publisherMaxLenFlag.Name)
	redisStream := redis.NewStream(redisClient, maxLen)

	topic := c.String(publisherTopicFlag.Name)
	handler := listener.NewHandler(l, topic, ethClient, blockKeeper, redisStream)

	return listener.New(l, ethClient, handler), nil
}

const (
	chainIDEthereum   = 1
	chainIDOptimism   = 10
	chainIDCronos     = 25
	chainIDBSC        = 56
	chainIDVelas      = 106
	chainIDPolygon    = 137
	chainIDBitTorrent = 199
	chainIDFantom     = 250
	chainIDArbitrum   = 42161
	chainIDOasis      = 42262
	chainIDAvalanche  = 43114
	chainIDAurora     = 1313161554
)

//nolint:cyclop
func chainIDToName(chainID int64) string {
	switch chainID {
	case chainIDEthereum:
		return "Ethereum"
	case chainIDOptimism:
		return "Optimism"
	case chainIDCronos:
		return "Cronos"
	case chainIDBSC:
		return "BSC"
	case chainIDVelas:
		return "Velas"
	case chainIDPolygon:
		return "Polygon"
	case chainIDBitTorrent:
		return "BitTorrent"
	case chainIDFantom:
		return "Fantom"
	case chainIDArbitrum:
		return "Arbitrum"
	case chainIDOasis:
		return "Oasis"
	case chainIDAvalanche:
		return "Avalanche"
	case chainIDAurora:
		return "Aurora"
	default:
		return strconv.FormatInt(chainID, 10)
	}
}
