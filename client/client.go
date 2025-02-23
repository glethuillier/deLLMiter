package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// promptConstraints specifies a constant string instruction for response handling without modifications or additions
// NOTE: not using system command as it is not enforced by all models
// TODO: assess how this constraint could be improved
const promptConstraints = "Just ping back the following text (do not add anything else, do not modify the original text)"

type SupportedModels struct {
	Object string `json:"object"`
	Data   []struct {
		ID                string `json:"id"`
		Object            string `json:"object"`
		Type              string `json:"type"`
		Publisher         string `json:"publisher"`
		Arch              string `json:"arch"`
		CompatibilityType string `json:"compatibility_type"`
		Quantization      string `json:"quantization"`
		State             string `json:"state"`
		MaxContextLength  int    `json:"max_context_length"`
	} `json:"data"`
}

type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	Stream      bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int       `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   UsageInfo `json:"usage"`
	Stats   StatsInfo `json:"stats"`
	ModelInfo
	RuntimeInfo
}

type Choice struct {
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
}

type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type StatsInfo struct {
	TokensPerSecond  float64 `json:"tokens_per_second"`
	TimeToFirstToken float64 `json:"time_to_first_token"`
	GenerationTime   float64 `json:"generation_time"`
	StopReason       string  `json:"stop_reason"`
}

type ModelInfo struct {
	Arch          string `json:"arch"`
	Quant         string `json:"quant"`
	Format        string `json:"format"`
	ContextLength int    `json:"context_length"`
}

type RuntimeInfo struct {
	Name             string   `json:"name"`
	Version          string   `json:"version"`
	SupportedFormats []string `json:"supported_formats"`
}

type Client struct {
	queryURL   string
	httpClient *http.Client
}

func NewClient(baseURL string, requestedModel string) (*Client, error) {
	modelsListURL := fmt.Sprintf("%s/api/v0/models", baseURL)

	resp, err := http.Get(modelsListURL)
	if err != nil {
		return nil, fmt.Errorf("failed to check models API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models API returned non-200 status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var supportedModels SupportedModels
	if marshalErr := json.Unmarshal(body, &supportedModels); marshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", marshalErr)
	}

	for _, model := range supportedModels.Data {
		if strings.EqualFold(requestedModel, model.ID) {
			return &Client{
				queryURL:   fmt.Sprintf(baseURL + "/api/v0/chat/completions"),
				httpClient: &http.Client{},
			}, nil
		}
	}

	return nil, fmt.Errorf("model %s not supported/loaded by the API server", requestedModel)
}

// Query sends a request with specified model and message content, returning the response text or an error if encountered.
func (c *Client) Query(modelName string, messageContent string) (string, error) {
	request := Request{
		Model: modelName,
		Messages: []Message{{
			Role:    "user",
			Content: fmt.Sprintf("%s: %s", promptConstraints, messageContent),
		}},
		Temperature: 0.8,
		Stream:      false,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.queryURL, bytes.NewBuffer(requestJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 response received: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Choices) == 0 || response.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("no valid response received")
	}

	return response.Choices[0].Message.Content, nil
}
