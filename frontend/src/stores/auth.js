import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../services/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const username = ref(localStorage.getItem('username') || '')

  function setToken(t, u) {
    token.value = t
    username.value = u
    localStorage.setItem('token', t)
    localStorage.setItem('username', u)
    
    // Set axios default header?
    // Better to use interceptor in api.js
  }

  function logout() {
    token.value = ''
    username.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('username')
  }

  async function login(usernameInput, password) {
    const res = await api.post('/auth/login', { username: usernameInput, password })
    setToken(res.data.token, res.data.username)
  }

  async function register(usernameInput, password) {
    await api.post('/auth/register', { username: usernameInput, password })
    // Auto login?
  }

  return { token, username, login, register, logout }
})
