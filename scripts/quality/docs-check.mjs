#!/usr/bin/env node

import { existsSync, readFileSync, readdirSync } from 'node:fs'
import { dirname, extname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../..')
const documents = ['README.md', 'docs', 'deploy/README.md', 'database/README.md', 'release/README.md', 'go-admin-plus/README.md', 'go-admin-plus/config/README.md']
const walk = path => {
  if (!existsSync(path)) return []
  try {
    return readdirSync(path, { withFileTypes: true }).flatMap(entry => entry.isDirectory() ? walk(join(path, entry.name)) : [join(path, entry.name)])
  } catch { return [path] }
}
const failures = []
for (const item of documents.flatMap(path => walk(join(root, path))).filter(path => extname(path) === '.md')) {
  const source = readFileSync(item, 'utf8')
  for (const match of source.matchAll(/\[[^\]]+\]\(([^)]+)\)/g)) {
    const target = match[1].split('#', 1)[0]
    if (!target || /^(?:https?:|mailto:)/.test(target)) continue
    if (!existsSync(resolve(dirname(item), decodeURIComponent(target)))) failures.push(`broken documentation link: ${item.slice(root.length + 1)} -> ${target}`)
  }
}
if (failures.length) {
  console.error(`DOCS_CHECK_FAIL\n${failures.join('\n')}`)
  process.exit(1)
}
console.log('DOCS_CHECK_PASS')
