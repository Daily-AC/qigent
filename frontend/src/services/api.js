import axios from 'axios'

// Create a single axios instance with the correct base URL
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const api = axios.create({ baseURL: apiBaseUrl })

export default api
