#!/usr/bin/env node

import { existsSync, readFileSync, readdirSync } from 'node:fs'
import { dirname, extname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '..', '..')
const removedFrontendFacades = [
  'apps/admin/src/api/admin/dict/type.ts',
  'apps/admin/src/api/admin/sys-api.ts',
  'apps/admin/src/api/admin/sys-dept.ts',
  'apps/admin/src/api/admin/sys-login-log.ts',
  'apps/admin/src/api/admin/sys-menu.ts',
  'apps/admin/src/api/admin/sys-opera-log.ts',
  'apps/admin/src/api/admin/sys-post.ts',
  'apps/admin/src/api/demo/product.ts',
  'apps/admin/src/api/job/sys-job.ts',
  'apps/admin/src/api/monitor/server.ts',
  'apps/admin/src/api/tools/gen.ts',
  'apps/admin/src/api/ws.ts'
]
const retainedFrontendFacades = {
  'apps/admin/src/api/admin/dict/data.ts': '@/api/admin/dict/data',
  'apps/admin/src/api/admin/sys-config.ts': '@/api/admin/sys-config',
  'apps/admin/src/api/admin/sys-role.ts': '@/api/admin/sys-role',
  'apps/admin/src/api/admin/sys-user.ts': '@/api/admin/sys-user'
}

const failures = []
const backendApi = join(ROOT, 'go-admin-plus/cmd/api')
const frontend = join(ROOT, 'go-admin-ui-plus')

const walk = directory => readdirSync(directory, { withFileTypes: true }).flatMap(entry => {
  const path = join(directory, entry.name)
  if (entry.name === 'node_modules' || entry.name === '.git') return []
  return entry.isDirectory() ? walk(path) : [path]
})

const sourceFiles = walk(frontend).filter(path => ['.js', '.ts', '.vue'].includes(extname(path)))
const source = sourceFiles.map(path => ({ path, text: readFileSync(path, 'utf8') }))

for (const file of walk(backendApi).filter(path => path.endsWith('.go'))) {
  if (/\bAppRouters\b/.test(readFileSync(file, 'utf8'))) failures.push(`backend legacy AppRouters remains in ${relative(ROOT, file)}`)
}

for (const file of removedFrontendFacades) {
  if (existsSync(join(frontend, file))) failures.push(`zero-consumer facade still exists: ${file}`)
}

const residuals = []
for (const [file, importPath] of Object.entries(retainedFrontendFacades)) {
  if (!existsSync(join(frontend, file))) failures.push(`active facade was removed: ${file}`)
  const consumers = source.filter(item => !item.path.endsWith(file) && item.text.includes(importPath))
  if (consumers.length === 0) failures.push(`retained facade has no observable consumer: ${file}`)
  residuals.push({ kind: 'active-api-facade', path: file, consumers: consumers.map(item => relative(frontend, item.path)) })
}

const requestConsumers = source
  .filter(item => /import\s+request\s+from\s+['"]@\/utils\/request['"]/.test(item.text))
  .map(item => relative(frontend, item.path))
if (requestConsumers.length === 0) failures.push('request compatibility facade has no observable consumer and should be removed')
residuals.push({ kind: 'active-request-facade', path: 'apps/admin/src/utils/request.ts', consumers: requestConsumers })

const permissionPath = join(frontend, 'apps/admin/src/stores/permission.ts')
const permissionSource = readFileSync(permissionPath, 'utf8')
if (!permissionSource.includes('legacyComponent') || !permissionSource.includes('viewsModules')) {
  failures.push('expected active legacy component fallback was not found')
}
residuals.push({
  kind: 'active-component-fallback',
  path: 'apps/admin/src/stores/permission.ts',
  reason: 'Shell routes and pre-routeKey menu rows still require exact-whitelist fallback'
})

const checkerPath = join(frontend, 'scripts/check-api-contract.mjs')
const checkerSource = readFileSync(checkerPath, 'utf8')
if (!checkerSource.includes('matchAll(')) failures.push('expected active heuristic API contract checker was not found')
residuals.push({
  kind: 'active-heuristic-contract-check',
  path: 'scripts/check-api-contract.mjs',
  reason: 'It covers Go models and frontend fields that the canonical OpenAPI does not yet describe'
})

const result = {
  schema_version: 1,
  removed: {
    backend_app_routers: true,
    frontend_zero_consumer_facades: removedFrontendFacades
  },
  residuals
}

if (failures.length) {
  console.error(`GO_ADMIN_COMPATIBILITY_SCAN_FAIL\n${failures.join('\n')}`)
  process.exit(1)
}
console.log(JSON.stringify(result, null, 2))
console.log('GO_ADMIN_COMPATIBILITY_SCAN_PASS')
