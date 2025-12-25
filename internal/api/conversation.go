package api

import (
	"net/http"
	"qigent/internal/chat"
	"qigent/internal/data"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetConversations returns list of conversations (metadata only ideally, but returning all for MVP is fine)
func GetConversations(c *gin.Context) {
	convs, err := data.LoadConversations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, convs)
}

// GetConversation returns details of a single conversation
func GetConversation(c *gin.Context) {
	id := c.Param("id")
	conv, err := data.GetConversation(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if conv == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}
	c.JSON(http.StatusOK, conv)
}

// CreateConversationRequest
type CreateConversationRequest struct {
	Topic  string           `json:"topic"`
	AgentA data.AgentConfig `json:"agentA"`
	AgentB data.AgentConfig `json:"agentB"`
}

// CreateConversation initializes a new conversation
func CreateConversation(c *gin.Context) {
	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	newID := uuid.New().String()
	conv := data.Conversation{
		ID:        newID,
		Topic:     req.Topic,
		Status:    "active",
		CreatedAt: time.Now(),
		AgentA:    req.AgentA,
		AgentB:    req.AgentB,
		History:   []chat.Message{}, // Empty history
		// Add implicit System message?
	}

	// Save
	if err := data.SaveConversation(conv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
		return
	}

	c.JSON(http.StatusOK, conv)
}

// DeleteConversation deletes a conversation
func DeleteConversation(c *gin.Context) {
	id := c.Param("id")
	if err := data.DeleteConversation(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// GetRoles returns available roles
func GetRoles(c *gin.Context) {
	roles, err := data.LoadRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, roles)
}

// CreateRole adds a new role
func CreateRole(c *gin.Context) {
	var role data.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role data"})
		return
	}
	if role.Name == "" || role.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and Prompt are required"})
		return
	}

	if err := data.AddRole(role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save role"})
		return
	}
	c.JSON(http.StatusOK, role)
}

// DeleteRole removes a role
func DeleteRole(c *gin.Context) {
	name := c.Param("name")
	if err := data.DeleteRole(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
