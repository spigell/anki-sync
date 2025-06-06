package cmd

import (
	"context"
	"fmt"

	"github.com/spigell/anki-sync/internal/anki"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var CliVersion = "unknown"

type VersionCmd struct {
	command *cobra.Command
}

func NewVersionCmd(ctx context.Context, logger *zap.Logger) *VersionCmd {
	return &VersionCmd{command: &cobra.Command{
		Use:   "version",
		Short: "Show CLI and AnkiConnect version",
		RunE: func(_ *cobra.Command, _ []string) error {
			client := anki.NewClient(Config.AnkiURL)
			ver, err := client.GetVersion(ctx)
			if err != nil {
				logger.Warn("AnkiConnect version fetch failed", zap.Error(err))
				ver = "unknown"
			}
			fmt.Printf(`anki-sync version: %s
API AnkiConnect version: %s
`, CliVersion, ver)
			return err
		},
	}}
}

func (c *VersionCmd) Command() *cobra.Command {
	return c.command
}

func (c *VersionCmd) SetFlags()       {}
func (c *VersionCmd) Validate() error { return nil }
