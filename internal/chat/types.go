package chat

// Message represents a single turn or a chunk in the conversation.
type Message struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
	Type    string `json:"type"` // "start", "chunk", "end", "system"
}

// Conversation holds the history of messages.
type Conversation struct {
	History []Message
}
