<script setup>
import { computed, ref, watch, nextTick } from 'vue'
import MarkdownIt from 'markdown-it'
import DOMPurify from 'dompurify'

const props = defineProps({
  messages: {
    type: Array,
    default: () => []
  }
})

const md = new MarkdownIt({
  html: false,
  linkify: true,
  breaks: true
})

const containerRef = ref(null)

const renderMarkdown = (text) => {
  if (!text) return ''
  const rawHtml = md.render(text)
  return DOMPurify.sanitize(rawHtml)
}

// Auto scroll
watch(() => props.messages, async () => {
    await nextTick()
    scrollToBottom()
  }, 
  { deep: true }
)

const scrollToBottom = () => {
  if (containerRef.value) {
    containerRef.value.scrollTop = containerRef.value.scrollHeight
  }
}
</script>

<template>
  <div ref="containerRef" class="flex-1 overflow-y-auto p-4 space-y-6 bg-gray-50 rounded-lg shadow-inner min-h-[500px] max-h-[70vh] scroll-smooth">
    <div 
      v-for="(msg, index) in messages" 
      :key="index"
      class="flex flex-col animate-fade-in-up"
      :class="msg.sender === '苏格拉底' ? 'items-start' : (msg.sender === 'System' ? 'items-center' : 'items-end')"
    >
      <!-- System Message -->
      <div v-if="msg.sender === 'System'" class="bg-gray-200 text-gray-600 px-4 py-1 rounded-full text-xs font-mono mb-2">
        {{ msg.content }}
      </div>

      <!-- Agent Message -->
      <div v-else class="max-w-[85%] flex flex-col" :class="msg.sender === '苏格拉底' ? 'items-start' : 'items-end'">
        <span class="text-xs text-gray-400 mb-1 px-1">{{ msg.sender }}</span>
        
        <div 
          class="px-5 py-3 rounded-2xl shadow-sm text-sm leading-relaxed prose prose-sm max-w-none break-words"
          :class="[
            msg.sender === '苏格拉底' 
              ? 'bg-white rounded-tl-none border border-gray-200 text-gray-800' 
              : 'bg-blue-600 text-white rounded-tr-none prose-invert'
          ]"
          v-html="renderMarkdown(msg.content)"
        ></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.animate-fade-in-up {
  animation: fadeInUp 0.3s ease-out;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Custom Scrollbar */
div::-webkit-scrollbar {
  width: 6px;
}
div::-webkit-scrollbar-track {
  background: transparent;
}
div::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 20px;
}
</style>
