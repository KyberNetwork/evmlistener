package app

import (
	"context"
	"strconv"

	"github.com/KyberNetwork/evmlistener/internal/publisher"
	"github.com/KyberNetwork/evmlistener/pkg/block"
	"github.com/KyberNetwork/evmlistener/pkg/evmclient"
	"github.com/KyberNetwork/evmlistener/pkg/listener"
	"github.com/KyberNetwork/evmlistener/pkg/pubsub"
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

	wsRPC := c.String(wsRPCFlag.Name)
	wsEVMClient, err := evmclient.DialContext(context.Background(), wsRPC)
	if err != nil {
		l.Errorw("Fail to connect to node", "rpc", wsRPC, "error", err)

		return nil, err
	}

	httpRPC := c.String(httpRPCFlag.Name)
	httpEVMClient, err := evmclient.DialContext(context.Background(), httpRPC)
	if err != nil {
		l.Errorw("Fail to connect to node", "rpc", httpRPC, "error", err)

		return nil, err
	}

	chainID, err := httpEVMClient.ChainID(context.Background())
	if err != nil {
		l.Errorw("Fail to get chainID", "error", err)

		return nil, err
	}

	l = l.With("chainName", chainIDToName(chainID.Int64()))

	sanityCheckInterval := c.Duration(sanityCheckIntervalFlag.Name)
	var sanityEVMClient evmclient.IClient
	sanityRPC := c.String(sanityNodeRPCFlag.Name)
	if sanityRPC != "" {
		sanityEVMClient, err = evmclient.DialContext(context.Background(), sanityRPC)
		if err != nil {
			l.Errorw("Fail to setup EVM client for sanity check", "error", err)

			return nil, err
		}
	}

	redisConfig := redisConfigFromCli(c)
	redisClient, err := redis.New(redisConfig)
	if err != nil {
		l.Errorw("Fail to connect to redis", "cfg", redisConfig, "error", err)

		return nil, err
	}

	maxNumBlocks := c.Int(maxNumBlocksFlag.Name)
	blockExpiration := c.Duration(blockExpirationFlag.Name)
	blockKeeper := block.NewRedisBlockKeeper(l, redisClient, maxNumBlocks, blockExpiration)

	var publishSvc publisher.Publisher
	topic := c.String(publisherTopicFlag.Name)

	publisherType := c.String(publisherTypeFlag.Name)
	switch publisherType {
	case publisher.DataCentral.String():
		// pubsub message queue publisher
		orderingKey := c.String(pubsubOrderingKeyFlag.Name)
		projectID := c.String(pubsubProjectIDFlag.Name)

		pubsubCli, err := pubsub.NewPubsub(c.Context, projectID)
		if err != nil {
			return nil, err
		}

		publishSvc = publisher.NewDataCentralPublisher(pubsubCli, publisher.Config{
			Topic:       topic,
			OrderingKey: orderingKey,
		})

	default:
		maxLen := c.Int64(publisherMaxLenFlag.Name)
		streamCli := redis.NewStream(redisClient, maxLen)

		publishSvc = publisher.NewRedisStreamPublisher(streamCli, publisher.Config{
			Topic: topic,
		})
	}

	handler := listener.NewHandler(l, topic, httpEVMClient, blockKeeper, publishSvc)

	return listener.New(l, wsEVMClient, httpEVMClient, handler, sanityEVMClient, sanityCheckInterval), nil
}

const (
	chainIDEthereum     = 1
	chainIDOptimism     = 10
	chainIDCronos       = 25
	chainIDBSC          = 56
	chainIDVelas        = 106
	chainIDPolygon      = 137
	chainIDBitTorrent   = 199
	chainIDFantom       = 250
	chainIDZKSyncEra    = 324
	chainIDPolygonZKEVM = 1101
	chainIDBase         = 8453
	chainIDArbitrum     = 42161
	chainIDOasis        = 42262
	chainIDAvalanche    = 43114
	chainIDLinea        = 59144
	chainIDAurora       = 1313161554
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
	case chainIDLinea:
		return "Linea"
	case chainIDPolygonZKEVM:
		return "Polygon zkEVM"
	case chainIDZKSyncEra:
		return "zkSync Era"
	case chainIDBase:
		return "Base"
	default:
		return strconv.FormatInt(chainID, 10)
	}
}
