#!/usr/bin/env node

import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const sha256 = /^[0-9a-f]{64}$/

export function verifyIdentity(identity) {
  assert.equal(identity.schemaVersion, 1)
  assert.equal(identity.identityStatus, 'approved')
  assert.equal(identity.bundleIdentifier, 'com.goadmin.plus')
  assert.deepEqual(identity.architectures, ['x86_64', 'arm64'])
  assert.equal(identity.releaseClass, 'signed-production')
  assert.equal(identity.signingRequired, true)
  assert.equal(identity.notarizationRequired, true)
  assert.equal(identity.hardenedRuntimeRequired, true)
  assert.equal(identity.remotePublish, false)
  assert.deepEqual(identity.evidence, ['SHA256SUMS', 'SPDX JSON', 'provenance.json', 'notary.json'])
  assert.match(identity.generatorRuntime.goVersion, /^\d+\.\d+\.\d+$/)
  assert.match(identity.generatorRuntime.nodeVersion, /^\d+\.\d+\.\d+$/)
  assert.match(identity.generatorRuntime.pnpmVersion, /^\d+\.\d+\.\d+$/)
  assert.deepEqual(Object.keys(identity.generatorRuntime.archives).sort(), [
    'goDarwinAmd64', 'goDarwinArm64', 'nodeDarwinArm64', 'nodeDarwinX64', 'pnpm'
  ])
  for (const archive of Object.values(identity.generatorRuntime.archives)) {
    assert.match(archive.name, /\.(?:tar\.gz|tgz)$/)
    assert.match(archive.sha256, sha256)
  }
}

export function verifyWorkflow(workflow) {
  assert.match(workflow, /environment:\s*macos-production/)
  assert.match(workflow, /source_ref:/)
  assert.match(workflow, /git rev-parse HEAD/)
  assert.match(workflow, /build-universal\.sh/)
  assert.match(workflow, /build\.mjs --target aarch64-apple-darwin[\s\S]*cargo test/)
  assert.match(workflow, /APPLE_SIGNING_IDENTITY/)
  assert.match(workflow, /APPLE_NOTARY_KEY_P8_BASE64/)
  assert.match(workflow, /sign-app\.sh/)
  assert.match(workflow, /notarize\.sh app/)
  assert.match(workflow, /notarize\.sh dmg/)
  assert.match(workflow, /verify-install\.sh/)
  assert.match(workflow, /GO_ADMIN_MACOS_SUPPLY_CHAIN_PASS/)
  assert.doesNotMatch(workflow, /wails/i)
  assert.doesNotMatch(workflow, /unsigned-self-use|ad-hoc|xattr\s+-dr/)
  assert.doesNotMatch(workflow, /gh\s+release|actions\/create-release|softprops\/action-gh-release/)
}

export async function verifyRepository(repository) {
  const read = relative => readFile(path.join(repository, relative), 'utf8')
  const [identityText, tauriConfigText, entitlements, runtime, builder, preparer, signer, verifier, installer, artifacts, workflow] = await Promise.all([
    read('release/macos/identity.json'),
    read('go-admin-plus-ui/apps/admin-desktop/src-tauri/tauri.conf.json'),
    read('release/macos/entitlements.plist'),
    read('go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs'),
    read('scripts/release/macos/build-universal.sh'),
    read('scripts/release/macos/prepare-generator-runtime.sh'),
    read('scripts/release/macos/sign-app.sh'),
    read('scripts/release/macos/verify-app.sh'),
    read('scripts/release/macos/verify-install.sh'),
    read('scripts/release/macos/emit-artifacts.sh'),
    read('.github/workflows/release-macos.yml')
  ])
  const identity = JSON.parse(identityText)
  const tauri = JSON.parse(tauriConfigText)
  verifyIdentity(identity)
  assert.equal(tauri.identifier, identity.bundleIdentifier)
  assert.deepEqual(entitlements.match(/<key>/g) ?? [], [])
  assert.match(runtime, /\.env_clear\(\)/)
  assert.match(runtime, /DEVELOPMENT_ENVIRONMENT_ALLOWLIST/)
  assert.match(runtime, /Contents|resource_dir/)
  assert.match(runtime, /darwin-arm64/)
  assert.match(runtime, /darwin-amd64/)
  assert.match(builder, /aarch64-apple-darwin x86_64-apple-darwin|aarch64-apple-darwin/)
  assert.match(builder, /universal-apple-darwin/)
  assert.match(builder, /--features custom-protocol/)
  assert.match(builder, /--no-sign/)
  assert.match(builder, /lipo -create/)
  assert.match(preparer, /git -C "\$generator\/repository" archive|git -C "\$repository" archive/)
  assert.match(preparer, /--frozen-lockfile/)
  assert.match(preparer, /--offline/)
  assert.match(preparer, /supportedArchitectures/)
  assert.match(preparer, /"darwin"/)
  assert.match(preparer, /"x64","arm64"/)
  assert.doesNotMatch(preparer, /fetch --frozen-lockfile --force/)
  assert.match(preparer, /shasum -a 256 -c/)
  assert.match(signer, /--options runtime/)
  assert.doesNotMatch(signer, /--deep --sign/)
  assert.match(verifier, /lipo -archs/)
  assert.match(verifier, /binding-darwin-arm64/)
  assert.match(verifier, /binding-darwin-x64/)
  assert.match(verifier, /verify-production\.mjs/)
  assert.match(verifier, /--files "\$host" "\$sidecar"/)
  assert.match(verifier, /codesign --verify --deep --strict/)
  assert.match(installer, /GITHUB_ACTIONS/)
  assert.match(installer, /security find-generic-password/)
  assert.match(installer, /security delete-generic-password/)
  assert.match(installer, /launch_and_stop restart/)
  assert.match(artifacts, /spdx-json=/)
  assert.match(artifacts, /remotePublished:false/)
  verifyWorkflow(workflow)
  for (const legacy of [
    'release/macos/package-dmg.sh',
    'release/macos/prepare-app.sh',
    'release/macos/verify-self-use-install.sh'
  ]) {
    await assert.rejects(read(legacy))
  }
}

const current = fileURLToPath(import.meta.url)
if (process.argv[1] && path.resolve(process.argv[1]) === current) {
  const repository = path.resolve(path.dirname(current), '../../..')
  await verifyRepository(repository)
  console.log('GO_ADMIN_MACOS_RELEASE_POLICY_PASS')
}
