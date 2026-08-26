import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { test } from 'node:test'
import { fileURLToPath } from 'node:url'
import { browserDiagnostic, runnerFailureLine, safeRunnerDiagnostic } from './diagnostics.mjs'

test('extracts a bounded browser assertion without replacing a string diagnostic', () => {
  const output = '<pre id="result">IAM_ADMIN_E2E_FAIL|ASSERTION|missing tab: roles</pre>'
  assert.equal(browserDiagnostic(output), 'missing tab: roles')
  assert.equal(browserDiagnostic('<pre>IAM_ADMIN_E2E_FAIL</pre>'), 'safe browser diagnostic unavailable')
  assert.equal(safeRunnerDiagnostic('x'.repeat(300)).length, 200)
})

test('redacts sensitive material and never includes an Error stack', () => {
  const error = new Error('request https://example.test/private password=hunter2 token=abc dsn=postgres://admin:secret@example.test/db')
  error.stack = 'STACK MUST NOT LEAK'
  const line = runnerFailureLine(error)
  assert.match(line, /^IAM_ADMIN_E2E_RUN_FAIL\|/)
  assert.match(line, /\[redacted-url\]/)
  assert.match(line, /password=\[redacted\]/)
  assert.match(line, /token=\[redacted\]/)
  assert.match(line, /dsn=\[redacted\]/)
  for (const forbidden of ['hunter2', 'abc', 'admin:secret', 'STACK MUST NOT LEAK', 'example.test']) assert.doesNotMatch(line, new RegExp(forbidden))
})

test('runner lifecycle remains an honest no-op without required opt-in', () => {
  const runner = fileURLToPath(new URL('./run.mjs', import.meta.url))
  const result = spawnSync(process.execPath, [runner], { encoding: 'utf8', env: { PATH: process.env.PATH ?? '' }, timeout: 10_000 })
  assert.equal(result.status, 0)
  assert.equal(result.stderr, '')
  assert.match(result.stdout, /^IAM_ADMIN_E2E_SKIP required opt-in is disabled\s*$/)
})
