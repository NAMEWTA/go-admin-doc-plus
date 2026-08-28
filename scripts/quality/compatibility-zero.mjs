#!/usr/bin/env node

import { existsSync, readFileSync, readdirSync } from 'node:fs'
import { dirname, extname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const removedPaths = [
  'go-admin-ui-plus', 'go-admin-plus/app', 'go-admin-plus/common', 'go-admin-plus/api',
  'go-admin-plus/internal/tenant', 'go-admin-plus/internal/profile', 'go-admin-plus/cmd/go-admin-desktop'
]
const forbidden = [
  ['old frontend name', /go-admin-ui-plus/], ['Wails runtime', /\bwails(?:app)?\b/i],
  ['old upstream core', /go-admin-core/], ['Casbin', /\bcasbin\b/i],
  ['Redis', /\bredis\b/i], ['tenant feature', /\btenant(?:s|_id)?\b/i],
  ['MySQL', /\bmysql\b/i], ['SQL Server', /\bsqlserver\b|\bsql server\b/i],
  ['old SQLite build tag', /(?:-tags[= ]+|build:?)sqlite3\b/i],
  ['old release class', /unsigned-self-use/i]
]
const textExtensions = new Set(['', '.go', '.json', '.md', '.mjs', '.sh', '.ts', '.yaml', '.yml'])

const walk = directory => {
  if (!existsSync(directory)) return []
  return readdirSync(directory, { withFileTypes: true }).flatMap(entry => {
    if (['node_modules', 'dist', 'target', '.git'].includes(entry.name)) return []
    const path = join(directory, entry.name)
    return entry.isDirectory() ? walk(path) : [path]
  })
}

export const checkCompatibility = root => {
  const failures = []
  for (const path of removedPaths) if (existsSync(join(root, path))) failures.push(`removed path still exists: ${path}`)
  const scanRoots = [
    'Taskfile.yml', '.github', '.agents/skills', 'scripts/go-admin-plus', 'scripts/go-admin-plus-ui',
    'release/manifest', 'README.md', 'docs', 'deploy/README.md', 'database/README.md',
    'release/README.md', 'go-admin-plus/README.md', 'go-admin-plus/config/README.md', 'go-admin-plus/go.mod'
  ]
  const ownFiles = new Set(['scripts/quality/compatibility-zero.mjs', 'scripts/quality/compatibility-zero.test.mjs'])
  for (const rootPath of scanRoots) {
    const absolute = join(root, rootPath)
    const files = existsSync(absolute) && !readdirSafe(absolute) ? [absolute] : walk(absolute)
    for (const file of files) {
      const path = relative(root, file)
      if (ownFiles.has(path) || !textExtensions.has(extname(file))) continue
      const source = readFileSync(file, 'utf8')
      for (const [name, pattern] of forbidden) if (pattern.test(source)) failures.push(`${name} remains in ${path}`)
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
