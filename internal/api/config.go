package api

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

type AgentConfig struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
}

type AppConfig struct {
	APIKey  string      `json:"apiKey"`
	BaseURL string      `json:"baseUrl"`
	Model   string      `json:"model"`
	Topic   string      `json:"topic"`
	AgentA  AgentConfig `json:"agentA"`
	AgentB  AgentConfig `json:"agentB"`
}

var (
	configLock sync.RWMutex
	configPath = "config.json"
)

// Default config
var defaultConfig = AppConfig{
	BaseURL: "https://api.openai.com/v1",
	Model:   "gpt-3.5-turbo",
	AgentA: AgentConfig{
		Name:   "苏格拉底",
		Prompt: "你是一个苏格拉底式的哲学家，喜欢用反问引导思考。",
	},
	AgentB: AgentConfig{
		Name:   "现代大学生",
		Prompt: "你是一个务实的现代大学生，喜欢寻找直接的答案。",
	},
}

// GetConfig reads the config from file or returns default.
func GetConfig(c *gin.Context) {
	configLock.RLock()
	defer configLock.RUnlock()

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusOK, defaultConfig)
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read config"})
		return
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		// Fallback to default if file is corrupted
		c.JSON(http.StatusOK, defaultConfig)
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateConfig writes the config to file.
func UpdateConfig(c *gin.Context) {
	var newConfig AppConfig
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config format"})
		return
	}

	configLock.Lock()
	defer configLock.Unlock()

	data, err := json.MarshalIndent(newConfig, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize config"})
		return
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write config file"})
		return
	}

	c.JSON(http.StatusOK, newConfig)
}
