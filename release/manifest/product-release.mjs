#!/usr/bin/env node

import { createHash } from 'node:crypto'
import { execFileSync } from 'node:child_process'
import { existsSync, readFileSync, readdirSync, writeFileSync } from 'node:fs'
import { basename, dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '../..')
const PRODUCT_REPOSITORY = 'NAMEWTA/go-admin-plus'
const SHA_PATTERN = /^[0-9a-f]{40}$/
const DIGEST_PATTERN = /^sha256:[0-9a-f]{64}$/
const VERSION_PATTERN = /^\d+\.\d+\.\d+$/

const platforms = {
  linux: {
    workflow: '.github/workflows/release-linux.yml',
    platforms: ['linux/amd64', 'linux/arm64'],
    host: 'server-web',
    releaseClass: 'oci-compose',
    checksums: ['SHA256SUMS'],
    sboms: [
      'sbom/go-admin-plus-server-linux-amd64.spdx.json',
      'sbom/go-admin-plus-server-linux-arm64.spdx.json',
      'sbom/go-admin-plus-web-linux-amd64.spdx.json',
      'sbom/go-admin-plus-web-linux-arm64.spdx.json'
    ],
    signature: { type: 'digest-provenance', required: true }
  },
  macos: {
    workflow: '.github/workflows/release-macos.yml',
    platforms: ['darwin/amd64', 'darwin/arm64'],
    host: 'desktop',
    releaseClass: 'signed-production',
    checksums: ['SHA256SUMS'],
    sboms: ['go-admin-plus-macos-universal.spdx.json'],
    signature: { type: 'developer-id', required: true, notarization: 'apple-notary' }
  },
  windows: {
    workflow: '.github/workflows/release-windows.yml',
    platforms: ['windows/amd64'],
    host: 'desktop',
    releaseClass: 'signed-production',
    checksums: ['SHA256SUMS'],
    sboms: ['go-admin-plus-windows-amd64.spdx.json'],
    signature: { type: 'authenticode', required: true, timestamp: 'required' }
  }
}

const fail = message => { throw new Error(message) }
const run = (command, args) => execFileSync(command, args, { cwd: ROOT, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }).trim()
const sha256File = path => createHash('sha256').update(readFileSync(path)).digest('hex')

const parseArgs = values => {
  const parsed = {}
  for (let index = 0; index < values.length; index += 1) {
    const value = values[index]
    if (!value.startsWith('--')) fail(`unexpected argument: ${value}`)
    const [name, inline] = value.slice(2).split('=', 2)
    if (inline !== undefined) parsed[name] = inline
    else if (values[index + 1] && !values[index + 1].startsWith('--')) parsed[name] = values[++index]
    else parsed[name] = true
  }
  return parsed
}

const requireOption = (options, name) => {
  const value = options[name]
  if (typeof value !== 'string' || value.length === 0) fail(`--${name} is required`)
  return value
}
const exactSha = (value, name) => {
  const normalized = String(value).toLowerCase()
  if (!SHA_PATTERN.test(normalized)) fail(`${name} must be an exact lowercase 40-character SHA`)
  return normalized
}
const numericId = (value, name) => {
  if (!/^\d+$/.test(String(value)) || Number(value) <= 0) fail(`${name} must be a positive numeric GitHub ID`)
  return Number(value)
}

const walk = directory => readdirSync(directory, { withFileTypes: true }).flatMap(entry => {
  const path = join(directory, entry.name)
  return entry.isDirectory() ? walk(path) : [path]
})

const migrationVersion = () => {
  const versions = walk(join(ROOT, 'go-admin-plus/internal'))
    .filter(path => path.endsWith('.sql'))
    .map(path => basename(path).match(/^(\d+)_/)?.[1])
    .filter(Boolean)
    .map(Number)
  if (versions.length === 0) fail('no numbered migration SQL files found')
  return String(Math.max(...versions))
}

const sourceContract = version => {
  if (!VERSION_PATTERN.test(version)) fail('version must use numeric major.minor.patch format')
  const tauri = JSON.parse(readFileSync(join(ROOT, 'go-admin-plus-ui/apps/admin-desktop/src-tauri/tauri.conf.json'), 'utf8'))
  const macos = JSON.parse(readFileSync(join(ROOT, 'release/macos/identity.json'), 'utf8'))
  const windows = JSON.parse(readFileSync(join(ROOT, 'release/windows/identity.json'), 'utf8'))
  const linux = JSON.parse(readFileSync(join(ROOT, 'release/linux/identity.json'), 'utf8'))
  if (tauri.version !== version) fail(`version ${version} does not match Tauri product version ${tauri.version}`)
  if (macos.releaseClass !== 'signed-production' || !macos.signingRequired || !macos.notarizationRequired) fail('macOS production identity is incomplete')
  if (windows.releaseClass !== 'signed-production' || !windows.signingRequired) fail('Windows production identity is incomplete')
  if (JSON.stringify(linux.platforms) !== JSON.stringify(platforms.linux.platforms)) fail('Linux platform identity is incomplete')
  const rootSha = exactSha(run('git', ['rev-parse', 'HEAD']), 'root SHA')
  const openapiPath = 'scripts/contracts/generated/openapi.json'
  return {
    rootSha,
    openapi: { path: openapiPath, sha256: sha256File(join(ROOT, openapiPath)) },
    migration: { max_version: migrationVersion() }
  }
}

const preflight = options => {
  const version = requireOption(options, 'version')
  const contract = sourceContract(version)
  if (options['root-ref'] && exactSha(options['root-ref'], 'root ref') !== contract.rootSha) fail('root ref does not match checkout')
  if (!options['allow-dirty']) {
    const dirty = run('git', ['status', '--porcelain', '--untracked-files=normal'])
    if (dirty) fail(`root workspace is dirty:\n${dirty}`)
  }
  process.stdout.write(`${JSON.stringify({ version, ...contract }, null, 2)}\n`)
}

const githubJson = async path => {
  const headers = {
    Accept: 'application/vnd.github+json',
    'X-GitHub-Api-Version': '2022-11-28',
    'User-Agent': 'go-admin-plus-product-release-contract'
  }
  const token = process.env.GH_TOKEN || process.env.GITHUB_TOKEN
  if (token) headers.Authorization = `Bearer ${token}`
  const response = await fetch(`https://api.github.com${path}`, { headers })
  if (!response.ok) fail(`GitHub API ${path} returned ${response.status}`)
  return response.json()
}

const collectPlatform = async (key, options, sourceSha, version) => {
  const definition = platforms[key]
  const runId = numericId(requireOption(options, `${key}-run-id`), `${key} run ID`)
  const artifactId = numericId(requireOption(options, `${key}-artifact-id`), `${key} artifact ID`)
  const [workflowRun, artifact] = await Promise.all([
    githubJson(`/repos/${PRODUCT_REPOSITORY}/actions/runs/${runId}`),
    githubJson(`/repos/${PRODUCT_REPOSITORY}/actions/artifacts/${artifactId}`)
  ])
  if (workflowRun.status !== 'completed' || workflowRun.conclusion !== 'success') fail(`${key} workflow run is not a completed success`)
  if (workflowRun.event !== 'workflow_dispatch' || workflowRun.path !== definition.workflow) fail(`${key} workflow identity is invalid`)
  if (exactSha(workflowRun.head_sha, `${key} head SHA`) !== sourceSha) fail(`${key} source SHA drifted`)
  if (workflowRun.display_title !== `${key} source=${sourceSha} version=${version}`) fail(`${key} run title does not bind source and version`)
  if (artifact.expired || artifact.workflow_run?.id !== runId) fail(`${key} artifact does not belong to the active run`)
  if (artifact.name !== `go-admin-plus-${key}-${version}-${sourceSha}`) fail(`${key} artifact name does not bind source and version`)
  if (!DIGEST_PATTERN.test(artifact.digest ?? '')) fail(`${key} artifact has no SHA-256 archive digest`)
  return {
    platforms: definition.platforms,
    host: definition.host,
    release: { product_version: version, class: definition.releaseClass },
    provenance: {
      repository: PRODUCT_REPOSITORY,
      workflow: definition.workflow,
      run_id: runId,
      run_attempt: workflowRun.run_attempt,
      head_sha: sourceSha,
      event: workflowRun.event,
      conclusion: workflowRun.conclusion,
      url: workflowRun.html_url
    },
    artifact: {
      id: artifactId,
      name: artifact.name,
      archive_sha256: artifact.digest,
      size_bytes: artifact.size_in_bytes,
      created_at: artifact.created_at,
      expires_at: artifact.expires_at
    },
    checksums: { algorithm: 'SHA-256', files: definition.checksums },
    sbom: { format: 'SPDX JSON', files: definition.sboms },
    signature: definition.signature
  }
}

const validateManifest = manifest => {
  if (manifest.schema_version !== 2) fail('manifest schema_version must be 2')
  if (manifest.product?.name !== 'Go Admin Plus') fail('manifest product name is invalid')
  if (!VERSION_PATTERN.test(manifest.product?.version ?? '')) fail('manifest product version is invalid')
  if (manifest.product.release_class !== 'production-candidate') fail('manifest release class is invalid')
  if (manifest.product.publication_authorized !== false) fail('manifest must not authorize publication')
  exactSha(manifest.provenance?.source_sha, 'manifest source SHA')
  if (!/^[0-9a-f]{64}$/.test(manifest.provenance?.openapi?.sha256 ?? '')) fail('manifest OpenAPI digest is invalid')
  if (!/^\d+$/.test(manifest.provenance?.migration?.max_version ?? '')) fail('manifest migration version is invalid')
  for (const [key, definition] of Object.entries(platforms)) {
    const item = manifest.artifacts?.[key]
    if (!item || JSON.stringify(item.platforms) !== JSON.stringify(definition.platforms) || item.host !== definition.host) fail(`${key} platform contract is invalid`)
    if (item.release?.product_version !== manifest.product.version || item.release?.class !== definition.releaseClass) fail(`${key} release identity drifted`)
    if (item.provenance?.head_sha !== manifest.provenance.source_sha) fail(`${key} source provenance drifted`)
    if (!DIGEST_PATTERN.test(item.artifact?.archive_sha256 ?? '')) fail(`${key} artifact digest is invalid`)
    if (!Array.isArray(item.checksums?.files) || item.checksums.files.length === 0) fail(`${key} checksums are missing`)
    if (!Array.isArray(item.sbom?.files) || item.sbom.files.length === 0) fail(`${key} SBOM is missing`)
    if (JSON.stringify(item.signature) !== JSON.stringify(definition.signature)) fail(`${key} signature evidence is invalid`)
  }
  return manifest
}

const collect = async options => {
  const version = requireOption(options, 'version')
  const output = resolve(ROOT, requireOption(options, 'output'))
  const contract = sourceContract(version)
  const artifacts = Object.fromEntries(await Promise.all(Object.keys(platforms).map(async key => [
    key, await collectPlatform(key, options, contract.rootSha, version)
  ])))
  const manifest = validateManifest({
    schema_version: 2,
    generated_at: new Date().toISOString(),
    product: { name: 'Go Admin Plus', version, release_class: 'production-candidate', publication_authorized: false },
    provenance: { source_sha: contract.rootSha, openapi: contract.openapi, migration: contract.migration },
    artifacts,
    policy: { protected_platform_gates_required: true, global_security_disable: false, publication_authorized: false }
  })
  writeFileSync(output, `${JSON.stringify(manifest, null, 2)}\n`)
  process.stdout.write(`GO_ADMIN_PRODUCT_MANIFEST_PASS output=${output}\n`)
}

const verify = options => {
  const path = resolve(ROOT, requireOption(options, 'manifest'))
  if (!existsSync(path)) fail(`manifest does not exist: ${path}`)
  validateManifest(JSON.parse(readFileSync(path, 'utf8')))
  process.stdout.write(`GO_ADMIN_PRODUCT_MANIFEST_VERIFY_PASS manifest=${path}\n`)
}

const main = async () => {
  const [command, ...values] = process.argv.slice(2)
  const options = parseArgs(values)
  if (command === 'preflight') return preflight(options)
  if (command === 'collect') return collect(options)
  if (command === 'verify') return verify(options)
  fail('usage: product-release.mjs <preflight|collect|verify> [options]')
}

main().catch(error => {
  console.error(`GO_ADMIN_PRODUCT_MANIFEST_FAIL ${error.message}`)
  process.exitCode = 1
})
