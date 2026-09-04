import assert from 'node:assert/strict'
import { test } from 'node:test'
import { parseGoTestEvents, runRequiredPostgres, validateRequiredEnvironment } from './required-postgres.mjs'

const environment = { GO_ADMIN_CI_REQUIRE_POSTGRES: '1', GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN: 'postgres://ci:secret@127.0.0.1:5432/disposable?sslmode=disable', GITHUB_RUN_ID: '42', GITHUB_RUN_ATTEMPT: '1', TEMP: '/tmp' }
const event = (Action, Test = 'TestRequired') => JSON.stringify({ Action, Package: 'example', Test })

test('parses only the exact top-level required test', () => {
  assert.deepEqual(parseGoTestEvents([event('run'), event('run', 'TestRequired/subtest'), event('pass', 'TestRequired/subtest'), event('pass')].join('\n'), 'TestRequired'), { run: 1, pass: 1, fail: 0, skip: 0 })
})

test('requires the explicit CI flag and a database URL without exposing it', () => {
  assert.throws(() => validateRequiredEnvironment({ ...environment, GO_ADMIN_CI_REQUIRE_POSTGRES: '' }), /REQUIRE_POSTGRES/)
  assert.throws(() => validateRequiredEnvironment({ ...environment, GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN: '' }), /DSN is required/)
  assert.throws(() => validateRequiredEnvironment({ ...environment, GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN: 'sqlite:data.db' }), /identify a database/)
})

test('fails zero targets, skips, missing runs and process failures', () => {
  assert.throws(() => runRequiredPostgres({ root: '/repo', environment, suites: [] }), /zero targets/)
  const suite = [{ name: 'required', packagePath: './required', test: 'TestRequired' }]
  for (const [output, status, expected] of [
    [[event('run'), event('skip')].join('\n'), 0, /skip=1/],
    ['', 0, /run=0/],
    [[event('run'), event('fail')].join('\n'), 1, /status=1/]
  ]) assert.throws(() => runRequiredPostgres({ root: '/repo', environment, suites: suite, prepare: () => {}, spawn: () => ({ status, stdout: output }) }), expected)
})

test('reports one pass and injects the required PostgreSQL environment names', () => {
  let child
  const report = runRequiredPostgres({ root: '/repo', environment, suites: [{ name: 'required', packagePath: './required', test: 'TestRequired' }], prepare: () => {}, spawn: (_command, _args, options) => { child = options.env; return { status: 0, stdout: [event('run'), event('pass')].join('\n') } } })
  assert.deepEqual(report, [{ name: 'required', test: 'TestRequired', schema: 'ci_01_required_42_1', executed: 1, passed: 1, skipped: 0 }])
  for (const name of ['GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN', 'GO_ADMIN_TEST_POSTGRES_FILES_LIFECYCLE_DSN', 'GO_ADMIN_TEST_POSTGRES_IAM_DELETION_DSN', 'GO_ADMIN_SCHEDULER_POSTGRES_DSN']) {
    assert.equal(new URL(child[name]).searchParams.get('search_path'), 'ci_01_required_42_1')
  }
})
