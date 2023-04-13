package app

import (
	"fmt"
	"io"
	"os"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type syncer interface {
	Sync() error
}

// NewFlusher creates a new syncer from given syncer that log a error message if failed to sync.
func NewFlusher(s syncer) func() {
	return func() {
		// ignore the error as the sync function will always fail in Linux
		// https://github.com/uber-go/zap/issues/370
		_ = s.Sync()
	}
}

// newLogger creates a new logger instance.
// The type of logger instance will be different with different application running modes.
func newLogger(c *cli.Context) (*zap.Logger, zap.AtomicLevel) {
	writers := []io.Writer{os.Stdout}

	w := io.MultiWriter(writers...)

	logLevel, err := zapcore.ParseLevel(c.String(logLevelFlag.Name))
	if err != nil {
		panic(err)
	}

	atom := zap.NewAtomicLevelAt(logLevel)

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	config.CallerKey = "caller"

	encoder := zapcore.NewConsoleEncoder(config)
	cc := zap.New(zapcore.NewCore(encoder, zapcore.AddSync(w), atom), zap.AddCaller())

	return cc, atom
}

// NewLogger creates a new sugared logger and a flush function. The flush function should be
// called by consumer before quitting application.
// This function should be use most of the time unless
// the application requires extensive performance, in this case use NewLogger.
func NewLogger(c *cli.Context) (*zap.Logger, zap.AtomicLevel, func(), error) {
	logger, atom := newLogger(c)

	// init sentry if flag dsn exists
	if len(c.String(sentryDSNFlag.Name)) != 0 {
		sentryClient, err := sentry.NewClient(
			sentry.ClientOptions{
				Dsn: c.String(sentryDSNFlag.Name),
			},
		)
		if err != nil {
			return nil, atom, nil, fmt.Errorf("failed to init sentry client: %w", err)
		}

		cfg := zapsentry.Configuration{
			DisableStacktrace: false,
		}

		sentryLevel, err := zapcore.ParseLevel(c.String(sentryLevelFlag.Name))
		if err != nil {
			return nil, atom, nil, err
		}

		cfg.Level = sentryLevel

		core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromClient(sentryClient))
		if err != nil {
			return nil, atom, nil, fmt.Errorf("failed to init zap sentry: %w", err)
		}
		// attach to logger core
		logger = zapsentry.AttachCoreToLogger(core, logger)
	}

	return logger, atom, NewFlusher(logger), nil
}
