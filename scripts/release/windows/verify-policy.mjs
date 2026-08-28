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
  assert.equal(identity.architecture, 'x86_64')
  assert.equal(identity.targetTriple, 'x86_64-pc-windows-msvc')
  assert.equal(identity.releaseClass, 'signed-production')
  assert.equal(identity.signingRequired, true)
  assert.equal(identity.signingProvider, 'azure-artifact-signing')
  assert.equal(identity.signerIdentity, 'protected-exact-thumbprint')
  assert.equal(identity.installerType, 'nsis')
  assert.equal(identity.installMode, 'currentUser')
  assert.equal(identity.webviewInstallMode, 'offlineInstaller')
  assert.deepEqual(identity.uninstallPolicy, {
    preserveAppData: true,
    preserveCredential: true,
    removeInstallDirectory: true
  })
  assert.equal(identity.remotePublish, false)
  assert.deepEqual(identity.evidence, ['SHA256SUMS', 'SPDX JSON', 'provenance.json', 'install-evidence.json'])
  assert.deepEqual(Object.keys(identity.generatorRuntime.archives).sort(), [
    'gitWindowsX64', 'goWindowsAmd64', 'nodeWindowsX64', 'pnpm'
  ])
  for (const archive of Object.values(identity.generatorRuntime.archives)) assert.match(archive.sha256, sha256)
}

export function verifyWorkflow(workflow) {
  assert.match(workflow, /environment:\s*windows-production/)
  assert.match(workflow, /source_ref:/)
  assert.match(workflow, /git rev-parse HEAD/)
  assert.match(workflow, /artifact-signing-cli --version 0\.11\.0 --locked/)
  assert.match(workflow, /tauri-driver --version 2\.0\.6 --locked/)
  assert.match(workflow, /AZURE_ARTIFACT_SIGNING_ENDPOINT/)
  assert.match(workflow, /AZURE_ARTIFACT_SIGNING_EXPECTED_THUMBPRINT/)
  assert.match(workflow, /build-nsis\.ps1/)
  assert.match(workflow, /verify-install\.ps1/)
  assert.match(workflow, /GO_ADMIN_WINDOWS_SUPPLY_CHAIN_PASS/)
  assert.doesNotMatch(workflow, /wails|unsigned-self-use|NotSigned|go-admin-ui-plus/i)
  assert.doesNotMatch(workflow, /gh\s+release|actions\/create-release|softprops\/action-gh-release/)
}

export async function verifyRepository(repository) {
  const read = relative => readFile(path.join(repository, relative), 'utf8')
  const [identityText, tauriText, runtime, preparer, builder, signer, verifier, installer, tracer, artifacts, workflow] = await Promise.all([
    read('release/windows/identity.json'),
    read('go-admin-plus-ui/apps/admin-desktop/src-tauri/tauri.conf.json'),
    read('go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs'),
    read('scripts/release/windows/prepare-generator-runtime.ps1'),
    read('scripts/release/windows/build-nsis.ps1'),
    read('scripts/release/windows/sign.cmd'),
    read('scripts/release/windows/verify-artifacts.ps1'),
    read('scripts/release/windows/verify-install.ps1'),
    read('scripts/release/windows/trace-installed.mjs'),
    read('scripts/release/windows/emit-artifacts.ps1'),
    read('.github/workflows/release-windows.yml')
  ])
  const identity = JSON.parse(identityText)
  const tauri = JSON.parse(tauriText)
  verifyIdentity(identity)
  assert.equal(tauri.identifier, identity.bundleIdentifier)
  assert.match(runtime, /\("windows", "x86_64"\).*windows-amd64/)
  assert.match(runtime, /git\.exe/)
  assert.match(runtime, /SYSTEMROOT.*WINDIR.*COMSPEC.*TEMP.*TMP/)
  assert.match(preparer, /gitWindowsX64/)
  assert.match(preparer, /--frozen-lockfile/)
  assert.match(preparer, /--offline/)
  assert.match(preparer, /supportedArchitectures/)
  assert.match(preparer, /"win32"/)
  assert.match(preparer, /Get-FileHash/)
  assert.match(preparer, /GOPROXY = 'https:\/\/proxy\.golang\.org'/)
  assert.match(preparer, /GOSUMDB = 'sum\.golang\.org'/)
  assert.ok(preparer.indexOf("GOPROXY = 'https://proxy.golang.org'") < preparer.indexOf('mod download'))
  assert.match(builder, /x86_64-pc-windows-msvc|targetTriple/)
  assert.match(builder, /webviewInstallMode/)
  assert.match(builder, /offlineInstaller/)
  assert.match(builder, /installMode = 'currentUser'/)
  assert.match(builder, /signCommand/)
  assert.match(builder, /verify-production\.mjs/)
  assert.match(builder, /--files/)
  assert.match(signer, /artifact-signing-cli/)
  assert.match(verifier, /Get-AuthenticodeSignature/)
  assert.match(verifier, /TimeStamperCertificate/)
  assert.match(verifier, /AZURE_ARTIFACT_SIGNING_EXPECTED_THUMBPRINT/)
  assert.match(installer, /GITHUB_ACTIONS/)
  assert.match(installer, /appDataPreserved/)
  assert.match(installer, /installDirectoryRemoved/)
  assert.match(installer, /TimeStamperCertificate/)
  assert.match(installer, /WebView2/)
  assert.match(installer, /msedgedriver\.microsoft\.com\/\$webViewVersion/)
  assert.match(installer, /SignerCertificate\.Subject -notmatch 'Microsoft'/)
  assert.match(installer, /trace-installed\.mjs/)
  assert.match(installer, /go-admin-plus\.db/)
  assert.doesNotMatch(installer, /go-admin\.sqlite3/)
  assert.match(tracer, /GO_ADMIN_WINDOWS_INSTALLED_TRACER_PASS/)
  assert.match(tracer, /firstLaunchLogin/)
  assert.match(tracer, /persistence/)
  assert.match(artifacts, /spdx-json=/)
  assert.match(artifacts, /remotePublished = \$false/)
  verifyWorkflow(workflow)
  for (const legacy of [
    'release/windows/prepare-project.ps1', 'release/windows/verify-installer.ps1',
    'release/windows/verify-self-use-install.ps1', 'release/windows/verify-payload.ps1'
  ]) await assert.rejects(read(legacy))
}

const current = fileURLToPath(import.meta.url)
if (process.argv[1] && path.resolve(process.argv[1]) === current) {
  const repository = path.resolve(path.dirname(current), '../../..')
  await verifyRepository(repository)
  console.log('GO_ADMIN_WINDOWS_RELEASE_POLICY_PASS')
}
