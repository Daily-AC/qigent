<script setup>
import { ref, onMounted } from 'vue'
import { useChatStore } from './stores/chat'
import ChatWindow from './components/ChatWindow.vue'
import AgentCard from './components/AgentCard.vue'
import SettingsModal from './components/SettingsModal.vue'
import axios from 'axios'

const chatStore = useChatStore()
const isSettingsOpen = ref(false)

// Default config
const config = ref({
  apiKey: '', 
  baseUrl: 'https://api.openai.com/v1',
  model: 'gpt-3.5-turbo',
  agentA: { name: '', prompt: '' },
  agentB: { name: '', prompt: '' }
})

// Configure Axios base URL if needed, but we are on localhost:5173 calling localhost:8080
const api = axios.create({
  baseURL: 'http://localhost:8080'
})

onMounted(async () => {
  try {
    const res = await api.get('/config')
    config.value = res.data
    
    // Open settings if no API key
    if (!config.value.apiKey) {
      isSettingsOpen.value = true
    }
  } catch (e) {
    console.error('Failed to load config from backend', e)
    // Fallback?
  }
})

const saveConfig = async (newConfig) => {
  try {
    const res = await api.post('/config', newConfig)
    config.value = res.data
    // alert('Configuration saved to server.')
  } catch (e) {
    console.error('Failed to save config', e)
    alert('Failed to save configuration.')
  }
}

const startChat = () => {
  if (!config.value.apiKey) {
    alert('Please configure API Key in settings first.')
    isSettingsOpen.value = true
    return
  }
  chatStore.connect(config.value)
}

const stopChat = () => {
  chatStore.disconnect()
}
</script>

<template>
  <div class="min-h-screen bg-gray-100 py-8 px-4 font-sans">
    <div class="max-w-4xl mx-auto space-y-6">
      <!-- Header -->
      <header class="flex justify-between items-center bg-white p-4 rounded-xl shadow-sm">
        <div>
          <h1 class="text-2xl font-bold text-gray-800 tracking-tight">Multi-Agent Debate</h1>
          <p class="text-xs text-gray-500">Powered by Go & Vue & LLM</p>
        </div>
        <button @click="isSettingsOpen = true" class="p-2 hover:bg-gray-100 rounded-lg transition text-gray-600">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543 .826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </button>
      </header>

      <!-- Agent Status Area -->
      <div class="grid grid-cols-2 gap-4">
        <AgentCard :name="config.agentA.name" :role="'Agent A'" :active="false" />
        <AgentCard :name="config.agentB.name" :role="'Agent B'" :active="false" />
      </div>

      <!-- Main Chat Area -->
      <div class="bg-white rounded-2xl p-2 shadow-xl ring-1 ring-black/5">
        <ChatWindow :messages="chatStore.messages" />
      </div>

      <!-- Controls -->
      <div class="flex justify-center gap-4">
        <button 
          v-if="!chatStore.isConnected"
          @click="startChat" 
          class="px-8 py-3 bg-black text-white rounded-full font-medium hover:bg-gray-800 transition shadow-lg hover:shadow-xl active:scale-95"
        >
          Start Debate
        </button>
        <button 
          v-else
          @click="stopChat" 
          class="px-8 py-3 bg-red-600 text-white rounded-full font-medium hover:bg-red-700 transition shadow-lg hover:shadow-xl active:scale-95"
        >
          Stop
        </button>
      </div>
    </div>

    <!-- Settings Modal -->
    <SettingsModal 
      :isOpen="isSettingsOpen" 
      :initialConfig="config"
      @close="isSettingsOpen = false"
      @save="saveConfig"
    />
  </div>
</template>
