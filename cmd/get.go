package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/spigell/anki-sync/internal/anki"
	"github.com/spigell/anki-sync/internal/logging"
)

// GetCmd represents the top level `get` command.
type GetCmd struct {
	command *cobra.Command
}

func NewGetCmd(ctx context.Context, logger *logging.Logger) *GetCmd {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve data from Anki",
	}

	c := &GetCmd{command: getCmd}

	// add subcommands
	modelCmd := newGetModelCmd(ctx, logger)
	getCmd.AddCommand(modelCmd.Command())

	return c
}

func (c *GetCmd) Command() *cobra.Command { return c.command }
func (c *GetCmd) SetFlags()               {}
func (c *GetCmd) Validate() error         { return nil }

// get model command

type GetModelCmd struct {
	command *cobra.Command
	ctx     context.Context
	logger  *logging.Logger
	name    string
}

func newGetModelCmd(ctx context.Context, logger *logging.Logger) *GetModelCmd {
	g := &GetModelCmd{ctx: ctx, logger: logger}
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Get model templates and styling by name",
		RunE:  g.runE,
	}
	cmd.Flags().StringVar(&g.name, "name", "", "Model name")
	cmd.MarkFlagRequired("name")
	g.command = cmd
	return g
}

func (g *GetModelCmd) Command() *cobra.Command { return g.command }
func (g *GetModelCmd) SetFlags()               {}
func (g *GetModelCmd) Validate() error {
	if g.name == "" {
		return fmt.Errorf("--name is required")
	}
	return nil
}

func (g *GetModelCmd) runE(cmd *cobra.Command, args []string) error {
	client := anki.NewClient(Config.AnkiURL)

	templates, err := client.GetModelTemplates(g.ctx, g.name)
	if err != nil {
		return fmt.Errorf("get templates: %w", err)
	}

	css, err := client.GetModelStyling(g.ctx, g.name)
	if err != nil {
		return fmt.Errorf("get styling: %w", err)
	}

	model := anki.Model{
		Name:          g.name,
		CSS:           css,
		CardTemplates: templates,
	}

	wrap := struct {
		Models []anki.Model `yaml:"models"`
	}{Models: []anki.Model{model}}

	out, err := yaml.Marshal(wrap)
	if err != nil {
		return err
	}

	fmt.Print(string(out))
	return nil
}
