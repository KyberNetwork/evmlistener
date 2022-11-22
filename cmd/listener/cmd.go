package main

import (
	"fmt"
	"log"
	"os"

	libapp "github.com/KyberNetwork/evmlistener/pkg/app"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

func main() {
	app := libapp.NewApp()
	app.Name = "Listener Service"
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		log.Panic(err)
	}
}

//nolint:funlen
func run(c *cli.Context) error {
	logger, _, flush, err := libapp.NewLogger(c)
	if err != nil {
		return fmt.Errorf("new logger: %w", err)
	}

	defer flush()

	zap.ReplaceGlobals(logger)
	l := logger.Sugar()
	l.Infow("App starting ..")

	l.Infow("App stopped!")

	return nil
}
