<script setup>
import { ref, onMounted, watch } from 'vue'
import api from '../services/api'

const props = defineProps({
  isOpen: Boolean
})

const emit = defineEmits(['close', 'create'])

const roles = ref([])
const topic = ref('')
const selectedAgentA = ref(null)
const selectedAgentB = ref(null)

const loadRoles = async () => {
  try {
    const res = await api.get('/roles')
    roles.value = res.data
    // Only set defaults if not yet set or if current selection is invalid (optional optimization)
    // For simplicity, if selections are empty, defaulting is fine. 
    // If we want to persist user selection across re-opens, check if selectedAgentA is in new list.
    if (!selectedAgentA.value && roles.value.length >= 1) selectedAgentA.value = roles.value[0]
    if (!selectedAgentB.value && roles.value.length >= 2) selectedAgentB.value = roles.value[1]
  } catch (e) {
    console.error('Failed to load roles', e)
  }
}

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    loadRoles()
  }
})

onMounted(() => {
  loadRoles()
})

const create = () => {
  if (!topic.value) {
    alert('Please enter a topic')
    return
  }
  if (!selectedAgentA.value || !selectedAgentB.value) {
    alert('Please select two agents')
    return
  }
  
  emit('create', {
    topic: topic.value,
    agentA: selectedAgentA.value,
    agentB: selectedAgentB.value
  })
}
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
    <div class="bg-white rounded-2xl shadow-xl w-full max-w-2xl overflow-hidden flex flex-col">
      <div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center bg-gray-50">
        <h2 class="text-xl font-bold text-gray-800">New Debate</h2>
        <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600">&times;</button>
      </div>

      <div class="p-6 space-y-6">
        <!-- Topic -->
        <div>
          <label class="block text-sm font-bold text-gray-700 mb-2">Topic</label>
          <input 
            v-model="topic" 
            type="text" 
            placeholder="What should they argue about?" 
            class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none text-lg"
          />
        </div>

        <!-- Role Selection -->
        <div class="grid grid-cols-2 gap-8">
          <!-- Agent A -->
          <div>
            <label class="block text-sm font-bold text-blue-600 mb-2">Agent A (Left)</label>
            <select v-model="selectedAgentA" class="w-full px-3 py-2 border border-gray-300 rounded-lg mb-2">
              <option v-for="r in roles" :key="r.name" :value="r">{{ r.name }}</option>
            </select>
            <div v-if="selectedAgentA" class="text-xs text-gray-500 bg-gray-50 p-2 rounded h-20 overflow-y-auto">
              {{ selectedAgentA.prompt }}
            </div>
          </div>

          <!-- Agent B -->
          <div>
            <label class="block text-sm font-bold text-indigo-600 mb-2">Agent B (Right)</label>
            <select v-model="selectedAgentB" class="w-full px-3 py-2 border border-gray-300 rounded-lg mb-2">
              <option v-for="r in roles" :key="r.name" :value="r">{{ r.name }}</option>
            </select>
            <div v-if="selectedAgentB" class="text-xs text-gray-500 bg-gray-50 p-2 rounded h-20 overflow-y-auto">
              {{ selectedAgentB.prompt }}
            </div>
          </div>
        </div>
      </div>

      <div class="px-6 py-4 border-t border-gray-100 bg-gray-50 flex justify-end gap-3">
        <button @click="$emit('close')" class="px-4 py-2 text-gray-600 hover:bg-gray-200 rounded-lg">Cancel</button>
        <button @click="create" class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 shadow-lg font-medium">Create Debate</button>
      </div>
    </div>
  </div>
</template>
