package app

import (
	"time"

	"github.com/urfave/cli/v2"
)

//nolint:gochecknoglobals
var (
	logLevelFlag = &cli.StringFlag{
		Name:    "log-level",
		EnvVars: []string{"LOG_LEVEL"},
		Value:   "info",
		Usage:   "Set log level for logger, values: debug, info, warn, error. Default: info",
	}
	nodeRPCFlag = &cli.StringFlag{
		Name:    "node-rpc",
		EnvVars: []string{"NODE_RPC"},
		Value:   "ws://localhost:8546",
		Usage:   "Websocket rpc to connect to blockchain node, default: ws://localhost:8546",
	}

	sentryDSNFlag = &cli.StringFlag{
		Name:    "sentry-dsn",
		EnvVars: []string{"SENTRY_DSN"},
		Usage:   "DSN for sentry client",
	}
	sentryLevelFlag = &cli.StringFlag{
		Name:    "sentry-level",
		EnvVars: []string{"SENTRY_LEVEL"},
		Usage:   "log level report message to sentry (info, error, warn, fatal)",
		Value:   "error",
	}

	redisMasterNameFlag = &cli.StringFlag{
		Name:    "redis-master-name",
		EnvVars: []string{"REDIS_MASTER_NAME"},
		Value:   "",
		Usage:   "Master name for redis sentinel",
	}
	redisAddrsFlag = &cli.StringSliceFlag{
		Name:    "redis-addrs",
		EnvVars: []string{"REDIS_ADDRS"},
		Value:   cli.NewStringSlice("localhost:6379"),
		Usage:   "A list of address for connecting to redis. Default: localhost:6379",
	}
	redisDBFlag = &cli.IntFlag{
		Name:    "redis-db",
		EnvVars: []string{"REDIS_DB"},
		Value:   0,
		Usage:   "Redis database index",
	}
	redisUsernameFlag = &cli.StringFlag{
		Name:    "redis-username",
		EnvVars: []string{"REDIS_USERNAME"},
		Value:   "",
		Usage:   "Username for authenticating with redis server",
	}
	redisPasswordFlag = &cli.StringFlag{
		Name:    "redis-password",
		EnvVars: []string{"REDIS_PASSWORD"},
		Value:   "",
		Usage:   "Password for authenticating with redis server",
	}
	redisKeyPrefixFlag = &cli.StringFlag{
		Name:    "redis-key-prefix",
		EnvVars: []string{"REDIS_KEY_PREFIX"},
		Value:   "",
		Usage:   "Prefix of key for redis",
	}
	redisReadTimeoutFlag = &cli.DurationFlag{
		Name:    "redis-read-timeout",
		EnvVars: []string{"REDIS_READ_TIMEOUT"},
		Value:   0,
		Usage:   "Timeout for redis read operation",
	}
	redisWriteTimeoutFlag = &cli.DurationFlag{
		Name:    "redis-write-timeout",
		EnvVars: []string{"REDIS_WRITE_TIMEOUT"},
		Value:   0,
		Usage:   "Timeout for redis write operation",
	}

	publisherTopicFlag = &cli.StringFlag{
		Name:     "publisher-topic",
		EnvVars:  []string{"PUBLISHER_TOPIC"},
		Value:    "",
		Required: true,
		Usage:    "Topic name of publisher to publish message to (Required)",
	}
	publisherMaxLenFlag = &cli.Int64Flag{
		Name:    "publisher-max-len",
		EnvVars: []string{"PUBLISHER_MAX_LEN"},
		Value:   7200, //nolint:gomnd
		Usage:   "Maximum length for publisher's queue. Default: 7200",
	}

	maxNumBlocksFlag = &cli.IntFlag{
		Name:    "max-num-blocks",
		EnvVars: []string{"MAX_NUM_BLOCKS"},
		Value:   64, //nolint:gomnd
		Usage:   "Maximum number of blocks for block keeper. Default: 64",
	}
	blockExpirationFlag = &cli.DurationFlag{
		Name:    "block-expiration",
		EnvVars: []string{"BLOCK_EXPIRATION"},
		Value:   24 * time.Hour, //nolint:gomnd
		Usage:   "Expiration time for storing block into datastore. Default: 24h",
	}
)

// NewSentryFlags returns flags to init sentry client.
func NewSentryFlags() []cli.Flag {
	return []cli.Flag{sentryDSNFlag, sentryLevelFlag}
}

// NewRedisFlags returns flags for redis.
func NewRedisFlags() []cli.Flag {
	return []cli.Flag{
		redisMasterNameFlag, redisAddrsFlag, redisDBFlag,
		redisUsernameFlag, redisPasswordFlag, redisKeyPrefixFlag,
		redisReadTimeoutFlag, redisWriteTimeoutFlag,
	}
}

// NewPublisherFlags returns flags for publishers.
func NewPublisherFlags() []cli.Flag {
	return []cli.Flag{publisherMaxLenFlag, publisherTopicFlag}
}

// NewBlockKeeperFlags returns flags for block keeper.
func NewBlockKeeperFlags() []cli.Flag {
	return []cli.Flag{maxNumBlocksFlag, blockExpirationFlag}
}

// NewFlags returns all flags for the application.
func NewFlags() []cli.Flag {
	flags := []cli.Flag{logLevelFlag, nodeRPCFlag}
	flags = append(flags, NewSentryFlags()...)
	flags = append(flags, NewRedisFlags()...)
	flags = append(flags, NewPublisherFlags()...)
	flags = append(flags, NewBlockKeeperFlags()...)

	return flags
}
