package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spigell/anki-sync/internal/anki"
	"github.com/spigell/anki-sync/internal/logging"
)

const defaultAnkiWebSyncTimeout = 120 * time.Second

// AnkiWebSyncCmd asks Anki to sync the open profile with AnkiWeb.
type AnkiWebSyncCmd struct {
	command *cobra.Command
	timeout time.Duration
}

func NewAnkiWebSyncCmd(ctx context.Context, logger *logging.Logger) *AnkiWebSyncCmd {
	c := &AnkiWebSyncCmd{}
	c.command = &cobra.Command{
		Use:   "ankiweb-sync",
		Short: "Sync the Anki profile with AnkiWeb",
		Long: `Sync the Anki profile with AnkiWeb through AnkiConnect's sync action.

The first sync of a profile may require choosing a full sync (upload or
download) in the Anki GUI. AnkiConnect cannot answer that choice, so the
command fails with AnkiConnect's error or times out instead of choosing.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			if Config.DryRun {
				logger.DryRunLogger().Info("skipping AnkiWeb sync")
				return nil
			}

			ctx, cancel := context.WithTimeout(ctx, c.timeout)
			defer cancel()

			err := newAnkiClient(anki.WithTimeout(c.timeout)).Sync(ctx)
			if errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("ankiweb sync timed out after %s: %w", c.timeout, err)
			}
			if err != nil {
				return fmt.Errorf("ankiweb sync failed: %w", err)
			}

			logger.Info("ankiweb sync done")
			return nil
		},
	}
	return c
}

func (c *AnkiWebSyncCmd) Command() *cobra.Command {
	return c.command
}

func (c *AnkiWebSyncCmd) SetFlags() {
	c.command.Flags().DurationVar(&c.timeout, "timeout", defaultAnkiWebSyncTimeout, "How long to wait for the sync to finish")
}

func (c *AnkiWebSyncCmd) Validate() error {
	if c.timeout <= 0 {
		return fmt.Errorf("--timeout must be greater than 0")
	}
	return nil
}
