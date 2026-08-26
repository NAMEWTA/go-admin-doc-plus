import assert from 'node:assert/strict'
import { existsSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync } from 'node:fs'
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
    assert.equal(existsSync(join(
      outputRoot,
      'go-admin-plus-ui/packages/domains/contract-fixture/src/generated/client.ts'
    )), false)
  } finally {
    rmSync(outputRoot, { recursive: true, force: true })
  }
})

test('regenerates a legal nested owner path without manifest drift', () => {
  const outputRoot = mkdtempSync(join(tmpdir(), 'contract-state-output-'))
  try {
    const contracts = [fixture('valid-nested-module.yaml')]
    synchronizeGeneration(outputRoot, contracts)
    assert.doesNotThrow(() => synchronizeGeneration(outputRoot, contracts))
    assert.doesNotThrow(() => checkGeneration(outputRoot, contracts))
  } finally {
    rmSync(outputRoot, { recursive: true, force: true })
  }
})

test('rejects a manifest path outside the generated owner grammar without deleting it', () => {
  const outputRoot = mkdtempSync(join(tmpdir(), 'contract-state-output-'))
  try {
    synchronizeGeneration(outputRoot, [])
    const unmanagedRelative = 'go-admin-plus-ui/packages/domains/iam/manual/generated/client.ts'
    const unmanaged = join(outputRoot, unmanagedRelative)
    mkdirSync(dirname(unmanaged), { recursive: true })
    writeFileSync(unmanaged, 'keep me')

    const manifestPath = join(outputRoot, 'scripts/contracts/generated/manifest.json')
    const manifest = JSON.parse(readFileSync(manifestPath, 'utf8'))
    manifest.outputs.push(unmanagedRelative)
    writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`)

    assert.throws(() => synchronizeGeneration(outputRoot, []), /unmanaged output paths/)
    assert.equal(readFileSync(unmanaged, 'utf8'), 'keep me')
  } finally {
    rmSync(outputRoot, { recursive: true, force: true })
  }
})
