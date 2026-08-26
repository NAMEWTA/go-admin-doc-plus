import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { readFileSync } from 'node:fs'
import { test } from 'node:test'
import { fileURLToPath } from 'node:url'
import { administrationMountDiagnostic, browserDiagnostic, runnerFailureLine, safeRunnerDiagnostic } from './diagnostics.mjs'

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

test('mount timeout exposes only stable controller state and known alert codes', () => {
  assert.equal(administrationMountDiagnostic({ failure: 'forbidden', canUsersRead: false, rows: 0, total: 0, loading: false, alertText: 'You do not have permission for that action.' }),
    'administration page did not load failure=forbidden can-users-read=false rows=0 total=0 loading=false alert=forbidden')
  const unknown = administrationMountDiagnostic({ failure: 'unexpected-secret', canUsersRead: true, rows: 2, total: 9, loading: true, alertText: 'password=hunter2 https://private.example' })
  assert.equal(unknown, 'administration page did not load failure=none can-users-read=true rows=2 total=9 loading=true alert=unrecognized')
  assert.doesNotMatch(unknown, /hunter2|private\.example|password/)
})

test('administration permission branches subscribe to the page revision', () => {
  const page = readFileSync(new URL('../../../../packages/web-domains/iam/src/administration/AdministrationPage.vue', import.meta.url), 'utf8')
  const template = page.split('<template>')[1] ?? ''

  assert.match(page, /const can = \(permissionCode: string\) => \{ void revision\.value; return props\.controller\.can\(permissionCode\) \}/)
  assert.doesNotMatch(template, /controller\.can\(/)
  assert.match(template, /\bcan\(/)
})

test('browser administration interactions wait for Vue DOM updates', () => {
  const driver = readFileSync(new URL('./browser-driver.ts', import.meta.url), 'utf8')
  const openTab = driver.match(/const openTab = async[\s\S]*?\n}\n/)?.[0] ?? ''

  assert.match(openTab, /aria-pressed/)
  assert.match(openTab, /section\[aria-labelledby=/)
  assert.doesNotMatch(driver, /(?<!await )openTab\('/)
  assert.doesNotMatch(driver, /clickRow\([^\n]+, 'edit'\)/)
  assert.match(driver, /clickRow\(key, action\)\n  await nextTick\(\)\n  await waitUntil/)
  assert.match(driver, /await waitUntil\(\(\) => !element<HTMLButtonElement>\('\[data-testid="delete-selected-users"\]'\)\.disabled/)
})
