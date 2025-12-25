import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useChatStore = defineStore('chat', () => {
  const messages = ref([])
  const isConnected = ref(false)
  const socket = ref(null)

  function connect(url, config) {
    if (socket.value) return

    // Connect to Backend WS
    socket.value = new WebSocket(url)

    socket.value.onopen = () => {
      isConnected.value = true
      console.log('Connected to Chat WS, sending config...')
      // Send Config immediately
      socket.value.send(JSON.stringify(config))
    }

    socket.value.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        
        if (msg.type === 'start') {
          // New message bubble start
          messages.value.push({
            sender: msg.sender,
            content: '', // Will be filled by chunks
            type: msg.type
          })
        } else if (msg.type === 'chunk') {
          // Append to last message
          // Check if last message matches sender?
          const lastMsg = messages.value[messages.value.length - 1]
          if (lastMsg && lastMsg.sender === msg.sender) {
            lastMsg.content += msg.content
          }
        } else if (msg.type === 'end') {
          // Finished turn, maybe mark as done?
        } else if (msg.type === 'system') {
           messages.value.push(msg)
        } else {
           // Fallback for "full" or unidentified
           messages.value.push(msg)
        }

      } catch (e) {
        console.error('Failed to parse message:', event.data)
      }
    }

    socket.value.onclose = () => {
      isConnected.value = false
      socket.value = null
      console.log('Disconnected from Chat WS')
    }
  }

  function disconnect() {
    if (socket.value) {
      socket.value.close()
      socket.value = null
      isConnected.value = false
    }
  }

  return { messages, isConnected, connect, disconnect }
})
