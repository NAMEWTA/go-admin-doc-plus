import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath } from 'node:url'
import { resolve } from 'node:path'

const output = process.env.GO_ADMIN_DEMO_E2E_OUT_DIR
if (!output) throw new Error('GO_ADMIN_DEMO_E2E_OUT_DIR is required')
const workspaceRoot = fileURLToPath(new URL('../../..', import.meta.url))
export default defineConfig({
  root: fileURLToPath(new URL('.', import.meta.url)),
  plugins: [vue()],
  resolve: { alias: {
    '@go-admin/web-domain-demo': resolve(workspaceRoot, 'packages/web-domains/demo/src/index.ts'),
    '@go-admin/domain-demo': resolve(workspaceRoot, 'packages/domains/demo/src/index.ts'),
    '@go-admin/ui': resolve(workspaceRoot, 'packages/ui/src/index.ts'),
    vue: resolve(workspaceRoot, 'packages/web-domains/demo/node_modules/vue'),
  } },
  build: { outDir: output, emptyOutDir: true },
})
