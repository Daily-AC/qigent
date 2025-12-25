package data

import (
	"qigent/internal/chat"
	"time"
)

type AgentConfig struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
}

type Conversation struct {
	ID        string         `json:"id"`
	Topic     string         `json:"topic"`
	Status    string         `json:"status"` // "active", "paused"
	AgentA    AgentConfig    `json:"agentA"`
	AgentB    AgentConfig    `json:"agentB"`
	History   []chat.Message `json:"history"`
	CreatedAt time.Time      `json:"createdAt"`
}

type Role struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
	Avatar string `json:"avatar"` // Optional URL or icon name
}
