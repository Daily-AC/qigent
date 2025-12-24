<script setup>
defineProps({
  name: String,
  role: String,
  active: Boolean
})
</script>

<template>
  <div class="p-4 bg-white rounded-xl shadow-sm border border-gray-100 transition-all duration-300" :class="{ 'ring-2 ring-blue-500 transform scale-105': active }">
    <div class="flex items-center gap-3">
      <div class="w-10 h-10 rounded-full bg-gradient-to-br from-indigo-500 to-purple-500 flex items-center justify-center text-white font-bold text-lg">
        {{ name.charAt(0) }}
      </div>
      <div>
        <h3 class="font-semibold text-gray-800">{{ name }}</h3>
        <p class="text-xs text-gray-500">{{ role }}</p>
      </div>
    </div>
  </div>
</template>
