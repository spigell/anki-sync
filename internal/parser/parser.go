package parser

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/spigell/anki-sync/internal/anki"
)

type DeckParsed struct {
	Deck   anki.Deck
	Parsed bool
	Path   string
}

var ErrDeckIsNotParseble = errors.New("deck file is not parseble")

func LoadModels(path string) ([]anki.Model, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var wrap struct {
		Models []anki.Model `yaml:"models"`
	}
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&wrap); err != nil {
		return nil, fmt.Errorf("failed to parse models: %w", err)
	}
	return wrap.Models, nil
}

//nolint:gocognit // To do.
func LoadDecks(path string, recursive bool) ([]DeckParsed, error) {
	var decks []DeckParsed

	processFile := func(p string) error {
		if ext := filepath.Ext(p); ext != ".yaml" && ext != ".yml" {
			return nil
		}

		f, err := os.Open(p)
		if err != nil {
			return fmt.Errorf("could not open deck file %s: %w", p, err)
		}
		defer f.Close()

		decks = append(decks, DeckParsed{Path: p})

		var deck anki.Deck
		dec := yaml.NewDecoder(f)
		dec.KnownFields(true)
		if err := dec.Decode(&deck); err != nil {
			return ErrDeckIsNotParseble
		}
		decks[len(decks)-1].Parsed = true
		decks[len(decks)-1].Deck = deck

		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Just file
	if !info.IsDir() {
		if err := processFile(path); err != nil && !errors.Is(err, ErrDeckIsNotParseble) {
			return nil, err
		}
		return decks, nil
	}

	// Directory. Walk in recursive way
	//nolint:nestif
	if recursive {
		err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			err = processFile(p)
			if errors.Is(err, ErrDeckIsNotParseble) {
				return nil
			}
			return err
		})
	} else {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if err := processFile(filepath.Join(path, entry.Name())); err != nil {
				if errors.Is(err, ErrDeckIsNotParseble) {
					continue
				}
				return nil, err
			}
		}
	}

	if err != nil {
		return nil, err
	}
	return decks, nil
}

func ValidateNotes(_ []anki.Deck, _ []anki.Model) []error {
	// TODO: Check fields, model existence, etc.
	return nil
}
