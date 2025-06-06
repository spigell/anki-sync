package deck

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/spigell/anki-sync/internal/anki"
	"github.com/spigell/anki-sync/internal/logging"
	"github.com/spigell/anki-sync/internal/workerpool"
	"go.uber.org/zap"
)

const NoteTag = "anki-sync"

type Manager struct {
	ctx      context.Context
	client   *anki.Client
	dryRun   bool
	logger   *logging.Logger
	data     *anki.Data
	parallel int
}

type ManagerOption func(*Manager)

func NewDeckManager(ctx context.Context, client *anki.Client, dryRun bool, logger *logging.Logger, data *anki.Data, opts ...ManagerOption) *Manager {
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

func WithNoteUploadParallelism(n int) ManagerOption {
	return func(m *Manager) {
		m.parallel = n
	}
}

//nolint:gocognit // To do.
func (m *Manager) Sync() error {
	if len(m.data.Decks) == 0 {
		m.logger.Info("no decks to sync")
		return nil
	}

	var (
		seen   = make(map[string]bool)
		seenMu sync.Mutex
		wg     sync.WaitGroup

		errsMu sync.Mutex
		errs   []error
	)

	for _, deck := range m.data.Decks {
		wg.Add(1)
		go func(deck anki.Deck) {
			defer wg.Done()

			deckLogger := m.logger.CloneWith(zap.String("deck", deck.Deck))

			// Must be moved to the validate layer.
			seenMu.Lock()
			if seen[deck.Deck] {
				m.logger.Warn("a duplicate deck found. Consider to merge it to the single file", zap.String("deck", deck.Deck))
			}
			seen[deck.Deck] = true
			seenMu.Unlock()

			exists, err := m.client.DeckExists(m.ctx, deck.Deck)
			if err != nil {
				errsMu.Lock()
				errs = append(errs, err)
				errsMu.Unlock()
				return
			}

			if m.dryRun {
				if !exists {
					deckLogger.DryRunLogger().Info("would create deck", zap.String("deck", deck.Deck))
				}
			}

			if !exists && !m.dryRun {
				if err := m.client.CreateDeck(m.ctx, deck.Deck); err != nil {
					m.logger.Error("Failed to create deck", zap.String("deck", deck.Deck), zap.Error(err))
					errsMu.Lock()
					errs = append(errs, err)
					errsMu.Unlock()
					return
				}
			}

			deckLogger.Info("launch new workerpool for uploading notes", zap.Int("worker_count", m.parallel))
			pool := workerpool.New(m.parallel)
			pool.Start(m.ctx)

			for _, note := range deck.Notes {
				n := note // capture range var
				pool.Submit(func(ctx context.Context) error {
					err := m.ensureNote(ctx, deck, n, deckLogger)
					if err != nil {
						errsMu.Lock()
						errs = append(errs, err)
						errsMu.Unlock()
					}
					return err
				})
			}

			pool.Stop()
		}(deck)
	}

	wg.Wait()

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (m *Manager) ensureNote(ctx context.Context, deck anki.Deck, note anki.Note, logger *logging.Logger) error {
	note.Tags = append(note.Tags, NoteTag)

	exists, id, err := m.client.NoteExists(m.ctx, deck.Deck, fmt.Sprintf("%s:%s", deck.PrimaryField, note.Fields[deck.PrimaryField]))
	if err != nil {
		return fmt.Errorf("error while getting status of note: %w", err)
	}

	l := logger.CloneWith(zap.Int64("noteId", id))

	if exists {
		l.Info("note exists", zap.String("primary_field", deck.PrimaryField))
	}

	if m.dryRun {
		if !exists {
			l.DryRunLogger().Info("would create note", zap.Any("fields", note.Fields))
		}
		l.DryRunLogger().Info("would update note fields", zap.Any("fields", note.Fields))
		l.DryRunLogger().Info("would update note tags", zap.Any("tags", note.Tags))
		return nil
	}

	if !exists {
		if err := m.client.AddNote(ctx, deck.Deck, deck.Model, note); err != nil {
			return err
		}
		_, id, err = m.client.NoteExists(m.ctx, deck.Deck, fmt.Sprintf("%s:%s", deck.PrimaryField, note.Fields[deck.PrimaryField]))
		if err != nil {
			return fmt.Errorf("error while getting status of note after creation: %w", err)
		}
		logger.Info("note created", zap.Int64("noteId", id))
	}

	if err := m.client.UpdateNoteTags(ctx, id, note.Tags); err != nil {
		return fmt.Errorf("error while updating tags of note: %w", err)
	}

	if err := m.client.UpdateNoteFields(ctx, id, note.Fields); err != nil {
		return fmt.Errorf("error while updating fields of note: %w", err)
	}

	logger.Info("tags and fields updated", zap.Int64("noteId", id))
	return nil
}
