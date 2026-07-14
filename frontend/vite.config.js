import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api/auth': 'http://localhost:8081',
      '/api/guestbook': 'http://localhost:8082'
    }
  }
})
