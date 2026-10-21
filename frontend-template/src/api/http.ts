import axios from 'axios'

// Authentication and error semantics will be added once the API contract is finalized.
export const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 10_000,
})
