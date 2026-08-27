import assert from 'node:assert/strict'
import test from 'node:test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { verifyComposeText, verifyIdentity, verifyRepository } from './verify-policy.mjs'

const repository = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../..')

test('repository Linux release policy is complete', async () => {
  await verifyRepository(repository)
})

test('policy rejects a privileged or single-architecture Compose regression', () => {
  const invalid = 'profiles: [postgres]\nprofiles: [sqlite]\napi-postgres:\napi-sqlite:\nread_only: true\ncap_drop: [ALL]\ninternal: true\nimage: base@sha256:' + 'a'.repeat(64) + '\nprivileged: true\nplatform: linux/amd64\n'
  assert.throws(() => verifyComposeText(invalid))
})

test('identity requires both architectures and forbids remote publication', () => {
  assert.throws(() => verifyIdentity({
    schemaVersion: 1,
    platforms: ['linux/amd64'],
    profiles: ['server-postgres', 'server-sqlite'],
    remotePublish: true,
    evidence: ['SHA256SUMS', 'SPDX JSON', 'provenance.json']
  }))
})
