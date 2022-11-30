package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	libapp "github.com/KyberNetwork/evmlistener/internal/app"
	"github.com/KyberNetwork/evmlistener/pkg/block"
	"github.com/KyberNetwork/evmlistener/pkg/listener"
	"github.com/KyberNetwork/evmlistener/pkg/redis"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func main() {
	app := libapp.NewApp()
	app.Name = "EVM compatible listener service"
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func run(c *cli.Context) error {
	logger, _, flush, err := libapp.NewLogger(c)
	if err != nil {
		return fmt.Errorf("new logger: %w", err)
	}

	defer flush()

	zap.ReplaceGlobals(logger)
	l := logger.Sugar()
	l.Infow("App starting ..")
	defer l.Infow("App stopped!")

	listener, err := setupListener(c)
	if err != nil {
		l.Errorw("Fail to setup Listener service", "error", err)

		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	return listener.Run(ctx)
}

func setupListener(_ *cli.Context) (*listener.Listener, error) {
	l := zap.S()

	rpc := "wss://polygon.kyberengineering.io"
	ethClient, err := ethclient.DialContext(context.Background(), rpc)
	if err != nil {
		l.Errorw("Fail to connect to node", "rpc", rpc, "error", err)

		return nil, err
	}

	redisConfig := redis.Config{
		Addrs:     []string{""},
		DB:        0,
		KeyPrefix: "test-listener-polygon:",
	}
	redisClient, err := redis.New(redisConfig)
	if err != nil {
		l.Errorw("Fail to connect to redis", "cfg", redisConfig, "error", err)

		return nil, err
	}

	blockKeeper := block.NewRedisBlockKeeper(redisClient, 128, time.Hour)
	redisStream := redis.NewStream(redisClient, 10000)
	handler := listener.NewHandler("test-polygon-stream", ethClient, blockKeeper, redisStream)

	return listener.New(ethClient, handler), nil
}
