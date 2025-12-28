<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const error = ref('')

const handleRegister = async () => {
    if (!username.value || !password.value) return
    try {
        await authStore.register(username.value, password.value)
        alert('Registration successful! Please login.')
        router.push('/login')
    } catch (e) {
        error.value = 'Registration failed: ' + (e.response?.data?.error || e.message)
    }
}
</script>

<template>
    <div class="min-h-screen flex items-center justify-center bg-gray-100">
        <div class="bg-white p-8 rounded-xl shadow-lg w-full max-w-md">
            <h1 class="text-2xl font-bold mb-6 text-center text-gray-800">注册账号</h1>
            
            <div v-if="error" class="bg-red-50 text-red-600 p-3 rounded mb-4 text-sm">{{ error }}</div>

            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">用户名</label>
                    <input v-model="username" type="text" class="w-full border rounded px-3 py-2 focus:ring-2 focus:ring-blue-500 outline-none">
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">密码</label>
                    <input v-model="password" type="password" class="w-full border rounded px-3 py-2 focus:ring-2 focus:ring-blue-500 outline-none">
                </div>
                
                <button @click="handleRegister" class="w-full bg-green-600 text-white py-2 rounded hover:bg-green-700 transition font-bold">注册</button>
                
                <div class="text-center text-sm text-gray-500 mt-4">
                    已有账号? <router-link to="/login" class="text-blue-600 hover:underline">直接登录</router-link>
                </div>
            </div>
        </div>
    </div>
</template>
