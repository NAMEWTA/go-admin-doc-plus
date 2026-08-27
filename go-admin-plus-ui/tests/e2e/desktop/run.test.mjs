import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

test('native runner is a default skip with no environment prerequisites', () => {
  const result = spawnSync(process.execPath, [fileURLToPath(new URL('./run.mjs', import.meta.url))], {
    env: { PATH: process.env.PATH ?? '' }, encoding: 'utf8'
  })
  assert.equal(result.status, 0)
  assert.deepEqual(JSON.parse(result.stdout), {
    state: 'skipped',
    reason: 'GO_ADMIN_DESKTOP_NATIVE_E2E is not enabled'
  })
  assert.equal(result.stderr, '')
})
