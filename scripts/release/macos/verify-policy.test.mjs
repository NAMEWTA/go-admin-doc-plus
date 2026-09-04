import assert from 'node:assert/strict'
import test from 'node:test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { verifyIdentity, verifyRepository, verifyWorkflow } from './verify-policy.mjs'

const repository = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../..')

test('repository macOS release policy is complete', async () => {
  await verifyRepository(repository)
})

test('identity rejects a single-architecture or unsigned release', () => {
  assert.throws(() => verifyIdentity({
    schemaVersion: 1,
    identityStatus: 'approved',
    bundleIdentifier: 'com.goadmin.plus',
    architectures: ['arm64'],
    releaseClass: 'development',
    signingRequired: false,
    notarizationRequired: false,
    hardenedRuntimeRequired: false,
    remotePublish: false,
    evidence: []
  }))
})

test('workflow rejects self-use and remote release regression', () => {
  const invalid = `
environment: macos-production
source_ref:
git rev-parse HEAD
universal-apple-darwin
APPLE_SIGNING_IDENTITY
APPLE_NOTARY_KEY_P8_BASE64
sign-app.sh
notarize.sh app
notarize.sh dmg
verify-install.sh
GO_ADMIN_MACOS_SUPPLY_CHAIN_PASS
unsigned-self-use
gh release create
`
  assert.throws(() => verifyWorkflow(invalid))
})
