<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const props = defineProps({
  activeId: String
})

const emit = defineEmits(['select', 'new-chat', 'delete', 'role-market'])

const conversations = ref([])

// Use a relative path or configured axios instance
const api = axios.create({ baseURL: 'http://localhost:8080' })

const loadConversations = async () => {
  try {
    const res = await api.get('/conversations')
    // sort by createdAt desc
    conversations.value = res.data.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt))
  } catch (e) {
    console.error('Failed to load conversations', e)
  }
}

const deleteConv = async (e, id) => {
  e.stopPropagation()
  if (!confirm('Are you sure you want to delete this conversation?')) return
  try {
    await api.delete(`/conversations/${id}`)
    await loadConversations()
    emit('delete', id)
  } catch (e) {
    console.error('Failed to delete', e)
  }
}

onMounted(() => {
  loadConversations()
})

// Expose reload
defineExpose({ reload: loadConversations })
</script>

<template>
  <div class="w-64 bg-gray-900 h-screen flex flex-col text-white flex-shrink-0">
    <!-- Header -->
    <div class="p-4 border-b border-gray-800 flex items-center justify-between">
      <h2 class="font-bold text-lg tracking-wide">Qigent</h2>
      <button 
        @click="$emit('new-chat')" 
        class="p-2 bg-blue-600 hover:bg-blue-500 rounded-lg transition shadow-md"
        title="New Chat"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" />
        </svg>
      </button>
    </div>

    <!-- List -->
    <div class="flex-1 overflow-y-auto py-2">
      <div v-if="conversations.length === 0" class="text-gray-500 text-sm p-4 text-center italic">
        No conversations yet. Start a new one!
      </div>
      <div 
        v-for="conv in conversations" 
        :key="conv.id"
        @click="$emit('select', conv.id)"
        class="group px-4 py-3 cursor-pointer transition relative hover:bg-gray-800"
        :class="activeId === conv.id ? 'bg-gray-800 border-l-4 border-blue-500' : 'border-l-4 border-transparent'"
      >
        <div class="text-sm font-medium truncate pr-6">{{ conv.topic || 'Untitled Chat' }}</div>
        <div class="text-xs text-gray-500 mt-1 flex justify-between">
           <span>{{ new Date(conv.createdAt).toLocaleDateString() }}</span>
           <span class="text-[10px] bg-gray-700 px-1 rounded">{{ (conv.agentA.name && conv.agentB.name) ? `${conv.agentA.name} vs ${conv.agentB.name}` : 'Unknown' }}</span>
        </div>
        
        <!-- Delete Button -->
        <button 
          @click="(e) => deleteConv(e, conv.id)"
          class="absolute right-2 top-3 text-gray-600 hover:text-red-400 opacity-0 group-hover:opacity-100 transition p-1"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
          </svg>
        </button>
      </div>
    </div>
    
    <!-- Footer Actions -->
    <div class="p-4 border-t border-gray-800 space-y-2">
      <button 
        @click="$emit('role-market')"
        class="w-full flex items-center justify-center gap-2 py-2 bg-gray-800 hover:bg-gray-700 rounded-lg text-sm text-gray-300 transition"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
        </svg>
        Role Market
      </button>
    </div>
    
    <!-- Footer User Info -->
    <div class="px-4 pb-4 text-xs text-gray-500 text-center">
      v1.0.0 MVP
    </div>
  </div>
</template>
