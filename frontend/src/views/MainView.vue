<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useChatStore } from '../stores/chat'
import { useAuthStore } from '../stores/auth'
import ChatWindow from '../components/ChatWindow.vue'
import AgentCard from '../components/AgentCard.vue'
import SettingsModal from '../components/SettingsModal.vue'
import Sidebar from '../components/Sidebar.vue'
import NewChatModal from '../components/NewChatModal.vue'
import RoleMarketModal from '../components/RoleMarketModal.vue'
import api from '../services/api'

const router = useRouter()
const chatStore = useChatStore()
const authStore = useAuthStore()
const sidebarRef = ref(null)

// State
const isSettingsOpen = ref(false)
const isNewChatOpen = ref(false)
const isRoleMarketOpen = ref(false)
const activeConversationId = ref(null)
const currentTopic = ref('')
const currentAgents = ref({
  agentA: { name: 'Agent A', prompt: '' },
  agentB: { name: 'Agent B', prompt: '' }
})

// Global Config (API Key)
const globalConfig = ref({
  apiKey: '', 
  baseUrl: 'https://api.openai.com/v1',
  model: 'gpt-3.5-turbo'
})

onMounted(async () => {
    // Check Auth? Router handles it, but we can verify
    // Load Global Config
    try {
        const res = await api.get('/config')
        globalConfig.value = res.data
        if (!globalConfig.value.apiKey) {
            isSettingsOpen.value = true
        }
    } catch (e) {
        console.error('Failed to load global config', e)
    }
})

// Actions
const handleSelectConversation = async (id) => {
  if (activeConversationId.value === id) return

  if (chatStore.isConnected) {
    chatStore.disconnect()
  }

  try {
    const res = await api.get(`/conversations/${id}`)
    const conv = res.data
    
    activeConversationId.value = id
    currentTopic.value = conv.topic
    currentAgents.value.agentA = conv.agentA
    currentAgents.value.agentB = conv.agentB
    
    chatStore.messages = conv.history || []
    
  } catch (e) {
    console.error('Failed to load conversation', e)
    if (e.response?.status === 403) {
        alert('Forbidden: You do not have access to this conversation')
    }
  }
}

const handleCreateConversation = async (payload) => {
  try {
    const res = await api.post('/conversations', payload)
    isNewChatOpen.value = false
    sidebarRef.value.reload()
    handleSelectConversation(res.data.id)
  } catch (e) {
    console.error('Failed to create conversation', e)
  }
}

const handleStartParams = () => {
  const apiBase = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
  const wsProtocol = apiBase.startsWith('https') ? 'wss:' : 'ws:'
  const host = apiBase.replace(/^https?:\/\//, '')
  
  // Handshake payload (for WS Auth via Query)
  // We need to pass token in query param for WS
  const token = authStore.token
  const wsUrl = `${wsProtocol}//${host}/ws/chat?conversationId=${activeConversationId.value}&token=${token}`
  
  const handshake = {
    apiKey: globalConfig.value.apiKey,
    baseUrl: globalConfig.value.baseUrl,
    model: globalConfig.value.model
  }
  
  return { wsUrl, handshake }
}

const startChat = () => {
  if (!activeConversationId.value) {
    alert('Select or create a conversation first.')
    return
  }
  if (!globalConfig.value.apiKey) {
    isSettingsOpen.value = true
    return
  }
  
  const { wsUrl, handshake } = handleStartParams()
  chatStore.connect(wsUrl, handshake)
}

const stopChat = () => {
  chatStore.disconnect()
}

const handleDelete = (id) => {
  if (activeConversationId.value === id) {
    activeConversationId.value = null
    currentTopic.value = ''
    chatStore.messages = []
    chatStore.disconnect()
  }
}

const handleExport = () => {
  if (!chatStore.messages.length) return
  
  const content = chatStore.messages.map(m => {
    if (m.sender === 'System') return `[System]: ${m.content}`
    return `### ${m.sender}\n${m.content}\n`
  }).join('\n')
  
  const blob = new Blob([content], { type: 'text/markdown' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `debate-${activeConversationId.value || 'export'}.md`
  a.click()
  URL.revokeObjectURL(url)
}

const userInjectionInput = ref('')

const handleUserInjection = () => {
  if (!userInjectionInput.value.trim()) return
  chatStore.sendMessage(userInjectionInput.value)
  userInjectionInput.value = ''
}

const saveGlobalConfig = async (newConfig) => {
  try {
    await api.post('/config', newConfig)
    globalConfig.value = newConfig
    isSettingsOpen.value = false
  } catch (e) {
    console.error('Failed to save config', e)
    alert('Failed to save settings')
  }
}

const concludeChat = () => {
    chatStore.conclude()
}

const handleLogout = () => {
    authStore.logout()
    router.push('/login')
}
</script>

<template>
  <div class="flex h-screen bg-gray-100 font-sans overflow-hidden">
    <!-- Sidebar -->
    <Sidebar 
      ref="sidebarRef"
      :activeId="activeConversationId"
      @select="handleSelectConversation"
      @new-chat="isNewChatOpen = true"
      @delete="handleDelete"
      @role-market="isRoleMarketOpen = true"
    >
        <!-- Custom slot or append for logout? Sidebar doesn't have slot yet. 
             Ideally create a slot in Sidebar for bottom actions 
        -->
    </Sidebar>

    <!-- Main Content -->
    <div class="flex-1 flex flex-col h-screen relative">
      <!-- Toolbar / Header -->
      <header class="h-16 bg-white border-b border-gray-200 flex justify-between items-center px-6 shadow-sm z-10">
        <div class="flex flex-col">
          <h1 class="font-bold text-gray-800 text-lg">{{ currentTopic || 'Qigent Debate Platform' }}</h1>
          <p v-if="activeConversationId" class="text-xs text-gray-400">ID: {{ activeConversationId.slice(0, 8) }}...</p>
        </div>
        
        <div class="flex items-center gap-4">
           <!-- Agent Status Indicators -->
           <div v-if="activeConversationId" class="flex gap-2 mr-4">
              <span class="px-2 py-1 bg-blue-50 text-blue-600 text-xs rounded border border-blue-100 font-medium">{{ currentAgents.agentA.name }}</span>
              <span class="text-gray-300">vs</span>
              <span class="px-2 py-1 bg-indigo-50 text-indigo-600 text-xs rounded border border-indigo-100 font-medium">{{ currentAgents.agentB.name }}</span>
           </div>
           
           <span class="text-sm font-bold text-gray-700 bg-gray-100 px-3 py-1 rounded-full flex items-center gap-2">
               <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path></svg>
               {{ authStore.username }}
           </span>
           <button @click="handleLogout" class="text-sm text-red-500 hover:text-red-700 font-medium">退出</button>

           <!-- Export Button -->
           <button 
             v-if="activeConversationId" 
             @click="handleExport" 
             class="p-2 hover:bg-gray-100 rounded-lg text-gray-500" 
             title="Export Chat"
           >
             <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
             </svg>
           </button>

           <button @click="isSettingsOpen = true" class="p-2 hover:bg-gray-100 rounded-lg text-gray-500">
             <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
               <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543 .826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
               <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
             </svg>
           </button>
        </div>
      </header>

      <!-- Chat Area -->
      <div class="flex-1 overflow-hidden p-6 bg-gray-100 flex flex-col">
        <div v-if="!activeConversationId" class="flex-1 flex flex-col items-center justify-center text-gray-400 space-y-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
            <p>请从侧边栏选择对话，或开始新的辩论。</p>
            <button @click="isNewChatOpen = true" class="px-6 py-2 bg-blue-600 text-white rounded-lg shadow-lg hover:bg-blue-700 transition">新建辩论</button>
        </div>
        <div v-else class="h-full flex flex-col space-y-4">
            <ChatWindow :messages="chatStore.messages" class="flex-1 shadow-sm border border-gray-200 bg-white" />
            
            <!-- Controls & Input -->
            <div class="flex flex-col items-center gap-4 pb-2">
                <div class="flex gap-4">
                  <button 
                  v-if="!chatStore.isConnected"
                  @click="startChat" 
                  class="px-8 py-3 bg-black text-white rounded-full font-medium hover:bg-gray-800 transition shadow-lg flex items-center gap-2"
                  >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd" />
                  </svg>
                  继续 / 开始辩论
                  </button>
                  
                  <template v-else>
                    <button 
                    @click="stopChat" 
                    class="px-8 py-3 bg-red-600 text-white rounded-full font-medium hover:bg-red-700 transition shadow-lg flex items-center gap-2"
                    >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zM7 8a1 1 0 012 0v4a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                    </svg>
                    暂停辩论
                    </button>
                    <button 
                    @click="concludeChat" 
                    class="px-6 py-3 bg-yellow-600 text-white rounded-full font-medium hover:bg-yellow-700 transition shadow-lg flex items-center gap-2"
                    >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                    </svg>
                    结束辩论
                    </button>
                  </template>
                </div>
                
                <!-- God Mode Input -->
                <div v-if="chatStore.isConnected" class="w-full max-w-3xl flex gap-2 animate-fade-in-up">
                  <input 
                    v-model="userInjectionInput" 
                    @keyup.enter="handleUserInjection"
                    type="text" 
                    placeholder="开启上帝模式: 输入你的观点直接干预辩论..." 
                    class="flex-1 px-4 py-3 rounded-xl border border-gray-300 shadow-sm focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none"
                  >
                  <button 
                    @click="handleUserInjection"
                    class="px-4 py-2 bg-purple-600 text-white rounded-xl hover:bg-purple-700 shadow transition font-medium disabled:opacity-50"
                    :disabled="!userInjectionInput"
                  >
                    发送 (Inject)
                  </button>
                </div>
            </div>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <SettingsModal 
      :isOpen="isSettingsOpen" 
      :initialConfig="globalConfig"
      @close="isSettingsOpen = false"
      @save="saveGlobalConfig"
    />
    
    <NewChatModal
        :isOpen="isNewChatOpen"
        @close="isNewChatOpen = false"
        @create="handleCreateConversation"
    />
    
    <RoleMarketModal
      :isOpen="isRoleMarketOpen"
      @close="isRoleMarketOpen = false"
    />
  </div>
</template>
