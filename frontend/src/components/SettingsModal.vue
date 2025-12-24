<script setup>
import { ref, onMounted } from 'vue'

const props = defineProps({
  isOpen: Boolean,
  initialConfig: Object
})

const emit = defineEmits(['close', 'save'])

const activeTab = ref('api') // api, agentA, agentB

const config = ref({
  apiKey: '',
  baseUrl: 'https://api.openai.com/v1',
  model: 'gpt-3.5-turbo',
  agentA: {
    name: 'Agent A',
    prompt: 'You are a helpful assistant.'
  },
  agentB: {
    name: 'Agent B',
    prompt: 'You are a critical thinker.'
  }
})

onMounted(() => {
  if (props.initialConfig) {
    // Deep copy to avoid mutating prop directly
    config.value = JSON.parse(JSON.stringify(props.initialConfig))
  }
})

const save = () => {
  emit('save', config.value)
  emit('close')
}
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
    <div class="bg-white rounded-2xl shadow-2xl w-full max-w-2xl overflow-hidden flex flex-col max-h-[90vh]">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center bg-gray-50/50">
        <h2 class="text-xl font-bold text-gray-800">Settings</h2>
        <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600 transition">
          <span class="text-2xl">&times;</span>
        </button>
      </div>

      <!-- Tabs -->
      <div class="flex border-b border-gray-200">
        <button 
          v-for="tab in ['api', 'agentA', 'agentB']" 
          :key="tab"
          @click="activeTab = tab"
          class="flex-1 py-3 text-sm font-medium transition duration-200 border-b-2"
          :class="activeTab === tab ? 'border-blue-500 text-blue-600 bg-blue-50/30' : 'border-transparent text-gray-500 hover:text-gray-700 hover:bg-gray-50'"
        >
          {{ tab === 'api' ? 'API Config' : (tab === 'agentA' ? 'Agent A' : 'Agent B') }}
        </button>
      </div>

      <!-- Content -->
      <div class="p-6 overflow-y-auto flex-1 space-y-4">
        
        <!-- API Config -->
        <div v-if="activeTab === 'api'" class="space-y-4 animate-fade-in">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Base URL</label>
            <input v-model="config.baseUrl" type="text" placeholder="https://api.openai.com/v1" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition" />
            <p class="text-xs text-gray-500 mt-1">For Deepseek, usage might be: https://api.deepseek.com</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">API Key</label>
            <input v-model="config.apiKey" type="password" placeholder="sk-..." class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Topic (Guide)</label>
            <input v-model="config.topic" type="text" placeholder="e.g., The Future of AI" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition" />
            <p class="text-xs text-gray-500 mt-1">Leave empty for open-ended conversation.</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Model Name</label>
            <input v-model="config.model" type="text" placeholder="gpt-3.5-turbo" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition" />
            <p class="text-xs text-gray-500 mt-1">e.g., deepseek-chat, deepseek-reasoner</p>
          </div>
        </div>

        <!-- Agent Config -->
        <div v-else class="space-y-4 animate-fade-in">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input v-model="config[activeTab].name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">System Prompt</label>
            <textarea v-model="config[activeTab].prompt" rows="8" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition font-mono text-sm"></textarea>
          </div>
        </div>

      </div>

      <!-- Footer -->
      <div class="px-6 py-4 border-t border-gray-100 bg-gray-50 flex justify-end gap-3">
        <button @click="$emit('close')" class="px-4 py-2 text-gray-600 hover:bg-gray-200 rounded-lg transition font-medium">Cancel</button>
        <button @click="save" class="px-6 py-2 bg-black text-white rounded-lg hover:bg-gray-800 transition shadow-lg font-medium">Save Configuration</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.2s ease-out;
}
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(5px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
