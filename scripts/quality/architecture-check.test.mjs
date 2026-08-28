import assert from 'node:assert/strict'
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'
import { checkArchitecture } from './architecture-check.mjs'

test('current repository satisfies canonical architecture', () => {
  assert.deepEqual(checkArchitecture(new URL('../..', import.meta.url).pathname), [])
})

test('rejects the historical short Go module path', () => {
  const root = mkdtempSync(join(tmpdir(), 'go-admin-architecture-'))
  mkdirSync(join(root, 'go-admin-plus'), { recursive: true })
  writeFileSync(join(root, 'go-admin-plus/go.mod'), 'module go-admin\n')

  assert.ok(checkArchitecture(root).includes(
    'Go module path must be github.com/NAMEWTA/go-admin-plus/go-admin-plus'
  ))
})

test('rejects the historical frontend workspace scope', () => {
  const root = mkdtempSync(join(tmpdir(), 'go-admin-architecture-'))
  mkdirSync(join(root, 'go-admin-plus-ui'), { recursive: true })
  writeFileSync(join(root, 'go-admin-plus-ui/package.json'), '{"name":"@go-admin/workspace"}\n')

  assert.ok(checkArchitecture(root).includes(
    'frontend workspace name must be @go-admin-plus/workspace'
  ))
})

test('rejects stale SpecDev verification commands', () => {
  const root = mkdtempSync(join(tmpdir(), 'go-admin-architecture-'))
  const configRoot = join(root, 'speculo/.speculo/specdev')
  const oldFrontend = ['go-admin-ui', 'plus'].join('-')
  mkdirSync(configRoot, { recursive: true })
  writeFileSync(join(configRoot, 'config.json'), JSON.stringify({
    verification: {
      test: `cd ${oldFrontend} && pnpm test:unit`,
      typecheck: `cd ${oldFrontend} && pnpm type-check`,
      lint: `cd ${oldFrontend} && pnpm lint`,
      build: `cd ${oldFrontend} && pnpm build:prod`
    }
  }))

  const failures = checkArchitecture(root)
  assert.ok(failures.includes('SpecDev verification.test must be task test'))
  assert.ok(failures.includes('SpecDev verification.typecheck must be pnpm --dir go-admin-plus-ui typecheck'))
  assert.ok(failures.includes('SpecDev verification.lint must be task lint'))
  assert.ok(failures.includes('SpecDev verification.build must be task build TARGET=all PROFILE=server-sqlite'))
})

test('rejects command filters for nonexistent workspace packages', () => {
  const root = mkdtempSync(join(tmpdir(), 'go-admin-architecture-'))
  mkdirSync(join(root, 'go-admin-plus-ui/apps/admin-web'), { recursive: true })
  writeFileSync(join(root, 'go-admin-plus-ui/package.json'), '{"name":"@go-admin-plus/workspace"}\n')
  writeFileSync(join(root, 'go-admin-plus-ui/apps/admin-web/package.json'), '{"name":"@go-admin-plus/admin-web"}\n')
  const unknownPackage = ['@go-admin-plus', 'missing'].join('/')
  writeFileSync(join(root, 'Taskfile.yml'), `pnpm --filter ${unknownPackage} build\n`)

  assert.ok(checkArchitecture(root).includes(
    `Taskfile.yml references unknown workspace package ${unknownPackage}`
  ))
})

test('rejects aggregate local frontend packaging', () => {
  const root = mkdtempSync(join(tmpdir(), 'go-admin-architecture-'))
  const scriptRoot = join(root, 'scripts/go-admin-plus-ui')
  mkdirSync(scriptRoot, { recursive: true })
  writeFileSync(join(scriptRoot, 'package.sh'), 'pnpm build:prod\n')

  assert.ok(checkArchitecture(root).includes(
    'local package script must not invoke the aggregate frontend build'
  ))
})
