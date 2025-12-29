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

// Conversation Routes

func CreateConversation(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var req struct {
		Topic  string           `json:"topic"`
		AgentA data.AgentConfig `json:"agentA"`
		AgentB data.AgentConfig `json:"agentB"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	conv := &data.Conversation{
		ID:        chat.NewID(),
		UserID:    userID,
		Topic:     req.Topic,
		Status:    "active",
		AgentA:    req.AgentA,
		AgentB:    req.AgentB,
		History:   []chat.Message{},
		CreatedAt: chat.Now(),
	}

	if err := data.CreateConversation(conv); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, conv)
}

func GetConversations(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	convs, err := data.GetConversations(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, convs)
}

func GetConversation(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(uint)

	conv, err := data.GetConversation(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}
	// Auth Check
	if conv.UserID != userID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}
	c.JSON(200, conv)
}

func DeleteConversation(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(uint)
	if err := data.DeleteConversation(id, userID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "deleted"})
}

// Role Routes

func GetRoles(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	roles, err := data.GetRoles(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, roles)
}

func CreateRole(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var role data.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(400, gin.H{"error": "Invalid json"})
		return
	}
	role.UserID = userID
	if err := data.AddRole(&role); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, role)
}

func DeleteRole(c *gin.Context) {
	name := c.Param("name")
	userID := c.MustGet("userID").(uint)
	if err := data.DeleteRole(name, userID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "deleted"})
}

// WebSocket Chat Handler
func HandleChat(c *gin.Context) {
	// Auth middleware should have set userID, BUT for WS, headers might be tricky.
	// We expect token in Query Param for WS connection, which AuthMiddleware handles.
	userID := c.MustGet("userID").(uint)

	conversationID := c.Query("conversationId")

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade:", err)
		return
	}
	defer ws.Close()

	if conversationID == "" {
		ws.WriteJSON(gin.H{"error": "conversationId conversationId required"})
		return
	}

	conv, err := data.GetConversation(conversationID)
	if err != nil || conv == nil {
		ws.WriteJSON(gin.H{"error": "Conversation not found"})
		return
	}

	if conv.UserID != userID {
		ws.WriteJSON(gin.H{"error": "Forbidden"})
		return
	}

	// Handshake
	var handshake struct {
		APIKey  string `json:"apiKey"`
		BaseURL string `json:"baseUrl"`
		Model   string `json:"model"`
	}
	if err := ws.ReadJSON(&handshake); err != nil {
		return
	}

	// Create Client
	llmCfg := llm.Config{
		BaseURL: handshake.BaseURL,
		APIKey:  handshake.APIKey,
		Model:   handshake.Model,
	}
	client := llm.NewClient(llmCfg)

	agentA := agent.NewAgent(conv.AgentA.Name, conv.AgentA.Prompt, client)
	agentB := agent.NewAgent(conv.AgentB.Name, conv.AgentB.Prompt, client)

	room := chat.NewRoom([]*agent.Agent{agentA, agentB})
	room.History = conv.History

	if len(conv.History) == 0 {
		room.StartLoop(conv.Topic)
	} else {
		room.StartLoop("")
	}

	defer func() {
		room.StopLoop()
		conv.History = room.History
		data.SaveConversation(conv)
	}()

	ws.WriteJSON(chat.Message{Sender: "System", Content: "Connected: " + conv.Topic, Type: "system"})

	// Reader Loop
	go func() {
		defer room.StopLoop()
		for {
			var msg chat.Message
			if err := ws.ReadJSON(&msg); err != nil {
				return
			}
			if msg.Type == "cmd" && msg.Content == "conclude" {
				log.Println("Received conclude command, starting Judge...")
				if len(room.Agents) > 0 {
					go room.Judge(room.Agents[0].LLMClient)
				}
			} else if msg.Sender == "User" {
				room.InjectMessage(msg)
			}
		}
	}()

	// Writer Loop
	for {
		msg, ok := <-room.Broadcast
		if !ok {
			break
		}
		if msg.Type == "full" {
			conv.History = room.History
			go data.SaveConversation(conv)
		}
		if err := ws.WriteJSON(msg); err != nil {
			break
		}
	}
}

// Config Handlers
func GetConfig(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	cfg, err := data.GetChatConfig(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, cfg)
}

func UpdateConfig(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var cfg data.ChatConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	if err := data.SaveChatConfig(userID, &cfg); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, cfg)
}
