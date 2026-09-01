#!/usr/bin/env node

import { existsSync, readFileSync, readdirSync } from 'node:fs'
import { dirname, extname, join, relative, resolve, sep } from 'node:path'
import { fileURLToPath } from 'node:url'

const removedPaths = [
  'go-admin-ui-plus', 'go-admin-plus/app', 'go-admin-plus/common', 'go-admin-plus/api',
  'go-admin-plus/internal/tenant', 'go-admin-plus/internal/profile', 'go-admin-plus/cmd/go-admin-desktop',
  'go-admin-plus-ui/apps/admin-desktop/src-tauri/src/demo_contract.rs'
]
const forbidden = [
  ['old frontend name', /go-admin-ui-plus/], ['Wails runtime', /\bwails(?:app)?\b/i],
  ['old Go module path', /(?:^|\n)\s*module\s+go-admin\s*(?:\n|$)|["']go-admin\/internal\//],
  ['old frontend package scope', /@go-admin\//],
  ['old upstream core', /go-admin-core/], ['Casbin', /\bcasbin\b/i],
  ['Redis', /\bredis\b/i], ['tenant feature', /\btenant(?:s|_id)?\b/i],
  ['MySQL', /\bmysql\b/i], ['SQL Server', /\bsqlserver\b|\bsql server\b/i],
  ['JWT', /\bjwts?\b/i], ['refresh token', /\brefresh[_ -]?tokens?\b/i],
  ['AutoMigrate', /\bauto[_ -]?migrate\b/i], ['GORM', /gorm\.io\/|\bgorm\b/i],
  ['old SQLite build tag', /(?:-tags[= ]+|build:?)sqlite3\b/i],
  ['old release class', /unsigned-self-use/i]
]
const textExtensions = new Set([
  '', '.cjs', '.cmd', '.css', '.env', '.example', '.go', '.html', '.js', '.json', '.lock',
  '.md', '.mjs', '.mod', '.ps1', '.rs', '.sh', '.sql', '.sum', '.toml', '.ts', '.tsx', '.txt', '.vue',
  '.xml', '.yaml', '.yml'
])
const ignoredDirectories = new Set([
  '.artifacts', '.data', '.git', '.playwright', 'coverage', 'dist', 'node_modules', 'speculo', 'target', 'vendor'
])
const allowedMatches = new Map(Object.entries({
  'NOTICE.md': ['old upstream core'],
  'go-admin-plus-ui/apps/admin-desktop/src-tauri/src/proxy.rs': ['MySQL', 'refresh token'],
  'go-admin-plus/internal/application/architecture_test.go': ['Wails runtime'],
  'go-admin-plus/internal/modules/files/migrations/0010-files/provider_test.go': ['tenant feature'],
  'go-admin-plus/internal/modules/files/migrations/0020-capacity/provider_test.go': ['tenant feature'],
  'go-admin-plus/internal/modules/generator/generator_test.go': ['Casbin', 'Redis', 'tenant feature', 'GORM'],
  'go-admin-plus/internal/modules/generator/writer.go': ['Casbin', 'Redis', 'tenant feature', 'GORM'],
  'go-admin-plus/internal/modules/iam/authorization/capability_registry_test.go': ['MySQL'],
  'go-admin-plus/internal/modules/organization/migrations/provider_test.go': ['Casbin', 'tenant feature', 'JWT'],
  'go-admin-plus/internal/modules/settings/security.go': ['MySQL', 'refresh token'],
  'go-admin-plus/internal/modules/settings/service_test.go': ['JWT'],
  'go-admin-plus/internal/platform/logging/redaction.go': ['MySQL'],
  'go-admin-plus/test/demo/products_sqlite_test.go': ['tenant feature'],
  'go-admin-plus/test/iam/authorization/administration_test.go': ['Casbin', 'tenant feature', 'JWT'],
  'release/macos/README.md': ['Wails runtime'],
  'scripts/quality/architecture-check.mjs': ['old frontend name'],
  'scripts/quality/architecture-check.test.mjs': ['old frontend package scope'],
  'scripts/release/linux/verify-policy.mjs': ['old frontend name'],
  'scripts/release/macos/verify-policy.mjs': ['Wails runtime', 'old release class'],
  'scripts/release/macos/verify-policy.test.mjs': ['old release class'],
  'scripts/release/windows/verify-policy.mjs': ['old frontend name', 'Wails runtime', 'old release class'],
  'scripts/release/windows/verify-policy.test.mjs': ['Wails runtime', 'old release class']
}).map(([path, names]) => [path, new Set(names)]))

const walk = directory => {
  if (!existsSync(directory)) return []
  return readdirSync(directory, { withFileTypes: true }).flatMap(entry => {
    if (ignoredDirectories.has(entry.name)) return []
    const path = join(directory, entry.name)
    return entry.isDirectory() ? walk(path) : [path]
  })
}

export const checkCompatibility = root => {
  const failures = []
  for (const path of removedPaths) if (existsSync(join(root, path))) failures.push(`removed path still exists: ${path}`)
  const scanRoots = ['.']
  const ownFiles = new Set(['scripts/quality/compatibility-zero.mjs', 'scripts/quality/compatibility-zero.test.mjs'])
  for (const rootPath of scanRoots) {
    const absolute = join(root, rootPath)
    const files = existsSync(absolute) && !readdirSafe(absolute) ? [absolute] : walk(absolute)
    for (const file of files) {
      const path = relative(root, file).split(sep).join('/')
      if (ownFiles.has(path) || !textExtensions.has(extname(file))) continue
      const content = readFileSync(file)
      if (content.includes(0)) continue
      const source = content.toString('utf8')
      for (const [name, pattern] of forbidden) {
        if (pattern.test(source) && !allowedMatches.get(path)?.has(name)) failures.push(`${name} remains in ${path}`)
      }
    }
  }
  return failures
}

const readdirSafe = path => {
  try { return readdirSync(path) } catch { return null }
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const root = resolve(process.argv[2] ?? dirname(fileURLToPath(import.meta.url)), process.argv[2] ? '.' : '../..')
  const failures = checkCompatibility(root)
  if (failures.length) {
    console.error(`COMPATIBILITY_ZERO_FAIL\n${failures.join('\n')}`)
    process.exit(1)
  }
  console.log('COMPATIBILITY_ZERO_PASS')
}
