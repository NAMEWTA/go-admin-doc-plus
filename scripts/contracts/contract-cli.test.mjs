import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const repositoryRoot = join(dirname(fileURLToPath(import.meta.url)), '..', '..')
const cli = join(repositoryRoot, 'scripts', 'contracts', 'cli.mjs')
const fixture = name => join(repositoryRoot, 'scripts', 'contracts', 'fixtures', name)

const run = (...args) => spawnSync(process.execPath, [cli, ...args], {
  cwd: repositoryRoot,
  encoding: 'utf8'
})

test('lints the canonical OpenAPI 3.1 contract', () => {
  const result = run('lint')
  assert.equal(result.status, 0, result.stderr || result.stdout)
  assert.match(result.stdout, /CONTRACT_LINT_PASS/)
})

test('checks deterministic Go and TypeScript transport generation', () => {
  const result = run('generate', '--check')
  assert.equal(result.status, 0, result.stderr || result.stdout)
  assert.match(result.stdout, /CONTRACT_GENERATE_CHECK_PASS/)
})

for (const [name, expected] of [
  ['invalid-schema.yaml', /type|schema/i],
  ['duplicate-operation-id.yaml', /operationId/i],
  ['leaked-internal-detail.yaml', /sensitive|internal detail/i]
]) {
  test(`rejects ${name}`, () => {
    const result = run('lint', '--contract', fixture(name))
    assert.notEqual(result.status, 0, 'invalid contract unexpectedly passed')
    assert.match(`${result.stdout}\n${result.stderr}`, expected)
  })
}
