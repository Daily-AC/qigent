package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Config holds the configuration for the LLM API.
type Config struct {
	BaseURL string
	APIKey  string
	Model   string
}

// Client handles communication with the LLM provider.
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient creates a new LLM client.
func NewClient(config Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	return &Client{
		config: config,
		httpClient: &http.Client{
			// Longer timeout for streaming sessions if needed,
			// though usually http.Client.Do timeout covers the initial connection.
			// For streaming, we might rely on loop cancellation.
			Timeout: 0, // No global timeout for streaming
		},
	}
}

// ChatMessage represents a message in the LLM conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents the payload sent to the API.
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// ChatCompletionChunk represents the streaming response chunk.
type ChatCompletionChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// ChatStream sends a streaming chat completion request.
// It returns a channel that emits chunks of text, and an error if the request setup fails.
func (c *Client) ChatStream(systemPrompt string, history []string) (<-chan string, error) {
	// Construct messages
	var messages []ChatMessage
	messages = append(messages, ChatMessage{Role: "system", Content: systemPrompt})

	valuableHistory := history
	// Simple rolling window to avoid exceeding context too quickly (MVP hack)
	if len(history) > 10 {
		valuableHistory = history[len(history)-10:]
	}

	for _, h := range valuableHistory {
		messages = append(messages, ChatMessage{Role: "user", Content: h})
	}

	reqBody := ChatRequest{
		Model:    c.config.Model,
		Messages: messages,
		Stream:   true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := c.config.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// Use a separate client or override timeout for streaming if needed
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	out := make(chan string)

	go func() {
		defer resp.Body.Close()
		defer close(out)

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					// Log error? For MVP just stop.
				}
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var chunk ChatCompletionChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				if content != "" {
					out <- content
				}
			}
		}
	}()

	return out, nil
}

// Chat (Non-streaming) - kept for compatibility if needed, but we focus on Stream
func (c *Client) Chat(systemPrompt string, history []string) (string, error) {
	// ... (Implementation omitted for brevity, focusing on Stream)
	return "", fmt.Errorf("use ChatStream instead")
}
