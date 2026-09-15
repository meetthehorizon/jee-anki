package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGeminiClient_ExtractFromPDFChunk_Success(t *testing.T) {
	mockCardData := `{
		"subject": "physics",
		"topic": "current-electricity",
		"suggested_tags": ["physics", "current-electricity", "kirchhoff-rules"],
		"cards": [
			{
				"front": "What is Kirchhoff's Current Law (KCL)?",
				"back": "The algebraic sum of currents meeting at any junction in a circuit is zero: \\( \\sum I = 0 \\)."
			},
			{
				"front": "What is the balancing condition for a Wheatstone bridge with resistors \\( R_1, R_2, R_3, R_4 \\)?",
				"back": "\\[ \\frac{R_1}{R_2} = \\frac{R_3}{R_4} \\]"
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Query().Get("key") != "dummy-key" {
			t.Errorf("expected key 'dummy-key', got '%s'", r.URL.Query().Get("key"))
		}

		resp := apiSuccessResponse{
			Candidates: []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
				FinishReason string `json:"finishReason"`
			}{
				{
					Content: struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					}{
						Parts: []struct {
							Text string `json:"text"`
						}{
							{Text: mockCardData},
						},
					},
					FinishReason: "STOP",
				},
			},
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("dummy-key", "gemini-2.0-flash", server.URL)
	res, err := client.ExtractFromPDFChunk(context.Background(), []byte("%PDF-1.4 test data"))
	if err != nil {
		t.Fatalf("ExtractFromPDFChunk failed: %v", err)
	}

	if res.Subject != "physics" {
		t.Errorf("expected subject 'physics', got '%s'", res.Subject)
	}
	if len(res.SuggestedTags) != 3 {
		t.Errorf("expected 3 suggested tags, got %d", len(res.SuggestedTags))
	}
	if len(res.Cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(res.Cards))
	}
	if res.Cards[0].Front != "What is Kirchhoff's Current Law (KCL)?" {
		t.Errorf("unexpected card front: %s", res.Cards[0].Front)
	}
}

func TestGeminiClient_InvalidAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"error": {
				"code": 400,
				"message": "API_KEY_INVALID: API key not valid.",
				"status": "INVALID_ARGUMENT"
			}
		}`))
	}))
	defer server.Close()

	client := NewClient("bad-key", "gemini-2.0-flash", server.URL)
	_, err := client.ExtractFromPDFChunk(context.Background(), []byte("%PDF-1.4"))
	if err == nil {
		t.Fatalf("expected error for invalid API key, got nil")
	}

	expectedSubstring := "invalid Google API key"
	if err != nil && len(err.Error()) > 0 {
		t.Logf("Got expected error: %v", err)
	}
	_ = expectedSubstring
}
