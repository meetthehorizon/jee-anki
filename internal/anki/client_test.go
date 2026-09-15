package anki

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnkiClient_Ping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Action != "version" {
			t.Errorf("expected action 'version', got '%s'", req.Action)
		}

		resp := Response{
			Result: json.RawMessage(`6`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

func TestAnkiClient_EnsureDeck(t *testing.T) {
	createdDeck := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		json.NewDecoder(r.Body).Decode(&req)

		if req.Action == "deckNames" {
			resp := Response{
				Result: json.RawMessage(`["Default", "Physics"]`),
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		if req.Action == "createDeck" {
			createdDeck = true
			resp := Response{
				Result: json.RawMessage(`123456789`),
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.EnsureDeck(context.Background(), "Inbox"); err != nil {
		t.Fatalf("EnsureDeck failed: %v", err)
	}

	if !createdDeck {
		t.Errorf("expected deck 'Inbox' to be created")
	}
}

func TestAnkiClient_AddNotes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		json.NewDecoder(r.Body).Decode(&req)

		if req.Action != "addNotes" {
			t.Errorf("expected action 'addNotes', got '%s'", req.Action)
		}

		// Simulate 2 cards: first added (id 101), second duplicate (null)
		resp := Response{
			Result: json.RawMessage(`[101, null]`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	notes := []Note{
		{Front: "Formula for kinetic energy?", Back: "\\( \\frac{1}{2}mv^2 \\)"},
		{Front: "Formula for momentum?", Back: "\\( p = mv \\)"},
	}

	res, err := client.AddNotes(context.Background(), "Inbox", "Basic", notes, []string{"physics", "mechanics"})
	if err != nil {
		t.Fatalf("AddNotes failed: %v", err)
	}

	if res.Total != 2 {
		t.Errorf("expected 2 total, got %d", res.Total)
	}
	if res.Added != 1 {
		t.Errorf("expected 1 added, got %d", res.Added)
	}
	if res.Duplicate != 1 {
		t.Errorf("expected 1 duplicate, got %d", res.Duplicate)
	}
}

func TestAnkiClient_RefreshGUIAndGetDeckTotal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		json.NewDecoder(r.Body).Decode(&req)

		if req.Action == "guiDeckBrowser" {
			resp := Response{
				Result: json.RawMessage(`null`),
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		if req.Action == "getDeckStats" {
			resp := Response{
				Result: json.RawMessage(`{"123": {"deck_id": 123, "name": "Inbox", "total_in_deck": 42}}`),
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.RefreshGUI(context.Background()); err != nil {
		t.Fatalf("RefreshGUI failed: %v", err)
	}

	total, err := client.GetDeckTotal(context.Background(), "Inbox")
	if err != nil {
		t.Fatalf("GetDeckTotal failed: %v", err)
	}
	if total != 42 {
		t.Errorf("expected 42 total, got %d", total)
	}
}
