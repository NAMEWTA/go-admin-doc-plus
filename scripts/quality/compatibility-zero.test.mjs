import assert from 'node:assert/strict'
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'
import { checkCompatibility } from './compatibility-zero.mjs'

test('detects removed paths and active compatibility references', () => {
  const root = mkdtempSync(join(tmpdir(), 'go-admin-compatibility-'))
  mkdirSync(join(root, 'go-admin-ui-plus'), { recursive: true })
  mkdirSync(join(root, 'docs'), { recursive: true })
  writeFileSync(join(root, 'README.md'), 'Use Redis and unsigned-self-use.\n')
  const failures = checkCompatibility(root)
  assert.ok(failures.some(message => message.includes('removed path')))
  assert.ok(failures.some(message => message.includes('Redis')))
  assert.ok(failures.some(message => message.includes('old release class')))
})
