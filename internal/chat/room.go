package chat

import (
	"log"
	"qigent/internal/agent"
	"qigent/internal/llm"
	"strings"
	"time"
)

// Room manages the agents and the conversation loop.
type Room struct {
	Agents    []*agent.Agent
	History   []Message
	Broadcast chan Message
	InputChan chan Message // Channel for external (user) injection
	Stop      chan struct{}
}

// NewRoom creates a new chat room with the given agents.
func NewRoom(agents []*agent.Agent) *Room {
	return &Room{
		Agents:    agents,
		Broadcast: make(chan Message),
		InputChan: make(chan Message), // Buffered?
		Stop:      make(chan struct{}),
	}
}

// InjectMessage allows external injection of a message into the flow
func (r *Room) InjectMessage(msg Message) {
	// Push to InputChan. Ideally this should be non-blocking or guarded?
	// But since StartLoop consumes it, we can send it there.
	select {
	case r.InputChan <- msg:
	case <-time.After(500 * time.Millisecond):
		log.Println("Warn: InjectMessage timed out (Room busy or stopped)")
	}
}

// StartLoop begins the conversation loop.
func (r *Room) StartLoop(initialTopic string) {
	log.Printf("Room StartLoop: Conversation started on topic: %s", initialTopic)

	// Seed history with topic if provided and if history is empty
	// (Check added to avoid dupes if re-starting logic changed, though strict empty check is usually fine)
	var initialHistory []string
	if initialTopic != "" {
		initialHistory = append(initialHistory, "Moderator: Please discuss the topic: "+initialTopic)
	}

	go func() {
		// TURN RECOVERY LOGIC
		startIndex := 0
		if len(r.History) > 0 {
			lastMsg := r.History[len(r.History)-1]
			// If the last message was from an Agent, the NEXT agent should speak.
			// If it was "User", maybe the SAME agent should respond? Or next?
			// Let's assume User intervention doesn't skip a turn, but here we are talking about Resume.

			// Find who spoke last
			for i, ag := range r.Agents {
				if ag.Name == lastMsg.Sender {
					// This agent spoke last, so start with the NEXT one
					startIndex = (i + 1) % len(r.Agents)
					log.Printf("Resuming conversation. Last speaker: %s. Next speaker: %s", ag.Name, r.Agents[startIndex].Name)
					break
				}
			}
		}

		for {
			select {
			case <-r.Stop:
				return
			default:
				// Use a loop that we can offset
				for i := 0; i < len(r.Agents); i++ {
					// Calculate actual index based on offset
					idx := (startIndex + i) % len(r.Agents)
					ag := r.Agents[idx]

					// Check for stop
					select {
					case <-r.Stop:
						return
					default:
					}

					// CHECK FOR USER INJECTION BEFORE AGENT SPEAKS
					select {
					case userMsg := <-r.InputChan:
						log.Printf("User injected message: %s", userMsg.Content)
						// 1. Broadcast to UI so it shows up immediately (as User)
						r.Broadcast <- userMsg

						// 2. Add to History so Agent will see it
						// Format: "User: content"
						r.History = append(r.History, Message{
							Sender:  userMsg.Sender, // "User"
							Content: "User (Intervention): " + userMsg.Content,
							Type:    "full", // Treat as a full message
						})

						// Also broadcast "full" type for UI persistence if needed?
						// Actually r.Broadcast <- userMsg handles the UI bubble.
						// But we might want to save it.
						r.Broadcast <- Message{Sender: "System", Content: "User intervened.", Type: "system"}

						// Add to next turn's prompt context? handled by r.History loop below

					default:
						// No input, proceed
					}

					log.Printf("Agent %s is thinking...", ag.Name)

					// Prepare history
					var histStrs []string
					histStrs = append(histStrs, initialHistory...)

					for _, h := range r.History {
						// Only include completed messages in context?
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
					var interrupted bool

					// Manual Loop for Select
				loop:
					for {
						select {
						case <-r.Stop:
							return
						// 1. Interruption Check (Inside the loop!)
						case userMsg := <-r.InputChan:
							interrupted = true
							log.Printf("User interrupted %s: %s", ag.Name, userMsg.Content)

							// Broadcast User Message immediately
							r.Broadcast <- userMsg

							// We stop consuming stream here.
							// Ideally we should cancel context, but we don't hold one here.
							// The stream goroutine in llm/client.go will verify write to 'out' channel.
							// If we stop reading 'out', it blocks.
							// So we should launch a drainer in background to avoid leak.
							go func(c <-chan string) {
								for range c {
								} // Drain
							}(stream)

							// Append pending content to history (Interrupted Agent)
							interruptedContent := fullContentBuilder.String() + " [Interrupted]"
							r.History = append(r.History, Message{
								Sender:  ag.Name,
								Content: ag.Name + ": " + interruptedContent,
								Type:    "full",
							})

							// Append User Message to history
							r.History = append(r.History, Message{
								Sender:  userMsg.Sender,
								Content: "User (Intervention): " + userMsg.Content,
								Type:    "full",
							})

							// Notify Frontend: End of Agent turn (even if abrupt)
							r.Broadcast <- Message{Sender: ag.Name, Type: "end"}

							// Break inner loop -> Next Agent's Turn
							break loop

							// 2. Stream Consumption
						case chunk, ok := <-stream:
							if !ok {
								break loop // Stream finished naturally
							}
							fullContentBuilder.WriteString(chunk)
							r.Broadcast <- Message{Sender: ag.Name, Content: chunk, Type: "chunk"}
						}
					}

					if !interrupted {
						fullContent := fullContentBuilder.String()
						log.Printf("Agent %s finished speaking. Length: %d", ag.Name, len(fullContent))

						// Notify Frontend: End of turn
						r.Broadcast <- Message{Sender: ag.Name, Type: "end"}

						// Save to History (formatted)
						r.History = append(r.History, Message{
							Sender:  ag.Name,
							Content: ag.Name + ": " + fullContent,
							Type:    "full",
						})

						// Small delay between turns
						time.Sleep(1 * time.Second)
					} else {
						// If interrupted, maybe smaller delay or immediate next turn?
						time.Sleep(500 * time.Millisecond)
					}
				}
			}
		}
	}()
}

// Judge concludes the debate by generating a summary/verdict
func (r *Room) Judge(client *llm.Client) {
	log.Println("Room Judge: Judging conversation...")

	// 1. Stop the debate loop if running
	// (Caller should likely call StopLoop, or we ensure it here, but StopLoop is async signal)
	// We assume the caller (Handler) stops the loop or the 'cmd' handling implies stopping agent turns.
	// If we are in the Reader goroutine, sending Stop signal works.
	r.StopLoop()

	// Give a moment for agents to silence
	time.Sleep(500 * time.Millisecond)

	// 2. Prepare History
	var histStrs []string
	for _, h := range r.History {
		histStrs = append(histStrs, h.Content)
	}

	judgePrompt := "你是一位公正、幽默的辩论裁判。请阅读以上辩论记录，对双方的表现进行点评，指出亮眼之处和逻辑漏洞，并最终判定胜负（或平局）。请用Markdown格式输出，字数控制在500字以内。"

	// 3. Call LLM (Streaming)
	r.Broadcast <- Message{Sender: "Judge", Type: "start"}

	stream, err := client.ChatStream(judgePrompt, histStrs)
	if err != nil {
		r.Broadcast <- Message{Sender: "Judge", Content: "裁判把自己关在厕所里了...", Type: "end"}
		return
	}

	var fullContentBuilder strings.Builder
	for chunk := range stream {
		fullContentBuilder.WriteString(chunk)
		r.Broadcast <- Message{Sender: "Judge", Content: chunk, Type: "chunk"}
	}

	fullContent := fullContentBuilder.String()
	r.Broadcast <- Message{Sender: "Judge", Type: "end"}

	// 4. Save Verdict
	r.History = append(r.History, Message{
		Sender:  "Judge",
		Content: "Judge: " + fullContent,
		Type:    "full",
	})
}

// StopLoop stops the conversation.
func (r *Room) StopLoop() {
	// Check if channel is already closed to avoid panic?
	// Or use sync.Once. For MVP, simple close.
	// If multiple calls, this panics. Let's fix.
	select {
	case <-r.Stop:
		// already closed
	default:
		close(r.Stop)
	}
}
