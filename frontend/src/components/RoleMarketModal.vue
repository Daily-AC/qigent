<script setup>
import { ref, onMounted, watch } from 'vue'
import api from '../services/api'

const roles = ref([])
const isAdding = ref(false)
const newRole = ref({
  name: '',
  prompt: '',
  avatar: '' 
})

const loadRoles = async () => {
  try {
    const res = await api.get('/roles')
    roles.value = res.data
  } catch (e) {
    console.error('Failed to load roles', e)
  }
}

watch(() => props.isOpen, (newVal) => {
  if (newVal) loadRoles()
})

const addRole = async () => {
  if (!newRole.value.name || !newRole.value.prompt) return
  
  try {
    await api.post('/roles', newRole.value)
    newRole.value = { name: '', prompt: '', avatar: '' }
    isAdding.value = false
    loadRoles()
  } catch (e) {
    console.error('Failed to add role', e)
    alert('Failed to add role')
  }
}

const deleteRole = async (name) => {
  if (!confirm(`Delete role "${name}"?`)) return
  try {
    await api.delete(`/roles/${name}`)
    loadRoles()
  } catch (e) {
    console.error('Failed to delete role', e)
  }
}
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
    <div class="bg-white rounded-2xl shadow-xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
      <div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center bg-gray-50">
        <h2 class="text-xl font-bold text-gray-800">Role Market</h2>
        <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600 font-bold text-xl">&times;</button>
      </div>

      <div class="flex-1 overflow-y-auto p-6 bg-gray-50">
        <!-- Add New Role Form -->
        <div v-if="isAdding" class="mb-8 bg-white p-6 rounded-xl shadow-sm border border-blue-100">
           <h3 class="font-bold text-lg mb-4 text-blue-600">Add New Role</h3>
           <div class="space-y-4">
             <div>
               <label class="block text-sm font-medium text-gray-700">Name</label>
               <input v-model="newRole.name" type="text" class="mt-1 w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 outline-none" placeholder="e.g. Batman">
             </div>
             <div>
               <label class="block text-sm font-medium text-gray-700">System Prompt</label>
               <textarea v-model="newRole.prompt" rows="3" class="mt-1 w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 outline-none" placeholder="You are Batman..."></textarea>
             </div>
             <div class="flex justify-end gap-3">
               <button @click="isAdding = false" class="px-4 py-2 text-gray-500 hover:text-gray-700">Cancel</button>
               <button @click="addRole" class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">Save Role</button>
             </div>
           </div>
        </div>

        <!-- Role Grid -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
           <!-- Add Button Card -->
           <div 
             v-if="!isAdding"
             @click="isAdding = true"
             class="border-2 border-dashed border-gray-300 rounded-xl flex flex-col items-center justify-center h-48 cursor-pointer hover:border-blue-400 hover:bg-blue-50 transition group"
            >
              <div class="h-12 w-12 rounded-full bg-blue-100 text-blue-500 flex items-center justify-center mb-2 group-hover:bg-blue-200">
                 <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                   <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                 </svg>
              </div>
              <span class="font-medium text-gray-500 group-hover:text-blue-600">Create New Role</span>
           </div>

           <!-- Role Cards -->
           <div v-for="role in roles" :key="role.name" class="bg-white p-5 rounded-xl shadow-sm border border-gray-100 hover:shadow-md transition relative group">
             <div class="flex items-center gap-3 mb-3">
               <div class="h-10 w-10 rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 text-white flex items-center justify-center font-bold text-lg">
                 {{ role.name[0]?.toUpperCase() }}
               </div>
               <h3 class="font-bold text-gray-800">{{ role.name }}</h3>
             </div>
             <p class="text-sm text-gray-500 line-clamp-3 h-16 leading-relaxed bg-gray-50 p-2 rounded">
               {{ role.prompt }}
             </p>
             
             <button 
               @click.stop="deleteRole(role.name)"
               class="absolute top-4 right-4 text-gray-300 hover:text-red-500 transition opacity-0 group-hover:opacity-100"
               title="Delete Role"
              >
               <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                 <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
               </svg>
             </button>
           </div>
        </div>
      </div>
    </div>
  </div>
</template>
