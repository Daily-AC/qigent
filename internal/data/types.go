package data

import (
	"qigent/internal/chat"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"unique"`
	Password string `json:"-"` // Hash
}

type AgentConfig struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
}

// Stored as JSON in MySQL for simplicity in MVP, or normalized tables?
// For MVP, let's keep History as a JSON blob if we use MySQL 5.7+ JSON type,
// or just text. GORM supports serializer.

type Conversation struct {
	ID     string `json:"id" gorm:"primaryKey"`
	UserID uint   `json:"userId"`
	Topic  string `json:"topic"`
	Status string `json:"status"` // "active", "paused"

	// Embedded fields or JSON columns?
	// For simplicity in GORM w/ MySQL, we can use `serializer:json`
	AgentA  AgentConfig    `json:"agentA" gorm:"serializer:json"`
	AgentB  AgentConfig    `json:"agentB" gorm:"serializer:json"`
	History []chat.Message `json:"history" gorm:"serializer:json"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Role struct {
	gorm.Model
	UserID uint   `json:"userId"` // 0 for public/system roles?
	Name   string `json:"name" gorm:"uniqueIndex:idx_name_user"`
	Prompt string `json:"prompt"`
	Avatar string `json:"avatar"`
}
