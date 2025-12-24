package agent

import (
	"log"
	"qigent/internal/llm"
)

// Agent represents an entity that can participate in a conversation.
type Agent struct {
	Name         string
	SystemPrompt string
	LLMClient    *llm.Client
}

// NewAgent creates a new Agent instance.
func NewAgent(name, systemPrompt string, client *llm.Client) *Agent {
	return &Agent{
		Name:         name,
		SystemPrompt: systemPrompt,
		LLMClient:    client,
	}
}

// SpeakStream calls the LLM using streaming and returns a channel of chunks.
func (a *Agent) SpeakStream(history []string) (<-chan string, error) {
	if a.LLMClient == nil {
		return nil, nil // Should handle error better
	}

	// Add "[Agent Name]: " prefix to history if not present?
	// The current history format in Room is "Sender: Content".

	stream, err := a.LLMClient.ChatStream(a.SystemPrompt, history)
	if err != nil {
		log.Printf("Agent %s LLM error: %v", a.Name, err)
		return nil, err
	}

	return stream, nil
}
