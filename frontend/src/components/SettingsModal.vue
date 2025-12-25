<script setup>
import { ref, watch, toRaw } from 'vue'

const props = defineProps({
  isOpen: Boolean,
  initialConfig: Object
})

const emit = defineEmits(['close', 'save'])

const config = ref({
  apiKey: '',
  baseUrl: 'https://api.openai.com/v1',
  model: 'gpt-3.5-turbo'
})

watch(() => props.initialConfig, (newVal) => {
  if (newVal) {
    config.value = {
        apiKey: newVal.apiKey,
        baseUrl: newVal.baseUrl,
        model: newVal.model
    }
  }
}, { immediate: true, deep: true })

const save = () => {
  emit('save', toRaw(config.value))
  emit('close')
}
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
    <div class="bg-white rounded-2xl shadow-xl w-full max-w-lg p-6 relative">
      <button @click="$emit('close')" class="absolute top-4 right-4 text-gray-400 hover:text-gray-600">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>

      <h2 class="text-xl font-bold mb-6 text-gray-800">System Settings</h2>

      <div class="space-y-4">
          <!-- API Key -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">API Key</label>
            <input 
              v-model="config.apiKey" 
              type="password" 
              placeholder="sk-..." 
              class="w-full px-3 py-2 border border-gray-300 rounded-lg shadow-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
            >
          </div>

          <!-- Base URL -->
          <div>
             <label class="block text-sm font-medium text-gray-700 mb-1">Base URL</label>
             <input 
               v-model="config.baseUrl" 
               type="text" 
               placeholder="https://api.openai.com/v1"
               class="w-full px-3 py-2 border border-gray-300 rounded-lg shadow-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
             >
          </div>

          <!-- Model -->
          <div>
             <label class="block text-sm font-medium text-gray-700 mb-1">Model</label>
             <input 
               v-model="config.model" 
               type="text" 
               placeholder="gpt-3.5-turbo"
               class="w-full px-3 py-2 border border-gray-300 rounded-lg shadow-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
             >
          </div>
      </div>

      <div class="mt-8 flex justify-end gap-3">
        <button @click="$emit('close')" class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg font-medium">Cancel</button>
        <button @click="save" class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 shadow-md font-medium">Save Configuration</button>
      </div>
    </div>
  </div>
</template>
