package chat

import (
	"log"
	"qigent/internal/agent"
	"strings"
	"time"
)

// Room manages the agents and the conversation loop.
type Room struct {
	Agents    []*agent.Agent
	History   []Message
	Broadcast chan Message
	Stop      chan struct{}
}

// NewRoom creates a new chat room with the given agents.
func NewRoom(agents []*agent.Agent) *Room {
	return &Room{
		Agents:    agents,
		Broadcast: make(chan Message),
		Stop:      make(chan struct{}),
	}
}

// StartLoop begins the conversation loop.
func (r *Room) StartLoop(initialTopic string) {
	log.Printf("Room StartLoop: Conversation started on topic: %s", initialTopic)

	// Seed history with topic if provided
	var initialHistory []string
	if initialTopic != "" {
		initialHistory = append(initialHistory, "Moderator: Please discuss the topic: "+initialTopic)
		// r.History = append(r.History, Message{Sender: "System", Content: "Topic: " + initialTopic, Type: "system"})
	}

	go func() {
		for {
			select {
			case <-r.Stop:
				return
			default:
				for _, ag := range r.Agents {
					// Check for stop
					select {
					case <-r.Stop:
						return
					default:
					}

					log.Printf("Agent %s is thinking...", ag.Name)

					// Prepare history
					// Convert struct history to string slice for LLM
					// We only include "full" messages from history
					var histStrs []string
					histStrs = append(histStrs, initialHistory...)

					for _, h := range r.History {
						// Only include completed messages in context?
						// Or just formatted string.
						histStrs = append(histStrs, h.Content)
					}

					// Notify Frontend: Start of turn
					r.Broadcast <- Message{Sender: ag.Name, Type: "start"}

					// Stream
					stream, err := ag.SpeakStream(histStrs)
					if err != nil {
						r.Broadcast <- Message{Sender: ag.Name, Content: "[Error: " + err.Error() + "]", Type: "end"}
						time.Sleep(2 * time.Second)
						continue
					}

					var fullContentBuilder strings.Builder

					for chunk := range stream {
						select {
						case <-r.Stop:
							return
						default:
						}

						fullContentBuilder.WriteString(chunk)
						r.Broadcast <- Message{Sender: ag.Name, Content: chunk, Type: "chunk"}
					}

					fullContent := fullContentBuilder.String()
					log.Printf("Agent %s finished speaking. Length: %d", ag.Name, len(fullContent))

					// Notify Frontend: End of turn
					r.Broadcast <- Message{Sender: ag.Name, Type: "end"}

					// Save to History (formatted)
					// We save standard format "Name: Content" for LLM context
					r.History = append(r.History, Message{
						Sender:  ag.Name,
						Content: ag.Name + ": " + fullContent,
						Type:    "full",
					})

					// Small delay between turns
					time.Sleep(1 * time.Second)
				}
			}
		}
	}()
}

// StopLoop stops the conversation.
func (r *Room) StopLoop() {
	close(r.Stop)
}
