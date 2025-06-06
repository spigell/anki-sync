package cmd

import (
	"context"
	"fmt"
	"runtime"

	"github.com/spigell/anki-sync/internal/anki"
	"github.com/spigell/anki-sync/internal/deck"
	"github.com/spigell/anki-sync/internal/logging"
	"go.uber.org/zap"

	"github.com/spigell/anki-sync/internal/model"
	"github.com/spigell/anki-sync/internal/parser"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SyncCmd struct {
	command *cobra.Command
}

func NewSyncCmd(ctx context.Context, logger *logging.Logger) *SyncCmd {
	return &SyncCmd{
		&cobra.Command{
			Use:   "sync",
			Short: "Sync notes, models, and decks with Anki",
			RunE: func(_ *cobra.Command, _ []string) error {
				ms, err := parser.LoadModels(Config.Models)
				if err != nil {
					return err
				}

				ns, err := parser.LoadDecks(Config.Decks, Config.Recursive)
				if err != nil {
					return err
				}

				var validDeckFiles []string
				var invalidDeckFiles []string
				var decks []anki.Deck
				for _, d := range ns {
					if !d.Parsed {
						invalidDeckFiles = append(invalidDeckFiles, d.Path)
						continue
					}
					validDeckFiles = append(validDeckFiles, d.Path)
					decks = append(decks, d.Deck)
				}

				logger.Info("parsed decks", zap.Any("files", validDeckFiles))
				
				if len(invalidDeckFiles) > 0 {
					logger.Warn("invalid decks files. They are skipped", zap.Any("files", invalidDeckFiles))
				}

				client := anki.NewClient(Config.AnkiURL)

				if err := model.NewModelManager(ctx, client, Config.DryRun, logger, &anki.Data{
					Models: ms,
				}).Sync(); err != nil {
					return fmt.Errorf("model sync failed: %w", err)
				}

				if err := deck.NewDeckManager(ctx, client, Config.DryRun, logger, &anki.Data{
					Models: ms,
					Decks:  decks,
				}, deck.WithNoteUploadParallelism(Config.UploadParallelism)).Sync(); err != nil {
					return fmt.Errorf("decks sync failed: %w", err)
				}

				logger.Info("note sync done")
				return nil
			},
		},
	}
}

func (c *SyncCmd) Command() *cobra.Command {
	return c.command
}

func (c *SyncCmd) SetFlags() {
	flags := c.command.PersistentFlags()

	c.command.PersistentFlags().String("decks", "", "Path to notes YAML file or directory (required)")
	c.command.PersistentFlags().String("models", "", "Path to models YAML file (required)")
	c.command.PersistentFlags().Bool("recursive", false, "Recurse into directories for notes")
	c.command.PersistentFlags().Int("upload-parallelism", runtime.NumCPU(), "Concurrent note uploads per file")

	viper.BindPFlag("models", c.command.PersistentFlags().Lookup("models"))
	viper.BindPFlag("decks", flags.Lookup("decks"))
	viper.BindPFlag("recursive", c.command.PersistentFlags().Lookup("recursive"))
	viper.BindPFlag("upload_parallelism", c.command.PersistentFlags().Lookup("upload-parallelism"))
}

func (c *SyncCmd) Validate() error {
	if Config.Models == "" {
		return fmt.Errorf("--models or config.models must be set")
	}
	if Config.Decks == "" {
		return fmt.Errorf("--decks or config.decks must be set")
	}

	if Config.UploadParallelism < 1 {
		return fmt.Errorf("--upload-parallelism or config.upload-parallelism must be greater or equal 1")
	}
	return nil
}
