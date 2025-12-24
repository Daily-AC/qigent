package chat

import (
	"qigent/internal/agent"
	"testing"
	"time"
)

func TestRoomLoop(t *testing.T) {
	a1 := agent.NewAgent("A1", "Prompt 1", nil)
	a2 := agent.NewAgent("A2", "Prompt 2", nil)

	room := NewRoom([]*agent.Agent{a1, a2})

	// Start the loop in a goroutine
	room.StartLoop("")

	// Helper function to read a message with timeout
	readMessage := func(timeout time.Duration) *Message {
		select {
		case msg := <-room.Broadcast:
			return &msg
		case <-time.After(timeout):
			return nil
		}
	}

	// Expect message from Agent A (or whoever goes first, order depends on slice)
	// Agent A is index 0
	msg1 := readMessage(4 * time.Second) // Takes some time due to Sleep
	if msg1 == nil {
		t.Fatal("Timeout waiting for message 1")
	}
	t.Logf("Received message 1 from %s: %s", msg1.Sender, msg1.Content)

	// Expect message from Agent B
	msg2 := readMessage(4 * time.Second)
	if msg2 == nil {
		t.Fatal("Timeout waiting for message 2")
	}
	t.Logf("Received message 2 from %s: %s", msg2.Sender, msg2.Content)

	room.StopLoop()
}
