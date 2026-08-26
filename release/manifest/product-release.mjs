#!/usr/bin/env node

import { createHash } from 'node:crypto'
import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { basename, dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { execFileSync } from 'node:child_process'

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '..', '..')
const PRODUCT_REPOSITORY = 'NAMEWTA/go-admin-plus'
const SHA_PATTERN = /^[0-9a-f]{40}$/
const DIGEST_PATTERN = /^sha256:[0-9a-f]{64}$/
const VERSION_PATTERN = /^\d+\.\d+\.\d+$/

const platforms = {
  linux: {
    platform: 'linux/amd64',
    host: 'server-compose',
    workflow: '.github/workflows/release-linux.yml',
    artifactPattern: /^linux-amd64-compose-(\d+\.\d+\.\d+)-(\d+)-(\d+)$/,
    checksums: ['SHA256SUMS'],
    sboms: ['sbom/go-admin-api.spdx.json', 'sbom/go-admin-web.spdx.json'],
    signature: { type: 'none', trust: 'checksum-and-source-provenance' }
  },
  macos: {
    platform: 'darwin/arm64',
    host: 'desktop',
    workflow: '.github/workflows/release-macos.yml',
    artifactPattern: /^macos-arm64-(\d+\.\d+\.\d+)-unsigned-self-use-(\d+)-(\d+)$/,
    checksums: ['SHA256SUMS'],
    sboms: ['go-admin-plus-macos-arm64.spdx.json'],
    signature: { type: 'adhoc', trust: 'unidentified-developer', notarization: 'not-applicable' }
  },
  windows: {
    platform: 'windows/amd64',
    host: 'desktop',
    workflow: '.github/workflows/release-windows.yml',
    artifactPattern: /^windows-amd64-(\d+\.\d+\.\d+)-unsigned-self-use-(\d+)-(\d+)$/,
    checksums: ['SHA256SUMS'],
    sboms: ['go-admin-plus-windows-amd64.spdx.json'],
    signature: { type: 'none', trust: 'unidentified-publisher' }
  }
}

const fail = message => {
  throw new Error(message)
}

const run = (command, args, cwd = ROOT) => execFileSync(command, args, {
  cwd,
  encoding: 'utf8',
  stdio: ['ignore', 'pipe', 'pipe']
}).trim()

const sha256File = path => createHash('sha256').update(readFileSync(path)).digest('hex')

const parseArgs = values => {
  const parsed = { _: [] }
  for (let index = 0; index < values.length; index += 1) {
    const value = values[index]
    if (!value.startsWith('--')) {
      parsed._.push(value)
      continue
    }
    const [rawName, inline] = value.slice(2).split('=', 2)
    if (inline !== undefined) parsed[rawName] = inline
    else if (values[index + 1] && !values[index + 1].startsWith('--')) parsed[rawName] = values[++index]
    else parsed[rawName] = true
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

const repositoryState = () => {
  const rootSha = exactSha(run('git', ['rev-parse', 'HEAD']), 'root SHA')
  return { rootSha, backendSha: rootSha, frontendSha: rootSha }
}

const migrationVersion = () => {
  const directory = join(ROOT, 'go-admin-plus/cmd/migrate/migration/version')
  const entries = run('find', [directory, '-maxdepth', '1', '-type', 'f', '-name', '[0-9]*_*.go']).split('\n')
  const versions = entries
    .filter(Boolean)
    .map(path => basename(path).match(/^(\d+)_/)?.[1])
    .filter(Boolean)
    .map(Number)
  if (versions.length === 0) fail('no numbered migration source files found')
  return String(Math.max(...versions))
}

const sourceContract = version => {
  if (!VERSION_PATTERN.test(version)) fail('version must use numeric major.minor.patch format')
  const identity = JSON.parse(readFileSync(join(ROOT, 'release/windows/identity.json'), 'utf8'))
  if (identity.product_version !== version) {
    fail(`version ${version} does not match Windows product identity ${identity.product_version}`)
  }
  const openapiPath = 'go-admin-plus/api/openapi/openapi.json'
  return {
    ...repositoryState(),
    openapi: {
      path: openapiPath,
      sha256: sha256File(join(ROOT, openapiPath))
    },
    migration: { max_version: migrationVersion() }
  }
}

const preflight = options => {
  const version = requireOption(options, 'version')
  const contract = sourceContract(version)
  if (options['root-ref'] && exactSha(options['root-ref'], 'root ref') !== contract.rootSha) {
    fail(`root ref ${options['root-ref']} does not match checkout ${contract.rootSha}`)
  }
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
  let response = await fetch(`https://api.github.com${path}`, { headers })
  if (!response.ok && token && (response.status === 403 || response.status === 404)) {
    delete headers.Authorization
    response = await fetch(`https://api.github.com${path}`, { headers })
  }
  if (!response.ok) fail(`GitHub API ${path} returned ${response.status}`)
  return response.json()
}

const collectPlatform = async (key, options, expectedBackendSha, version) => {
  const definition = platforms[key]
  const runId = numericId(requireOption(options, `${key}-run-id`), `${key} run ID`)
  const artifactId = numericId(requireOption(options, `${key}-artifact-id`), `${key} artifact ID`)
  const [workflowRun, artifact] = await Promise.all([
    githubJson(`/repos/${PRODUCT_REPOSITORY}/actions/runs/${runId}`),
    githubJson(`/repos/${PRODUCT_REPOSITORY}/actions/artifacts/${artifactId}`)
  ])
  if (workflowRun.status !== 'completed' || workflowRun.conclusion !== 'success') {
    fail(`${key} workflow run ${runId} is not a completed success`)
  }
  if (workflowRun.event !== 'workflow_dispatch' || workflowRun.path !== definition.workflow) {
    fail(`${key} run ${runId} is not ${definition.workflow} workflow_dispatch`)
  }
  if (exactSha(workflowRun.head_sha, `${key} run head SHA`) !== expectedBackendSha) {
    fail(`${key} run ${runId} used backend ${workflowRun.head_sha}, expected ${expectedBackendSha}`)
  }
  const expectedTitles = {
    linux: `linux product=${version} root=${options['root-sha']} frontend=${options['frontend-sha']}`,
    macos: `macos product=${version} root=${options['root-sha']} frontend=${options['frontend-sha']} mode=unsigned-self-use`,
    windows: `windows product=${version} root=${options['root-sha']} frontend=${options['frontend-sha']} mode=unsigned-self-use`
  }
  if (workflowRun.display_title !== expectedTitles[key]) {
    fail(`${key} run ${runId} provenance title does not bind the requested root/frontend/version`)
  }
  if (artifact.expired) fail(`${key} artifact ${artifactId} is expired`)
  if (artifact.workflow_run?.id !== runId) fail(`${key} artifact ${artifactId} does not belong to run ${runId}`)
  if (!DIGEST_PATTERN.test(artifact.digest ?? '')) fail(`${key} artifact ${artifactId} has no SHA-256 archive digest`)
  const nameMatch = artifact.name?.match(definition.artifactPattern)
  if (!nameMatch || nameMatch[1] !== version || Number(nameMatch[2]) !== runId || Number(nameMatch[3]) !== workflowRun.run_attempt) {
    fail(`${key} artifact name ${JSON.stringify(artifact.name)} does not bind run and attempt`)
  }
  return {
    platform: definition.platform,
    host: definition.host,
    release: {
      product_version: version,
      build_version: key === 'linux' ? expectedBackendSha : version,
      class: key === 'linux' ? 'offline-compose' : 'unsigned-self-use'
    },
    provenance: {
      repository: PRODUCT_REPOSITORY,
      workflow: definition.workflow,
      run_id: runId,
      run_attempt: workflowRun.run_attempt,
      head_sha: expectedBackendSha,
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
      expires_at: artifact.expires_at,
      api_url: artifact.url
    },
    checksums: { algorithm: 'SHA-256', files: definition.checksums },
    sbom: { format: 'SPDX JSON', files: definition.sboms },
    signature: definition.signature
  }
}

const validateManifest = manifest => {
  if (manifest.schema_version !== 1) fail('manifest schema_version must be 1')
  if (manifest.product?.name !== 'Go Admin Plus') fail('manifest product name is invalid')
  if (!VERSION_PATTERN.test(manifest.product?.version ?? '')) fail('manifest product version is invalid')
  if (manifest.product.release_class !== 'unsigned-self-use') fail('manifest release class must be unsigned-self-use')
  if (manifest.product.external_distribution !== false || manifest.product.production_deployment !== false) {
    fail('manifest must not authorize external distribution or production deployment')
  }
  for (const key of ['root_sha', 'backend_sha', 'frontend_sha']) {
    exactSha(manifest.provenance?.[key], `manifest provenance ${key}`)
  }
  if (!/^[0-9a-f]{64}$/.test(manifest.provenance?.openapi?.sha256 ?? '')) fail('manifest OpenAPI SHA-256 is invalid')
  if (!/^\d+$/.test(manifest.provenance?.migration?.max_version ?? '')) fail('manifest migration max version is invalid')
  for (const [key, definition] of Object.entries(platforms)) {
    const item = manifest.artifacts?.[key]
    if (!item || item.platform !== definition.platform || item.host !== definition.host) fail(`manifest ${key} platform contract is invalid`)
    if (item.release?.product_version !== manifest.product.version) fail(`manifest ${key} product version drifted`)
    if (item.provenance?.head_sha !== manifest.provenance.backend_sha) fail(`manifest ${key} backend provenance drifted`)
    if (!DIGEST_PATTERN.test(item.artifact?.archive_sha256 ?? '')) fail(`manifest ${key} artifact digest is invalid`)
    if (!Array.isArray(item.checksums?.files) || item.checksums.files.length === 0) fail(`manifest ${key} has no checksum contract`)
    if (!Array.isArray(item.sbom?.files) || item.sbom.files.length === 0) fail(`manifest ${key} has no SBOM contract`)
    if (!item.signature?.type) fail(`manifest ${key} has no explicit signature status`)
  }
  if (manifest.policy?.global_security_disable !== false) fail('manifest must prohibit global security disablement')
  if (manifest.policy?.external_publish_authorized !== false) fail('manifest must not authorize external publish')
  return manifest
}

const collect = async options => {
  const version = requireOption(options, 'version')
  const output = resolve(ROOT, requireOption(options, 'output'))
  const contract = sourceContract(version)
  options['root-sha'] = contract.rootSha
  options['frontend-sha'] = contract.frontendSha
  const artifacts = Object.fromEntries(await Promise.all(Object.keys(platforms).map(async key => [
    key,
    await collectPlatform(key, options, contract.backendSha, version)
  ])))
  const manifest = validateManifest({
    schema_version: 1,
    generated_at: new Date().toISOString(),
    product: {
      name: 'Go Admin Plus',
      version,
      release_class: 'unsigned-self-use',
      external_distribution: false,
      production_deployment: false
    },
    provenance: {
      root_sha: contract.rootSha,
      backend_sha: contract.backendSha,
      frontend_sha: contract.frontendSha,
      openapi: contract.openapi,
      migration: contract.migration
    },
    artifacts,
    policy: {
      intended_use: 'owner-authorized-self-use',
      global_security_disable: false,
      external_publish_authorized: false,
      macos_authorization: 'Privacy & Security Open Anyway or scoped quarantine removal after checksum verification',
      windows_authorization: 'interactive Run anyway where local policy permits after checksum verification'
    }
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
