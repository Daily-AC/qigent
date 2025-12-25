package api

import (
	"log"
	"net/http"
	"qigent/internal/agent"
	"qigent/internal/chat"
	"qigent/internal/data"
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
	// Need conversation ID to resume/connect
	conversationID := c.Query("conversationId")

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade:", err)
		return
	}
	defer ws.Close()

	if conversationID == "" {
		// Error: ID required
		ws.WriteJSON(gin.H{"error": "conversationId query param is required"})
		return
	}

	// Load Conversation
	conv, err := data.GetConversation(conversationID)
	if err != nil || conv == nil {
		log.Println("Conversation not found:", conversationID)
		ws.WriteJSON(gin.H{"error": "Conversation not found"})
		return
	}

	// 1. Handshake: Read Config (Optional update? Or just Ack?)
	// For now, we assume config (API Key) comes from global config or handshake.
	// But Topic and Agents are FIXED in the conversation data.
	// Let's still read the handshake for API KEY injection, but ignore agents/topic if they exist in conv.

	var handshake struct {
		APIKey  string `json:"apiKey"`
		BaseURL string `json:"baseUrl"`
		Model   string `json:"model"`
	}

	log.Println("Waiting for config handshake (Auth)...")
	if err := ws.ReadJSON(&handshake); err != nil {
		log.Println("Failed to read config:", err)
		return
	}

	// Create Shared LLM Client
	llmCfg := llm.Config{
		BaseURL: handshake.BaseURL,
		APIKey:  handshake.APIKey,
		Model:   handshake.Model,
	}
	client := llm.NewClient(llmCfg)

	// Create Agents from Conversation Data
	agentA := agent.NewAgent(conv.AgentA.Name, conv.AgentA.Prompt, client)
	agentB := agent.NewAgent(conv.AgentB.Name, conv.AgentB.Prompt, client)

	// Initialize Room
	room := chat.NewRoom([]*agent.Agent{agentA, agentB})

	// Hydrate Room History
	room.History = conv.History

	// Start the orchestration loop (Resume)
	// If History is empty, start with Topic. If not, resume.
	if len(conv.History) == 0 {
		room.StartLoop(conv.Topic)
	} else {
		room.StartLoop("") // Empty topic means resume
	}

	defer func() {
		room.StopLoop()
		// Final Save on disconnect
		conv.History = room.History
		data.SaveConversation(*conv)
	}()

	// Send an initial status
	ws.WriteJSON(chat.Message{Sender: "System", Content: "Connected to conversation: " + conv.Topic, Type: "system"})

	// Listen for Broadcasts and send to WebSocket
	for {
		msg, ok := <-room.Broadcast
		if !ok {
			break
		}

		// Optimization: Save periodically?
		// For MVP, we save on every "full" message to ensure persistence if crash
		if msg.Type == "full" {
			conv.History = room.History // Room history is updated by now
			go data.SaveConversation(*conv)
		}

		err := ws.WriteJSON(msg)
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}
