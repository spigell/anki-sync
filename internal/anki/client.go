package anki

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"time"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Transport: &http.Transport{
				// AnkiConnect sometimes drop connections from their pool.
				// Crate a new connection all the time.
				// It is ok for the tool.
				DisableKeepAlives: true,
			},
			Timeout: 10 * time.Second,
		},
	}
}

type request struct {
	Action  string `json:"action"`
	Version int    `json:"version"`
	Params  any    `json:"params,omitempty"`
}

type response struct {
	Result json.RawMessage `json:"result"`
	Error  *string         `json:"error"`
}

func (c *Client) GetVersion(ctx context.Context) (string, error) {
	var v int
	err := c.do(ctx, request{Action: "version", Version: 6}, &v)
	return fmt.Sprintf("%d", v), err
}

func (c *Client) ModelExists(ctx context.Context, name string) (bool, error) {
	var models []string
	err := c.do(ctx, request{
		Action:  "modelNames",
		Version: 6,
	}, &models)
	if err != nil {
		return false, err
	}
	if slices.Contains(models, name) {
		return true, nil
	}
	return false, nil
}

func (c *Client) CreateModel(ctx context.Context, m Model) error {
	return c.do(ctx, request{
		Action:  "createModel",
		Version: 6,
		Params:  m,
	}, nil)
}

func (c *Client) UpdateModelTemplates(ctx context.Context, name string, templates []CardTemplate) error {
	tmpls := make(map[string]map[string]string, len(templates))

	for _, tmpl := range templates {
		tmpls[tmpl.Name] = map[string]string{
			"Front": tmpl.Front,
			"Back":  tmpl.Back,
		}
	}
	data := map[string]any{
		"model": map[string]any{
			"name":      name,
			"templates": tmpls,
		},
	}

	return c.do(ctx, request{
		Action:  "updateModelTemplates",
		Version: 6,
		Params:  data,
	}, nil)
}

func (c *Client) UpdateModelStyling(ctx context.Context, name string, css string) error {
	data := map[string]any{
		"model": map[string]any{
			"name": name,
			"css":  css,
		},
	}
	return c.do(ctx, request{
		Action:  "updateModelStyling",
		Version: 6,
		Params:  data,
	}, nil)
}

func (c *Client) GetModelTemplates(ctx context.Context, name string) ([]CardTemplate, error) {
	var result map[string]struct {
		Front string `json:"Front"`
		Back  string `json:"Back"`
	}

	err := c.do(ctx, request{
		Action:  "modelTemplates",
		Version: 6,
		Params: map[string]string{
			"modelName": name,
		},
	}, &result)
	if err != nil {
		return nil, err
	}

	templates := make([]CardTemplate, 0, len(result))
	for name, t := range result {
		templates = append(templates, CardTemplate{Name: name, Front: t.Front, Back: t.Back})
	}
	sort.Slice(templates, func(i, j int) bool { return templates[i].Name < templates[j].Name })

	return templates, nil
}

func (c *Client) GetModelStyling(ctx context.Context, name string) (string, error) {
	var result struct {
		CSS string `json:"css"`
	}

	err := c.do(ctx, request{
		Action:  "modelStyling",
		Version: 6,
		Params: map[string]string{
			"modelName": name,
		},
	}, &result)
	if err != nil {
		return "", err
	}

	return result.CSS, nil
}

func (c *Client) DeckExists(ctx context.Context, name string) (bool, error) {
	var result []string
	err := c.do(ctx, request{
		Action:  "deckNames",
		Version: 6,
	}, &result)
	if err != nil {
		return false, err
	}

	if slices.Contains(result, name) {
		return true, nil
	}
	return false, nil
}

func (c *Client) CreateDeck(ctx context.Context, name string) error {
	return c.do(ctx, request{
		Action:  "createDeck",
		Version: 6,
		Params: map[string]string{
			"deck": name,
		},
	}, nil)
}

func (c *Client) AddNote(ctx context.Context, deck, model string, n Note) error {
	note := map[string]any{
		"deckName":  deck,
		"modelName": model,
		"fields":    n.Fields,
		"tags":      n.Tags,
		"options": map[string]any{
			"allowDuplicate": false,
		},
	}
	return c.do(ctx, request{
		Action:  "addNote",
		Version: 6,
		Params:  map[string]any{"note": note},
	}, nil)
}

func (c *Client) NoteExists(ctx context.Context, deck, searchField string) (bool, int64, error) {
	ids := make([]int64, 1)
	exists := false

	query := fmt.Sprintf(`deck:%s "%s"`, deck, searchField)

	err := c.do(ctx, request{
		Action:  "findNotes",
		Version: 6,
		Params: map[string]any{
			"query": query,
		},
	}, &ids)
	if err != nil {
		return false, 0, err
	}

	if len(ids) == 1 {
		exists = true
	}

	if len(ids) == 0 {
		ids = append(ids, 0)
	}

	if len(ids) > 1 {
		return false, 0, errors.New("more than 1 ids received")
	}

	return exists, ids[0], nil
}

func (c *Client) UpdateNoteFields(ctx context.Context, noteID int64, fields map[string]string) error {
	return c.do(ctx, request{
		Action:  "updateNoteFields",
		Version: 6,
		Params: map[string]any{
			"note": map[string]any{
				"id":     noteID,
				"fields": fields,
			},
		},
	}, nil)
}

func (c *Client) UpdateNoteTags(ctx context.Context, noteID int64, tags []string) error {
	return c.do(ctx, request{
		Action:  "updateNoteTags",
		Version: 6,
		Params: map[string]any{
			"note": noteID,
			"tags": tags,
		},
	}, nil)
}

func (c *Client) do(ctx context.Context, req request, result any) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if r.Error != nil {
		return fmt.Errorf("anki error: %s", *r.Error)
	}
	if result != nil {
		return json.Unmarshal(r.Result, result)
	}
	return nil
}
