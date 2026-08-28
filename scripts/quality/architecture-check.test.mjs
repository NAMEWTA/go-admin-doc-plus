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
