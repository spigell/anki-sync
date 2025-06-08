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
	getCmd.AddCommand(modelCmd.Command)

	return c
}

func (c *GetCmd) Command() *cobra.Command { return c.command }
func (c *GetCmd) SetFlags()               {}
func (c *GetCmd) Validate() error         { return nil }

// get model command.
type GetModelCmd struct {
	ctx    context.Context
	logger *logging.Logger
	name   string

	Command *cobra.Command
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
	g.Command = cmd
	return g
}

func (g *GetModelCmd) runE(_ *cobra.Command, _ []string) error {
	client := anki.NewClient(Config.AnkiURL)

	fields, err := client.GetModelFieldNames(g.ctx, g.name)
	if err != nil {
		return fmt.Errorf("get fields: %w", err)
	}

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
		InOrderFields: fields,
		CSS:           css,
		CardTemplates: templates,
	}

	out, err := yaml.Marshal(model)
	if err != nil {
		return err
	}

	fmt.Print(string(out))
	return nil
}
