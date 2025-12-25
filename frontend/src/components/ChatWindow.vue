<script setup>
import { computed, ref, watch, nextTick } from 'vue'
import MarkdownIt from 'markdown-it'
import mathjax3 from 'markdown-it-mathjax3'
import hljs from 'highlight.js'
import DOMPurify from 'dompurify'

// Styles
import 'highlight.js/styles/github-dark.css' // or atom-one-dark

const props = defineProps({
  messages: {
    type: Array,
    default: () => []
  }
})

const md = new MarkdownIt({
  html: false,
  linkify: true,
  breaks: true,
  highlight: function (str, lang) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return '<pre class="hljs"><code>' +
               hljs.highlight(str, { language: lang, ignoreIllegals: true }).value +
               '</code></pre>';
      } catch (__) {}
    }

    return '<pre class="hljs"><code>' + md.utils.escapeHtml(str) + '</code></pre>';
  }
})

md.use(mathjax3)

const containerRef = ref(null)
const isUserAtBottom = ref(true)
const showScrollButton = ref(false)

const renderMarkdown = (text) => {
  if (!text) return ''
  // MathJax3 ... (omitted for brevity, assume existing logic)
  const rawHtml = md.render(text)
  return DOMPurify.sanitize(rawHtml)
}

// Handle Scroll Events to detect if user moved up
const handleScroll = () => {
    if (!containerRef.value) return
    const { scrollTop, scrollHeight, clientHeight } = containerRef.value
    // If user is within 50px of bottom, they are "at bottom"
    const atBottom = scrollHeight - scrollTop - clientHeight < 50
    isUserAtBottom.value = atBottom
    showScrollButton.value = !atBottom
}

// Auto scroll only if user was already at bottom
watch(() => props.messages, async () => {
    await nextTick()
    if (isUserAtBottom.value) {
        scrollToBottom()
    }
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
  <div class="relative flex-1 min-h-[500px] max-h-[70vh] flex flex-col">
      <div 
        ref="containerRef" 
        @scroll="handleScroll"
        class="flex-1 overflow-y-auto p-4 space-y-6 bg-gray-50 rounded-lg shadow-inner scroll-smooth"
      >
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

      <!-- Scroll to Bottom Button -->
      <button 
        v-if="showScrollButton"
        @click="scrollToBottom"
        class="absolute bottom-6 right-6 p-3 bg-blue-600 text-white rounded-full shadow-lg hover:bg-blue-700 transition animate-bounce z-10"
        title="Go to latest"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7m0 0l-7-7m7 7V3" />
        </svg>
      </button>
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
