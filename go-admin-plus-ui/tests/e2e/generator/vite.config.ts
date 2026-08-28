import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath } from 'node:url'
import { resolve } from 'node:path'
const output = process.env.GO_ADMIN_GENERATOR_E2E_OUT_DIR
if (!output) throw new Error('GO_ADMIN_GENERATOR_E2E_OUT_DIR is required')
const workspaceRoot = fileURLToPath(new URL('../../..', import.meta.url))
export default defineConfig({ root: fileURLToPath(new URL('.', import.meta.url)), plugins:[vue()], resolve:{alias:{'@go-admin-plus/web-domain-generator':resolve(workspaceRoot,'packages/web-domains/generator/src/index.ts'),'@go-admin-plus/domain-generator':resolve(workspaceRoot,'packages/domains/generator/src/index.ts'),'@go-admin-plus/domain-iam/administration':resolve(workspaceRoot,'packages/domains/iam/src/administration/index.ts'),'@go-admin-plus/domain-iam/session':resolve(workspaceRoot,'packages/domains/iam/src/session/index.ts'),'@go-admin-plus/web-domain-iam/administration':resolve(workspaceRoot,'packages/web-domains/iam/src/administration/index.ts'),'@go-admin-plus/web-domain-iam/session':resolve(workspaceRoot,'packages/web-domains/iam/src/session/index.ts'),'@go-admin-plus/ui':resolve(workspaceRoot,'packages/ui/src/index.ts'),vue:resolve(workspaceRoot,'packages/web-domains/generator/node_modules/vue')}},server:{proxy:process.env.GO_ADMIN_GENERATOR_E2E_API_TARGET?{'/api':{target:process.env.GO_ADMIN_GENERATOR_E2E_API_TARGET,secure:false}}:undefined},build:{outDir:output,emptyOutDir:true} })
