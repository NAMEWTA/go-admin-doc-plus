import assert from 'node:assert/strict'
import { mkdtempSync, readFileSync, readdirSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'
import { checkGeneration, synchronizeGeneration } from './cli.mjs'

const directory = dirname(fileURLToPath(import.meta.url))
const fixture = name => join(directory, 'fixtures', name)
const contractTemporaryDirectories = () => readdirSync(tmpdir())
  .filter(name => /^go-admin-(?:contract|generate|module|oapi)/.test(name))
  .sort()

test('failed generation leaves existing outputs and temporary resources unchanged', () => {
  const outputRoot = mkdtempSync(join(tmpdir(), 'contract-state-output-'))
  try {
    synchronizeGeneration(outputRoot, [fixture('valid-module.yaml')])
    const manifest = join(outputRoot, 'scripts', 'contracts', 'generated', 'manifest.json')
    const beforeManifest = readFileSync(manifest, 'utf8')
    const beforeTemporaryDirectories = contractTemporaryDirectories()

    assert.throws(
      () => synchronizeGeneration(outputRoot, [fixture('invalid-module-owner.yaml')]),
      /owner|output/i
    )

    assert.equal(readFileSync(manifest, 'utf8'), beforeManifest)
    assert.deepEqual(contractTemporaryDirectories(), beforeTemporaryDirectories)
  } finally {
    rmSync(outputRoot, { recursive: true, force: true })
  }
})

test('detects and contracts outputs left by a removed module fragment', () => {
  const outputRoot = mkdtempSync(join(tmpdir(), 'contract-state-output-'))
  try {
    synchronizeGeneration(outputRoot, [fixture('valid-module.yaml')])
    assert.throws(() => checkGeneration(outputRoot, []), /manifest\.json|drift/i)

    synchronizeGeneration(outputRoot, [])
    assert.doesNotThrow(() => checkGeneration(outputRoot, []))
  } finally {
    rmSync(outputRoot, { recursive: true, force: true })
  }
})
