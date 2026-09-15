package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/meetthehorizon/jee-anki/internal/anki"
)

const (
	DefaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"
	DefaultTimeout = 120 * time.Second
)

// ExtractionResult represents the structured output returned by Gemini.
type ExtractionResult struct {
	Subject       string      `json:"subject"`
	Topic         string      `json:"topic"`
	SuggestedTags []string    `json:"suggested_tags"`
	Cards         []anki.Note `json:"cards"`
}

// Client interacts with the Google Gemini REST API.
type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewClient creates a new Gemini API client.
func NewClient(apiKey, model string, optionalBaseURL ...string) *Client {
	baseURL := DefaultBaseURL
	if len(optionalBaseURL) > 0 && optionalBaseURL[0] != "" {
		baseURL = optionalBaseURL[0]
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// SetModel updates the model identifier.
func (c *Client) SetModel(model string) {
	c.model = model
}

// SetAPIKey updates the API key.
func (c *Client) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

type apiRequest struct {
	Contents          []apiContent          `json:"contents"`
	SystemInstruction *apiContent           `json:"system_instruction,omitempty"`
	GenerationConfig  apiGenerationConfig   `json:"generation_config"`
}

type apiContent struct {
	Role  string    `json:"role,omitempty"`
	Parts []apiPart `json:"parts"`
}

type apiPart struct {
	Text       string         `json:"text,omitempty"`
	InlineData *apiInlineData `json:"inline_data,omitempty"`
}

type apiInlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type apiGenerationConfig struct {
	ResponseMimeType string      `json:"response_mime_type"`
	ResponseSchema   interface{} `json:"response_schema"`
}

type apiErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

type apiSuccessResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
}

// buildResponseSchema returns the Gemini JSON schema for extraction.
func buildResponseSchema() interface{} {
	return map[string]interface{}{
		"type": "OBJECT",
		"properties": map[string]interface{}{
			"subject": map[string]interface{}{
				"type": "STRING",
			},
			"topic": map[string]interface{}{
				"type": "STRING",
			},
			"suggested_tags": map[string]interface{}{
				"type": "ARRAY",
				"items": map[string]interface{}{
					"type": "STRING",
				},
			},
			"cards": map[string]interface{}{
				"type": "ARRAY",
				"items": map[string]interface{}{
					"type": "OBJECT",
					"properties": map[string]interface{}{
						"front": map[string]interface{}{
							"type": "STRING",
						},
						"back": map[string]interface{}{
							"type": "STRING",
						},
					},
					"required": []string{"front", "back"},
				},
			},
		},
		"required": []string{"suggested_tags", "cards"},
	}
}

// ExtractFromPDFChunk sends a base64-encoded PDF chunk to Gemini and parses the structured response.
func (c *Client) ExtractFromPDFChunk(ctx context.Context, pdfBytes []byte) (*ExtractionResult, error) {
	if len(pdfBytes) == 0 {
		return nil, fmt.Errorf("pdf data is empty")
	}

	b64PDF := base64.StdEncoding.EncodeToString(pdfBytes)

	reqPayload := apiRequest{
		SystemInstruction: &apiContent{
			Parts: []apiPart{
				{Text: SystemInstruction},
			},
		},
		Contents: []apiContent{
			{
				Role: "user",
				Parts: []apiPart{
					{
						InlineData: &apiInlineData{
							MimeType: "application/pdf",
							Data:     b64PDF,
						},
					},
					{
						Text: UserPrompt,
					},
				},
			},
		},
		GenerationConfig: apiGenerationConfig{
			ResponseMimeType: "application/json",
			ResponseSchema:   buildResponseSchema(),
		},
	}

	payloadBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	endpoint := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error contacting Google Gemini: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr apiErrorResponse
		if err := json.Unmarshal(bodyBytes, &apiErr); err == nil && apiErr.Error.Message != "" {
			switch {
			case apiErr.Error.Code == 400 && strings.Contains(apiErr.Error.Message, "API_KEY_INVALID"):
				return nil, fmt.Errorf("invalid Google API key. Please verify your key at https://aistudio.google.com/")
			case apiErr.Error.Code == 404:
				return nil, fmt.Errorf("model '%s' not found or deprecated: %s. Try 'gemini-flash-latest'", c.model, apiErr.Error.Message)
			case apiErr.Error.Code == 429:
				return nil, fmt.Errorf("google Gemini rate limit exceeded. Please wait a few seconds and retry")
			default:
				return nil, fmt.Errorf("gemini API error (%d): %s", apiErr.Error.Code, apiErr.Error.Message)
			}
		}
		return nil, fmt.Errorf("gemini API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var successResp apiSuccessResponse
	if err := json.Unmarshal(bodyBytes, &successResp); err != nil {
		return nil, fmt.Errorf("failed to decode Gemini API response: %w", err)
	}

	if len(successResp.Candidates) == 0 || len(successResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response received from Gemini model")
	}

	rawJSON := successResp.Candidates[0].Content.Parts[0].Text
	rawJSON = strings.TrimSpace(rawJSON)
	// Strip any potential markdown code blocks if present
	if strings.HasPrefix(rawJSON, "```json") {
		rawJSON = strings.TrimPrefix(rawJSON, "```json")
		rawJSON = strings.TrimSuffix(rawJSON, "```")
		rawJSON = strings.TrimSpace(rawJSON)
	} else if strings.HasPrefix(rawJSON, "```") {
		rawJSON = strings.TrimPrefix(rawJSON, "```")
		rawJSON = strings.TrimSuffix(rawJSON, "```")
		rawJSON = strings.TrimSpace(rawJSON)
	}

	var result ExtractionResult
	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		return nil, fmt.Errorf("failed to parse structured cards JSON: %w (raw response: %s)", err, rawJSON)
	}

	return &result, nil
}
