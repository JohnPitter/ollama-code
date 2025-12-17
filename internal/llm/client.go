package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client cliente para Ollama API
type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewClient cria novo cliente Ollama
func NewClient(baseURL, model string) *Client {
	return &Client{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

// Complete faz uma chamada completa (não streaming)
func (c *Client) Complete(ctx context.Context, messages []Message, opts *CompletionOptions) (string, error) {
	reqOpts := Options{}
	if opts != nil {
		reqOpts.Temperature = opts.Temperature
		reqOpts.NumPredict = opts.MaxTokens
	}

	// Adicionar system prompt se fornecido
	if opts != nil && opts.SystemPrompt != "" {
		messages = append([]Message{{
			Role:    "system",
			Content: opts.SystemPrompt,
		}}, messages...)
	}

	req := Request{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
		Options:  reqOpts,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return response.Message.Content, nil
}

// CompleteStreaming faz chamada com streaming
func (c *Client) CompleteStreaming(ctx context.Context, messages []Message, opts *CompletionOptions, onChunk func(string)) (string, error) {
	reqOpts := Options{}
	if opts != nil {
		reqOpts.Temperature = opts.Temperature
		reqOpts.NumPredict = opts.MaxTokens
	}

	// Adicionar system prompt se fornecido
	if opts != nil && opts.SystemPrompt != "" {
		messages = append([]Message{{
			Role:    "system",
			Content: opts.SystemPrompt,
		}}, messages...)
	}

	req := Request{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
		Options:  reqOpts,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)

	for {
		var response Response
		if err := decoder.Decode(&response); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("decode response: %w", err)
		}

		// Chamar callback com chunk
		if onChunk != nil && response.Message.Content != "" {
			onChunk(response.Message.Content)
		}

		fullResponse.WriteString(response.Message.Content)

		if response.Done {
			break
		}
	}

	return fullResponse.String(), nil
}

// GetModel retorna o modelo configurado
func (c *Client) GetModel() string {
	return c.model
}

// SetModel altera o modelo
func (c *Client) SetModel(model string) {
	c.model = model
}
