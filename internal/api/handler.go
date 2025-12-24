package api

import (
	"log"
	"net/http"
	"qigent/internal/agent"
	"qigent/internal/chat"
	"qigent/internal/llm"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all CORS for MVP
	},
}

// HandleChat upgrades the HTTP connection to WebSocket and manages the chat session.
func HandleChat(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade:", err)
		return
	}
	defer ws.Close()

	// 1. Handshake: Read Config
	var config struct {
		APIKey  string `json:"apiKey"`
		BaseURL string `json:"baseUrl"`
		Model   string `json:"model"`
		Topic   string `json:"topic"`
		AgentA  struct {
			Name   string `json:"name"`
			Prompt string `json:"prompt"`
		} `json:"agentA"`
		AgentB struct {
			Name   string `json:"name"`
			Prompt string `json:"prompt"`
		} `json:"agentB"`
	}

	log.Println("Waiting for config handshake...")
	if err := ws.ReadJSON(&config); err != nil {
		log.Println("Failed to read config:", err)
		return
	}

	log.Printf("Received config: Model=%s, Topic=%s", config.Model, config.Topic)

	// Create Shared LLM Client
	// If BaseURL is empty, it defaults to OpenAI inside NewClient
	llmCfg := llm.Config{
		BaseURL: config.BaseURL,
		APIKey:  config.APIKey,
		Model:   config.Model,
	}
	client := llm.NewClient(llmCfg)

	// Create Agents with Config
	agentA := agent.NewAgent(config.AgentA.Name, config.AgentA.Prompt, client)
	agentB := agent.NewAgent(config.AgentB.Name, config.AgentB.Prompt, client)

	// Initialize Room
	room := chat.NewRoom([]*agent.Agent{agentA, agentB})

	// Start the orchestration loop with Topic
	room.StartLoop(config.Topic)
	defer room.StopLoop()

	// Send an initial system message
	ws.WriteJSON(chat.Message{Sender: "System", Content: "Debate initialized with model: " + config.Model, Type: "system"})

	// Listen for Broadcasts and send to WebSocket
	for {
		// Streaming from Room to Client
		msg, ok := <-room.Broadcast
		if !ok {
			break
		}

		err := ws.WriteJSON(msg)
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}
