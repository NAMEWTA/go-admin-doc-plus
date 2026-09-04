import assert from 'node:assert/strict'
import { mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { spawnSync } from 'node:child_process'
import test from 'node:test'

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '../..')
const SCRIPT = join(ROOT, 'release/manifest/product-release.mjs')
const sha = 'a'.repeat(40)
const digest = `sha256:${'b'.repeat(64)}`
const signatures = {
  linux: { type: 'digest-provenance', required: true },
  macos: { type: 'developer-id', required: true, notarization: 'apple-notary' },
  windows: { type: 'authenticode', required: true, timestamp: 'required' }
}

const validManifest = () => ({
  schema_version: 2,
  product: { name: 'Go Admin Plus', version: '0.0.1', release_class: 'production-candidate', publication_authorized: false },
  provenance: { source_sha: sha, openapi: { sha256: 'c'.repeat(64) }, migration: { max_version: '7500000000000' } },
  artifacts: Object.fromEntries([
    ['linux', ['linux/amd64', 'linux/arm64'], 'server-web', 'oci-compose'],
    ['macos', ['darwin/amd64', 'darwin/arm64'], 'desktop', 'signed-production'],
    ['windows', ['windows/amd64'], 'desktop', 'signed-production']
  ].map(([key, platforms, host, releaseClass]) => [key, {
    platforms, host,
    release: { product_version: '0.0.1', class: releaseClass },
    provenance: { head_sha: sha },
    artifact: { archive_sha256: digest },
    checksums: { files: ['SHA256SUMS'] },
    sbom: { files: ['artifact.spdx.json'] },
    signature: signatures[key]
  }])),
  policy: { protected_platform_gates_required: true, global_security_disable: false, publication_authorized: false }
})

const verify = manifest => {
  const directory = mkdtempSync(join(tmpdir(), 'go-admin-product-manifest-'))
  const path = join(directory, 'manifest.json')
  writeFileSync(path, JSON.stringify(manifest))
  return spawnSync(process.execPath, [SCRIPT, 'verify', '--manifest', path], { encoding: 'utf8' })
}

test('accepts complete protected production candidate evidence', () => {
  const result = verify(validManifest())
  assert.equal(result.status, 0, result.stderr)
  assert.match(result.stdout, /GO_ADMIN_PRODUCT_MANIFEST_VERIFY_PASS/)
})

test('rejects product version drift', () => {
  const manifest = validManifest()
  manifest.artifacts.windows.release.product_version = '0.2.0'
  const result = verify(manifest)
  assert.notEqual(result.status, 0)
  assert.match(result.stderr, /windows release identity drifted/)
})

test('rejects missing platform signing evidence', () => {
  const manifest = validManifest()
  manifest.artifacts.macos.signature.required = false
  const result = verify(manifest)
  assert.notEqual(result.status, 0)
  assert.match(result.stderr, /macos signature evidence is invalid/)
})

test('rejects publication authorization', () => {
  const manifest = validManifest()
  manifest.product.publication_authorized = true
  const result = verify(manifest)
  assert.notEqual(result.status, 0)
  assert.match(result.stderr, /must not authorize publication/)
})
