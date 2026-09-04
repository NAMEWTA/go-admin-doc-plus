#!/usr/bin/env node

import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import path from 'node:path'

const shaReference = /@sha256:[0-9a-f]{64}/

export function verifyComposeText(compose) {
  assert.match(compose, /profiles:\s*\[postgres\]/)
  assert.match(compose, /profiles:\s*\[sqlite\]/)
  assert.match(compose, /api-postgres:/)
  assert.match(compose, /api-sqlite:/)
  assert.match(compose, /read_only:\s*true/)
  assert.match(compose, /cap_drop:\s*\[ALL\]/)
  assert.match(compose, /internal:\s*true/)
  assert.match(compose, shaReference)
  assert.doesNotMatch(compose, /privileged:\s*true/)
  assert.doesNotMatch(compose, /network_mode:\s*host/)
  assert.doesNotMatch(compose, /platform:\s*linux\/amd64/)
  assert.doesNotMatch(compose, /go-admin-ui-plus/)
}

export function verifyContainerfile(text) {
  const references = [...text.matchAll(/(?:FROM|ARG\s+\w+=)([^\s]+)/g)].map(match => match[1])
  assert.ok(references.some(reference => shaReference.test(reference)))
  assert.match(text, /USER (?:10001:10001|101:101)/)
  assert.doesNotMatch(text, /--mount=type=secret[^\n]*required=false/)
}

export function verifyIdentity(identity) {
  assert.equal(identity.schemaVersion, 1)
  assert.equal(identity.product, 'go-admin-plus')
  assert.deepEqual(identity.artifacts, ['go-admin-plus-server'])
  assert.deepEqual(identity.platforms, ['linux/amd64', 'linux/arm64'])
  assert.deepEqual(identity.profiles, ['server-postgres', 'server-sqlite'])
  assert.equal(identity.remotePublish, false)
  assert.deepEqual(identity.evidence, ['SHA256SUMS', 'SPDX JSON', 'provenance.json'])
}

export async function verifyRepository(repository) {
  const read = relative => readFile(path.join(repository, relative), 'utf8')
  const [compose, build, failure, server, web, workflow, imageBuild, artifacts, identityText, postgresConfig, sqliteConfig, service, serviceScript, install] = await Promise.all([
    read('deploy/compose/compose.yml'),
    read('deploy/compose/compose.build.yml'),
    read('deploy/compose/compose.migration-failure.yml'),
    read('release/linux/Containerfile.server'),
    read('release/linux/Containerfile.web'),
    read('.github/workflows/release.yml'),
    read('scripts/release/linux/build-images.sh'),
    read('scripts/release/linux/emit-artifacts.sh'),
    read('release/linux/identity.json'),
    read('deploy/compose/config/server-postgres.json'),
    read('deploy/compose/config/server-sqlite.json'),
    read('release/linux/go-admin-plus-server.service'),
    read('scripts/release/linux/build-service.sh'),
    read('release/linux/SERVER-INSTALL.md')
  ])
  verifyComposeText(compose)
  verifyContainerfile(server)
  verifyContainerfile(web)
  assert.match(server, /go build -trimpath -buildvcs=false/)
  assert.doesNotMatch(server, /pnpm --dir go-admin-plus-ui install/)
  assert.doesNotMatch(server, /git init --quiet/)
  assert.doesNotMatch(server, /\/opt\/go-admin-plus\/repository/)
  assert.match(build, /release\/linux\/Containerfile\.server/)
  assert.match(build, /release\/linux\/Containerfile\.web/)
  assert.match(failure, /--profile=server-postgres/)
  assert.match(failure, /--profile=server-sqlite/)
  assert.doesNotMatch(failure, /repository-root|missing-release-skeleton/)
  assert.match(workflow, /build-service\.sh/)
  assert.match(workflow, /gh release create/)
  assert.doesNotMatch(workflow, /docker\s+(?:login|push)|release-linux\.yml/i)
  assert.match(service, /ExecStart=.*go-admin-plus-server serve --profile server-sqlite/)
  assert.match(serviceScript, /GOOS=linux GOARCH=/)
  assert.match(serviceScript, /go-admin-plus-server-postgres\.service/)
  assert.match(install, /systemctl enable --now/)
  verifyIdentity(JSON.parse(identityText))
  const postgres = JSON.parse(postgresConfig)
  const sqlite = JSON.parse(sqliteConfig)
  assert.equal(postgres.profile, 'server-postgres')
  assert.equal(sqlite.profile, 'server-sqlite')
  assert.equal(Object.hasOwn(postgres, 'database'), false)
  assert.equal(sqlite.database.path, '/var/lib/go-admin-plus/database.sqlite3')
  for (const text of [postgresConfig, sqliteConfig]) {
    assert.doesNotMatch(text, /"(?:dsn|password|token|secret)"\s*:/i)
  }
}

const current = fileURLToPath(import.meta.url)
if (process.argv[1] && path.resolve(process.argv[1]) === current) {
  const repository = path.resolve(path.dirname(current), '../../..')
  await verifyRepository(repository)
  console.log('GO_ADMIN_LINUX_RELEASE_POLICY_PASS')
}
