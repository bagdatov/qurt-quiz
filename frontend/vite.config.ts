import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const backendHttp = process.env.BACKEND_URL ?? 'http://localhost:8080'
const backendWs   = backendHttp.replace(/^http/, 'ws')

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': backendHttp,
      '/ws':  { target: backendWs, ws: true },
    },
  },
})
