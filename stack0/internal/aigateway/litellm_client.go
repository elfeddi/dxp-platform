package aigateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LiteLLMClient est le client HTTP vers LiteLLM Gateway.
type LiteLLMClient struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

// ChatMessage représente un message dans la conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest est la requête envoyée à LiteLLM.
type ChatRequest struct {
	Model     string        `json:"model"`
	Messages  []ChatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens"`
	Temp      float64       `json:"temperature"`
}

// ChatResponse est la réponse reçue de LiteLLM.
type ChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewLiteLLMClient crée un client LiteLLM.
func NewLiteLLMClient(baseURL, apiKey, model string) *LiteLLMClient {
	return &LiteLLMClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		client:  &http.Client{Timeout: 240 * time.Second},
	}
}

// Complete envoie un prompt au LLM et retourne la réponse.
func (c *LiteLLMClient) Complete(ctx context.Context, systemPrompt, userPrompt string) (*ChatResponse, error) {
	req := ChatRequest{
		Model: c.model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens: 1000,
		Temp:      0.1, // Faible température pour des réponses déterministes (YAML)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("litellm: marshal error: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		c.baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("litellm: request error: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("litellm: connection error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("litellm: read error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("litellm: error %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("litellm: decode error: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("litellm: empty response")
	}

	return &chatResp, nil
}

// ExtractContent retourne le contenu texte de la réponse LLM.
func (c *LiteLLMClient) ExtractContent(resp *ChatResponse) string {
	if len(resp.Choices) == 0 {
		return ""
	}
	return resp.Choices[0].Message.Content
}
