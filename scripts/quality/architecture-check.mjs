#!/usr/bin/env node

import { existsSync, readFileSync, readdirSync } from 'node:fs'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

export const checkArchitecture = root => {
  const failures = []
  const canonicalGoModule = 'github.com/NAMEWTA/go-admin-plus/go-admin-plus'
  const canonicalWorkspaceName = '@go-admin-plus/workspace'
  const required = [
    'Taskfile.yml', '.github/workflows/ci.yml', 'go-admin-plus/go.mod',
    'go-admin-plus/cmd/go-admin-plus/main.go', 'go-admin-plus/cmd/desktop-sidecar/main.go',
    'go-admin-plus/cmd/config-check/main.go', 'go-admin-plus/cmd/migrate/main.go',
    'go-admin-plus/internal/app/product/registry.go', 'go-admin-plus/internal/modules',
    'go-admin-plus-ui/package.json', 'go-admin-plus-ui/pnpm-workspace.yaml',
    'go-admin-plus-ui/apps/admin-web/package.json',
    'go-admin-plus-ui/apps/admin-desktop/src-tauri/tauri.conf.json'
  ]
  const forbidden = ['go-admin-ui-plus', 'go-admin-plus/app', 'go-admin-plus/common', 'go-admin-plus/api', 'go-admin-plus/cmd/go-admin-desktop']
  for (const path of required) if (!existsSync(join(root, path))) failures.push(`missing canonical path: ${path}`)
  for (const path of forbidden) if (existsSync(join(root, path))) failures.push(`removed path still exists: ${path}`)

  const goModulePath = join(root, 'go-admin-plus/go.mod')
  if (existsSync(goModulePath)) {
    const declaration = readFileSync(goModulePath, 'utf8').match(/^module\s+(\S+)$/m)?.[1]
    if (declaration !== canonicalGoModule) failures.push(`Go module path must be ${canonicalGoModule}`)
  }

  const frontendManifestPath = join(root, 'go-admin-plus-ui/package.json')
  if (existsSync(frontendManifestPath)) {
    const name = JSON.parse(readFileSync(frontendManifestPath, 'utf8')).name
    if (name !== canonicalWorkspaceName) failures.push(`frontend workspace name must be ${canonicalWorkspaceName}`)
  }

  const internalRoot = join(root, 'go-admin-plus/internal')
  if (existsSync(internalRoot)) {
    const allowed = new Set(['app', 'application', 'contracts', 'host', 'modules', 'platform'])
    for (const entry of readdirSync(internalRoot, { withFileTypes: true })) {
      if (entry.isDirectory() && !allowed.has(entry.name)) failures.push(`backend layer is outside the canonical architecture: internal/${entry.name}`)
    }
  }
  const commandRoot = join(root, 'go-admin-plus/cmd')
  if (existsSync(commandRoot)) {
    const allowed = new Set(['config-check', 'desktop-sidecar', 'go-admin-plus', 'migrate'])
    for (const entry of readdirSync(commandRoot, { withFileTypes: true })) {
      if (entry.isDirectory() && !allowed.has(entry.name)) failures.push(`backend command is outside the canonical command plane: cmd/${entry.name}`)
    }
  }
  const workspacePath = join(root, 'go-admin-plus-ui/pnpm-workspace.yaml')
  if (existsSync(workspacePath)) {
    const workspace = readFileSync(workspacePath, 'utf8')
    for (const pattern of ['apps/*', 'packages/*', 'packages/adapters/*', 'packages/domains/*', 'packages/web-domains/*']) {
      if (!workspace.includes(pattern)) failures.push(`workspace does not declare ${pattern}`)
    }
  }
  return failures
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const root = resolve(process.argv[2] ?? dirname(fileURLToPath(import.meta.url)), process.argv[2] ? '.' : '../..')
  const failures = checkArchitecture(root)
  if (failures.length) {
    console.error(`ARCHITECTURE_CHECK_FAIL\n${failures.join('\n')}`)
    process.exit(1)
  }
  console.log('ARCHITECTURE_CHECK_PASS')
}
