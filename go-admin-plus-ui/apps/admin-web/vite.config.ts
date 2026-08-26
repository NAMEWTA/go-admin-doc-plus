import vue from '@vitejs/plugin-vue'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: process.env.GO_ADMIN_BUILD_DIR ?? join(tmpdir(), 'go-admin-plus-ui', 'admin-web'),
    emptyOutDir: true
  },
  server: {
    host: '127.0.0.1',
    port: 5173
  }
})
