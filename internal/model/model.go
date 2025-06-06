package model

import (
	"context"
	"fmt"

	"github.com/spigell/anki-sync/internal/anki"
	"github.com/spigell/anki-sync/internal/logging"
	"go.uber.org/zap"
)

type Manager struct {
	ctx    context.Context
	client *anki.Client
	dryRun bool
	logger *logging.Logger
	data   *anki.Data
}

type ManagerOption func(*Manager)

func NewModelManager(ctx context.Context, client *anki.Client, dryRun bool, logger *logging.Logger, data *anki.Data, opts ...ManagerOption) *Manager {
	m := &Manager{
		ctx:    ctx,
		client: client,
		dryRun: dryRun,
		logger: logger,
		data:   data,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Manager) Sync() error {
	for _, model := range m.data.Models {
		modelLogger := m.logger.CloneWith(zap.String("name", model.Name))
		select {
		case <-m.ctx.Done():
			return m.ctx.Err()
		default:
		}

		exists, err := m.client.ModelExists(m.ctx, model.Name)
		if err != nil {
			return fmt.Errorf("getting status model %s: %w", model.Name, err)
		}

		if m.dryRun {
			if !exists {
				modelLogger.DryRunLogger().Info("would create model")
			}
			for _, t := range model.CardTemplates {
				modelLogger.DryRunLogger().Info("would update template", zap.String("name", t.Name), zap.String("front", t.Front), zap.String("back", t.Back))
			}
			if model.CSS != "" {
				modelLogger.DryRunLogger().Info("would update css", zap.String("css", model.CSS))
			}
			continue
		}

		if !exists {
			modelLogger.Info("creating model")

			if err := m.client.CreateModel(m.ctx, model); err != nil {
				return fmt.Errorf("create model %s: %w", model.Name, err)
			}
		}

		modelLogger.Info("updating model")

		if err := m.client.UpdateModelTemplates(m.ctx, model.Name, model.CardTemplates); err != nil {
			return fmt.Errorf("update model templates `%s`: %w", model.Name, err)
		}

		if model.CSS != "" {
			if err := m.client.UpdateModelStyling(m.ctx, model.Name, model.CSS); err != nil {
				return fmt.Errorf("update model css %s: %w", model.Name, err)
			}
		}
	}

	return nil
}
