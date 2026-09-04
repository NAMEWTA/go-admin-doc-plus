import assert from 'node:assert/strict'
import { mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import test from 'node:test'

import { main, suites, validateEnvironment, verifySuiteResult } from './required-web.mjs'

const fixtureEnvironment = () => {
  const root = mkdtempSync(join(tmpdir(), 'required-web-contract-'))
  const chromium = join(root, 'chromium')
  writeFileSync(chromium, '')
  return {
    GO_ADMIN_REQUIRE_WEB_E2E: '1',
    GO_ADMIN_TEST_CHROMIUM_EXECUTABLE: chromium,
    GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN: 'postgres://localhost/disposable',
  }
}

test('required environment rejects opt-out and missing real providers', () => {
  assert.match(validateEnvironment({}) ?? '', /REQUIRE_WEB_E2E/)
  assert.match(validateEnvironment({ GO_ADMIN_REQUIRE_WEB_E2E: '1' }) ?? '', /Chromium/)
})

test('suite verification rejects skip, failure, and missing execution marker', () => {
  const suite = suites[0]
  assert.match(verifySuiteResult(suite, { status: 0, stdout: 'WEB_SHELL_E2E_SKIP' }) ?? '', /skip/)
  assert.match(verifySuiteResult(suite, { status: 1, stdout: suite.marker }) ?? '', /child failed/)
  assert.match(verifySuiteResult(suite, { status: 0, stdout: 'unrelated output' }) ?? '', /marker/)
  assert.equal(verifySuiteResult(suite, { status: 0, stdout: suite.marker }), null)
})

test('runner requires every suite to execute with an exact pass marker', () => {
  let calls = 0
  const status = main(fixtureEnvironment(), (_command, _arguments, options) => {
    const suite = suites[calls++]
    assert.equal(options.env[suite.require], '1')
    return { status: 0, stdout: suite.marker, stderr: '' }
  })
  assert.equal(status, 0)
  assert.equal(calls, suites.length)
})

test('runner stops at the first non-green suite', () => {
  let calls = 0
  const status = main(fixtureEnvironment(), () => {
    const suite = suites[calls++]
    return calls === 3 ? { status: 0, stdout: `${suite.marker}\nANY_E2E_SKIP` } : { status: 0, stdout: suite.marker }
  })
  assert.equal(status, 1)
  assert.equal(calls, 3)
})
