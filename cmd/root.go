package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/spigell/anki-sync/internal/anki"
	"github.com/spigell/anki-sync/internal/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	Instance *logging.Logger
	Level    zap.AtomicLevel
}

type AppConfig struct {
	Decks             string `mapstructure:"decks"`
	Models            string `mapstructure:"models"`
	AnkiURL           string `mapstructure:"anki_url"`
	AnkiAPIKey        string `mapstructure:"anki_api_key"`
	Recursive         bool   `mapstructure:"recursive"`
	UploadParallelism int    `mapstructure:"upload_parallelism"`
	DryRun            bool   `mapstructure:"dry_run"`
	LogLevel          string `mapstructure:"log_level"`
}

var (
	cfgFile string
	Config  = &AppConfig{} // populated during PersistentPreRunE
)

const (
	DefaultConfigFile = "anki-sync.yaml"
)

type ValidatedCommand interface {
	Command() *cobra.Command
	Validate() error
	SetFlags()
}

func NewRootCmd(ctx context.Context, logger *Logger) *cobra.Command {
	commands := []ValidatedCommand{
		NewSyncCmd(ctx, logger.Instance),
		NewGetCmd(ctx, logger.Instance),
		NewAnkiWebSyncCmd(ctx, logger.Instance),
		NewVersionCmd(ctx, logger.Instance.Logger),
	}

	rootCmd := &cobra.Command{
		Use:   "anki-sync",
		Short: "Sync Anki notes, decks, and models from YAML via AnkiConnect",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := initConfig(); err != nil {
				missing := os.IsNotExist(err)

				if !missing {
					// Any other error is fatal
					return err
				}

				// Only fail if the user explicitly provided a config file
				if missing && cfgFile != DefaultConfigFile {
					return err
				}

				// Otherwise, ignore missing config file
			}

			if err := viper.Unmarshal(Config); err != nil {
				return err
			}

			if Config.LogLevel != "" {
				var lvl zapcore.Level
				if err := lvl.UnmarshalText([]byte(Config.LogLevel)); err != nil {
					return fmt.Errorf("invalid log level %q: %w", Config.LogLevel, err)
				}
				logger.Level.SetLevel(lvl)
				logger.Instance.Info("set logLevel", zap.String("level", lvl.String()))
			}

			logger.Instance.Debug("using AnkiConnect", zap.String("url", Config.AnkiURL))

			if Config.DryRun {
				logger.Instance.EnableDryRunLogger()
				logger.Instance.DryRunLogger().Info("dryRunLogger is enabled")
			}

			// Run Validate() on matching command only
			for _, vc := range commands {
				if vc.Command().Name() == cmd.Name() {
					if err := vc.Validate(); err != nil {
						return err
					}
				}
			}

			return nil
		},
	}

	for _, cmd := range commands {
		cmd.SetFlags()
		rootCmd.AddCommand(cmd.Command())
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", DefaultConfigFile, "Config file")
	rootCmd.PersistentFlags().String("anki-url", "http://127.0.0.1:8765", "AnkiConnect API URL")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Simulate sync actions")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")

	viper.BindPFlag("anki_url", rootCmd.PersistentFlags().Lookup("anki-url"))
	viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("dry_run", rootCmd.PersistentFlags().Lookup("dry-run"))

	viper.SetEnvPrefix("anki_sync")
	viper.AutomaticEnv()
	// Explicit env name skips the prefix. Env wins over the config file.
	viper.BindEnv("anki_api_key", "ANKI_API_KEY")
	viper.BindEnv("anki_url", "ANKI_URL")

	return rootCmd
}

func newAnkiClient(opts ...anki.Option) *anki.Client {
	return anki.NewClient(Config.AnkiURL, append([]anki.Option{anki.WithAPIKey(Config.AnkiAPIKey)}, opts...)...)
}

func initConfig() error {
	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}
