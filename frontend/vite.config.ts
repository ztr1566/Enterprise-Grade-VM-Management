import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true,
        // Suppress noisy ECONNRESET / EPIPE errors that occur when the backend
        // (SSH terminal, log stream) closes a WebSocket connection normally.
        configure: (proxy) => {
          proxy.on('error', (err: NodeJS.ErrnoException) => {
            if (err.code === 'ECONNRESET' || err.code === 'EPIPE') return
            console.error('[proxy error]', err.message)
          })
        },
      },
    },
  },
})

