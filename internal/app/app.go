package app

import (
	"net/http"
	"time"

	"github.com/KyberNetwork/evmlistener/pkg/block"
	"github.com/KyberNetwork/evmlistener/pkg/encoder"
	"github.com/KyberNetwork/evmlistener/pkg/evmclient"
	"github.com/KyberNetwork/evmlistener/pkg/listener"
	publisherpkg "github.com/KyberNetwork/evmlistener/pkg/publisher"
	"github.com/KyberNetwork/evmlistener/pkg/publisher/kafka"
	"github.com/KyberNetwork/evmlistener/pkg/redis"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

const (
	defaultRequestTimeout = 10 * time.Second
)

// NewApp creates a new cli App instance with common flags preloaded.
func NewApp() *cli.App {
	app := cli.NewApp()
	app.Flags = NewFlags()

	return app
}

func redisConfigFromCli(c *cli.Context) redis.Config {
	cfg := redis.Config{
		MasterName:   redisMasterNameFlag.Value,
		Addrs:        redisAddrsFlag.Get(c),
		DB:           redisDBFlag.Value,
		KeyPrefix:    redisKeyPrefixFlag.Value,
		ReadTimeout:  redisReadTimeoutFlag.Value,
		WriteTimeout: redisWriteTimeoutFlag.Value,
	}

	cfg.SentinelUsername = redisUsernameFlag.Value
	cfg.SentinelPassword = redisPasswordFlag.Value
	cfg.Username = redisUsernameFlag.Value
	cfg.Password = redisPasswordFlag.Value

	return cfg
}

// NewListener setups and returns listener service.
func NewListener(c *cli.Context) (*listener.Listener, error) {
	l := zap.S()

	evmclient.UseCustomClient = useCustomClientFlag.Value

	rpcRequestTimeout := rpcRequestTimeoutFlag.Value
	if rpcRequestTimeout == 0 {
		rpcRequestTimeout = defaultRequestTimeout
	}

	httpClient := &http.Client{
		Timeout: rpcRequestTimeout,
	}
	wsRPC := wsRPCFlag.Value
	l.Infow("Connect to node websocket rpc", "rpc", wsRPC)
	wsEVMClient, err := evmclient.DialContextWithTimeout(
		c.Context, wsRPC, httpClient, rpcRequestTimeout)
	if err != nil {
		l.Errorw("Fail to connect to node", "rpc", wsRPC, "error", err)

		return nil, err
	}

	httpRPC := httpRPCFlag.Value
	l.Infow("Connect to node http rpc", "rpc", httpRPC)
	httpEVMClient, err := evmclient.DialContextWithTimeout(
		c.Context, httpRPC, httpClient, rpcRequestTimeout)
	if err != nil {
		l.Errorw("Fail to connect to node", "rpc", httpRPC, "error", err)

		return nil, err
	}

	l.Infow("Get chainID from node")
	chainID, err := httpEVMClient.ChainID(c.Context)
	if err != nil {
		l.Errorw("Fail to get chainID", "error", err)

		return nil, err
	}

	l = l.With("chainID", chainID.Int64())

	sanityCheckInterval := sanityCheckIntervalFlag.Value
	var sanityEVMClient evmclient.IClient
	sanityRPC := sanityNodeRPCFlag.Value
	if sanityRPC != "" {
		l.Infow("Connect to public node rpc for sanity check", "rpc", sanityRPC)
		sanityEVMClient, err = evmclient.DialContext(c.Context, sanityRPC, httpClient)
		if err != nil {
			l.Errorw("Fail to setup EVM client for sanity check", "error", err)

			return nil, err
		}
	}

	redisConfig := redisConfigFromCli(c)
	redisConfigForLog := redisConfig
	redisConfigForLog.SentinelPassword = "***"
	redisConfigForLog.Password = "***"
	l.Infow("Connect to redis", "cfg", redisConfigForLog)
	redisClient, err := redis.New(redisConfig)
	if err != nil {
		l.Errorw("Fail to connect to redis", "cfg", redisConfigForLog, "error", err)

		return nil, err
	}

	maxNumBlocks := maxNumBlocksFlag.Value
	blockExpiration := blockExpirationFlag.Value
	l.Infow("Setup new BlockKeeper", "maxNumBlocks", maxNumBlocks, "expiration", blockExpiration)
	blockKeeper := block.NewRedisBlockKeeper(l, redisClient, maxNumBlocks, blockExpiration)

	topic := publisherTopicFlag.Value
	publisher, err := getPublisher(c, redisClient, topic)
	if err != nil {
		l.Errorw("Fail to get publisher", "error", err)

		return nil, err
	}
	msgEncoder := getMessageEncoder()

	l.Infow("Setup handler", "topic", topic)
	handler := listener.NewHandler(listener.HandlerConfig{BlockSlowWarningThreshold: blockSlowWarningThresholdFlag.Value},
		l, topic, httpEVMClient, blockKeeper, publisher, msgEncoder,
		listener.WithEventLogs(nil, nil))

	l.Infow("Setup listener")

	return listener.New(l, wsEVMClient, httpEVMClient, handler, sanityEVMClient, sanityCheckInterval,
		listener.WithEventLogs(nil, nil)), nil
}

func getPublisher(c *cli.Context, redisClient *redis.Client, topic string) (publisherpkg.Publisher, error) {
	var publisher publisherpkg.Publisher
	var err error

	publisherType := publisherTypeFlag.Value
	switch publisherType {
	case publisherpkg.PublisherTypeKafka:
		config := &kafka.Config{
			Addresses:         kafkaAddrsFlag.Get(c),
			UseAuthentication: kafkaUseAuthenticationFlag.Value,
			Username:          kafkaUsernameFlag.Value,
			Password:          kafkaPasswordFlag.Value,
		}
		publisher, err = kafka.NewPublisher(config)
		if err != nil {
			return nil, err
		}
		if err := kafka.ValidateTopicName(topic); err != nil {
			return nil, err
		}
	default:
		maxLen := publisherMaxLenFlag.Value
		publisher = redis.NewStream(redisClient, maxLen)
	}

	return publisher, err
}

func getMessageEncoder() encoder.Encoder {
	switch encoderTypeFlag.Value {
	case encoder.EncoderTypeProtobuf:
		return encoder.NewProtobufEncoder()
	default:
		return encoder.NewJSONEncoder()
	}
}
