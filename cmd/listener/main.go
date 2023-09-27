package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	libapp "github.com/KyberNetwork/evmlistener/internal/app"
	_ "github.com/KyberNetwork/kyber-trace-go/tools"
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

	listener, err := libapp.NewListener(c)
	if err != nil {
		l.Errorw("Fail to setup Listener service", "error", err)

		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	return listener.Run(ctx)
}
