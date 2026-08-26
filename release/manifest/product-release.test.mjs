import assert from 'node:assert/strict'
import { mkdtempSync, readFileSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
import { spawnSync } from 'node:child_process'
import test from 'node:test'

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '..', '..')
const SCRIPT = join(ROOT, 'release/manifest/product-release.mjs')
const sha = 'a'.repeat(40)
const digest = `sha256:${'b'.repeat(64)}`

const validManifest = () => ({
  schema_version: 1,
  product: {
    name: 'Go Admin Plus', version: '0.1.0', release_class: 'unsigned-self-use',
    external_distribution: false, production_deployment: false
  },
  provenance: {
    root_sha: sha, backend_sha: sha, frontend_sha: sha,
    openapi: { sha256: 'c'.repeat(64) }, migration: { max_version: '1786700004000' }
  },
  artifacts: Object.fromEntries([
    ['linux', 'linux/amd64', 'server-compose'],
    ['macos', 'darwin/arm64', 'desktop'],
    ['windows', 'windows/amd64', 'desktop']
  ].map(([key, platform, host]) => [key, {
    platform, host,
    release: { product_version: '0.1.0' },
    provenance: { head_sha: sha },
    artifact: { archive_sha256: digest },
    checksums: { files: ['SHA256SUMS'] },
    sbom: { files: ['artifact.spdx.json'] },
    signature: { type: 'none' }
  }])),
  policy: { global_security_disable: false, external_publish_authorized: false }
})

const verify = manifest => {
  const directory = mkdtempSync(join(tmpdir(), 'go-admin-product-manifest-'))
  const path = join(directory, 'manifest.json')
  writeFileSync(path, JSON.stringify(manifest))
  return spawnSync(process.execPath, [SCRIPT, 'verify', '--manifest', path], { encoding: 'utf8' })
}

test('accepts a complete unsigned self-use product manifest', () => {
  const result = verify(validManifest())
  assert.equal(result.status, 0, result.stderr)
  assert.match(result.stdout, /GO_ADMIN_PRODUCT_MANIFEST_VERIFY_PASS/)
})

test('rejects platform product version drift', () => {
  const manifest = validManifest()
  manifest.artifacts.windows.release.product_version = '0.2.0'
  const result = verify(manifest)
  assert.notEqual(result.status, 0)
  assert.match(result.stderr, /windows product version drifted/)
})

test('rejects an external publish authorization', () => {
  const manifest = validManifest()
  manifest.policy.external_publish_authorized = true
  const result = verify(manifest)
  assert.notEqual(result.status, 0)
  assert.match(result.stderr, /must not authorize external publish/)
})
