package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spigell/anki-sync/cmd"
	"github.com/spigell/anki-sync/internal/logging"
	"go.uber.org/zap"
)

// Default level.
var logLevel zap.AtomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := logging.NewLogger(logLevel)
	rootCmd := cmd.NewRootCmd(ctx, &cmd.Logger{Instance: logger, Level: logLevel})
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		logger.Error("cli error", zap.Error(err))
		return 1
	}
	return 0
}
