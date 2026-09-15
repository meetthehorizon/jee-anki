package anki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultEndpoint = "http://127.0.0.1:8765"
	DefaultDeckName = "Inbox"
	DefaultTimeout  = 10 * time.Second
)

// Client interacts with Anki Desktop via the AnkiConnect add-on.
type Client struct {
	endpoint   string
	httpClient *http.Client
}

// Request is the standard AnkiConnect JSON-RPC request structure.
type Request struct {
	Action  string      `json:"action"`
	Version int         `json:"version"`
	Params  interface{} `json:"params,omitempty"`
}

// Response is the standard AnkiConnect JSON-RPC response structure.
type Response struct {
	Result json.RawMessage `json:"result"`
	Error  *string         `json:"error"`
}

// Note represents an atomic flashcard to be inserted into Anki.
type Note struct {
	Front string `json:"front"`
	Back  string `json:"back"`
}

// AddNotesResult summarizes the batch insertion results.
type AddNotesResult struct {
	Total     int
	Added     int
	Duplicate int
}

// NewClient initializes a new AnkiConnect client.
func NewClient(endpoint ...string) *Client {
	target := DefaultEndpoint
	if len(endpoint) > 0 && endpoint[0] != "" {
		target = endpoint[0]
	}
	return &Client{
		endpoint: target,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// invoke sends a JSON-RPC request to AnkiConnect.
func (c *Client) invoke(ctx context.Context, action string, params interface{}, result interface{}) error {
	reqBody := Request{
		Action:  action,
		Version: 6,
		Params:  params,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to encode anki request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not connect to Anki: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response from Anki: %w", err)
	}

	var ankiResp Response
	if err := json.Unmarshal(body, &ankiResp); err != nil {
		return fmt.Errorf("invalid json response from Anki: %w", err)
	}

	if ankiResp.Error != nil && *ankiResp.Error != "" {
		return fmt.Errorf("anki error: %s", *ankiResp.Error)
	}

	if result != nil && len(ankiResp.Result) > 0 {
		if err := json.Unmarshal(ankiResp.Result, result); err != nil {
			return fmt.Errorf("failed to parse result: %w", err)
		}
	}

	return nil
}

// Ping checks whether Anki Desktop is running and AnkiConnect is responsive.
func (c *Client) Ping(ctx context.Context) error {
	var version int
	return c.invoke(ctx, "version", nil, &version)
}

// EnsureDeck checks if the target deck exists, and creates it if missing.
func (c *Client) EnsureDeck(ctx context.Context, deckName string) error {
	var decks []string
	if err := c.invoke(ctx, "deckNames", nil, &decks); err != nil {
		return err
	}

	for _, d := range decks {
		if d == deckName {
			return nil // Deck already exists
		}
	}

	params := map[string]string{"deck": deckName}
	var deckID int64
	return c.invoke(ctx, "createDeck", params, &deckID)
}

// GetBasicModelName returns the name of the standard 2-field model ("Basic").
func (c *Client) GetBasicModelName(ctx context.Context) (string, error) {
	var models []string
	if err := c.invoke(ctx, "modelNames", nil, &models); err != nil {
		return "Basic", err
	}

	for _, m := range models {
		if m == "Basic" {
			return "Basic", nil
		}
	}

	if len(models) > 0 {
		return models[0], nil
	}

	return "Basic", nil
}

// AddNotes adds a batch of flashcards to the specified deck with tags.
func (c *Client) AddNotes(ctx context.Context, deckName, modelName string, notes []Note, tags []string) (*AddNotesResult, error) {
	if len(notes) == 0 {
		return &AddNotesResult{}, nil
	}

	type NoteField struct {
		Front string `json:"Front"`
		Back  string `json:"Back"`
	}

	type NoteOption struct {
		AllowDuplicate bool   `json:"allowDuplicate"`
		DuplicateScope string `json:"duplicateScope"`
	}

	type AnkiNote struct {
		DeckName  string     `json:"deckName"`
		ModelName string     `json:"modelName"`
		Fields    NoteField  `json:"fields"`
		Tags      []string   `json:"tags"`
		Options   NoteOption `json:"options"`
	}

	ankiNotes := make([]AnkiNote, len(notes))
	for i, n := range notes {
		ankiNotes[i] = AnkiNote{
			DeckName:  deckName,
			ModelName: modelName,
			Fields: NoteField{
				Front: n.Front,
				Back:  n.Back,
			},
			Tags: tags,
			Options: NoteOption{
				AllowDuplicate: false,
				DuplicateScope: "deck",
			},
		}
	}

	params := map[string]interface{}{
		"notes": ankiNotes,
	}

	// Result is a slice of *int64 (null if duplicate or error)
	var ids []*int64
	if err := c.invoke(ctx, "addNotes", params, &ids); err != nil {
		return nil, err
	}

	res := &AddNotesResult{
		Total: len(notes),
	}

	for _, id := range ids {
		if id != nil && *id > 0 {
			res.Added++
		} else {
			res.Duplicate++
		}
	}

	return res, nil
}
