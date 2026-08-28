import assert from 'node:assert/strict'
import { chmodSync, existsSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { delimiter, join } from 'node:path'
import { spawnSync } from 'node:child_process'
import { test } from 'node:test'
import { fileURLToPath } from 'node:url'

const common = fileURLToPath(new URL('./common.sh', import.meta.url))

const executable = (path, body) => {
  writeFileSync(path, `#!/bin/sh\n${body}\n`)
  chmodSync(path, 0o755)
}

const probe = ({ withPnpm = false, withCorepack = true, functionName = 'run_pnpm', exitStatus = 0, argument = 'verify' }) => {
  const root = mkdtempSync(join(tmpdir(), 'go-admin-pnpm-contract-'))
  const output = join(root, 'output')
  if (withCorepack) executable(join(root, 'corepack'), `
test "$1" = enable
test "$2" = pnpm
test "$3" = --install-directory
mkdir -p "$4"
{
  printf '%s\\n' '#!/bin/sh'
  printf '%s\\n' 'if [ "$1" = outer ]; then exec pnpm nested; fi'
  printf '%s\\n' 'printf "corepack-shim:%s\\n" "$*" > "$GO_ADMIN_PNPM_PROBE"'
} > "$4/pnpm"
chmod +x "$4/pnpm"
`)
  if (withPnpm) {
    executable(join(root, 'pnpm'), `printf 'pnpm:%s\\n' "$*" > "$GO_ADMIN_PNPM_PROBE"\nexit ${exitStatus}`)
  }
  const result = spawnSync('/bin/sh', ['-c', `. "$1"; ${functionName} "$2"`, 'probe', common, argument], {
    env: {
      ...process.env,
      GO_ADMIN_ARTIFACTS_DIR: join(root, 'artifacts'),
      GO_ADMIN_PNPM_PROBE: output,
      PATH: [root, '/usr/bin', '/bin'].join(delimiter)
    },
    encoding: 'utf8'
  })
  return { result, output: existsSync(output) ? readFileSync(output, 'utf8').trim() : null }
}

test('uses Corepack when pnpm is unavailable to the managed shell', { skip: process.platform === 'win32' }, () => {
  const { result, output } = probe({})
  assert.equal(result.status, 0, result.stderr)
  assert.equal(output, 'corepack-shim:verify')
})

test('exports the Corepack shim for nested pnpm package scripts', { skip: process.platform === 'win32' }, () => {
  const { result, output } = probe({ argument: 'outer' })
  assert.equal(result.status, 0, result.stderr)
  assert.equal(output, 'corepack-shim:nested')
})

test('prefers an installed pnpm command over Corepack', { skip: process.platform === 'win32' }, () => {
  const { result, output } = probe({ withPnpm: true })
  assert.equal(result.status, 0, result.stderr)
  assert.equal(output, 'pnpm:verify')
})

test('exec_pnpm preserves the package manager exit status', { skip: process.platform === 'win32' }, () => {
  const { result, output } = probe({ withPnpm: true, functionName: 'exec_pnpm', exitStatus: 23 })
  assert.equal(result.status, 23, result.stderr)
  assert.equal(output, 'pnpm:verify')
})

test('fails deterministically when neither pnpm nor Corepack is installed', { skip: process.platform === 'win32' }, () => {
  const { result, output } = probe({ withCorepack: false })
  assert.equal(result.status, 1)
  assert.equal(output, null)
  assert.match(result.stderr, /required tool is not installed: pnpm or Corepack/)
})
