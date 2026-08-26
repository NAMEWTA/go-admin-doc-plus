import { spawn, spawnSync } from 'node:child_process'
import { existsSync, mkdtempSync, readFileSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { browserDiagnostic, runnerFailureLine, safeRunnerDiagnostic } from './diagnostics.mjs'

const required = process.env.GO_ADMIN_REQUIRE_IAM_ADMIN_E2E === '1'
const chromium = process.env.GO_ADMIN_TEST_CHROMIUM_EXECUTABLE
const postgresKey = 'GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN'
const postgres = process.env[postgresKey]
if (!required) { console.log('IAM_ADMIN_E2E_SKIP required opt-in is disabled'); process.exit(0) }
if (!chromium || !postgres) throw new Error('IAM administration E2E requires Chromium and disposable PostgreSQL')
const root = dirname(fileURLToPath(import.meta.url)); const uiRoot = resolve(root, '../../../..'); const backend = resolve(uiRoot, '../go-admin-plus')
const temporary = mkdtempSync(join(tmpdir(), 'go-admin-iam-admin-e2e-')); const staticRoot = join(temporary, 'static')
const allowedKeys = ['HOME', 'PATH', 'TMPDIR', 'TEMP', 'LANG', 'LC_ALL', 'GOCACHE', 'GOMODCACHE', 'GOPATH', 'GOROOT', 'GOPROXY', 'GOSUMDB', 'GOENV', 'GOTOOLCHAIN', 'PNPM_HOME', 'COREPACK_HOME']
const environment = (extra = {}, includePostgres = false) => { const result = {}; for (const key of allowedKeys) if (process.env[key] !== undefined) result[key] = process.env[key]; Object.assign(result, extra); if (includePostgres) result[postgresKey] = postgres; else delete result[postgresKey]; return result }
const checked = (command, args, options) => { const result = spawnSync(command, args, { ...options, encoding: 'utf8', timeout: 120_000, killSignal: 'SIGKILL' }); if (result.status !== 0 || result.error || result.signal) throw new Error(`${command} failed`); return result.stdout ?? '' }
const childErrors = new WeakMap()
const track = (child) => { child.once('error', (error) => childErrors.set(child, error)); return child }
const waitReady = async (path, child) => { const deadline = Date.now()+60_000; while (Date.now()<deadline) { if (childErrors.has(child)) throw new Error('HTTPS host could not start'); if (existsSync(path)) return readFileSync(path, 'utf8').trim(); if (child.exitCode !== null) throw new Error('HTTPS host exited'); await new Promise((resolvePromise) => setTimeout(resolvePromise, 100)) } throw new Error('HTTPS host readiness timed out') }
const waitForExit = (child, timeout) => new Promise((resolvePromise) => { if (child.exitCode !== null) { resolvePromise(true); return } const timer = setTimeout(() => { child.removeListener('exit', exited); resolvePromise(false) }, timeout); const exited = () => { clearTimeout(timer); resolvePromise(true) }; child.once('exit', exited) })
const stop = async (child) => {
  if (!child || child.exitCode !== null) return
  child.kill('SIGTERM')
  if (await waitForExit(child, 5_000)) return
  child.kill('SIGKILL')
  if (!await waitForExit(child, 5_000)) throw new Error('HTTPS host cleanup timed out')
}

let failure = ''
try {
  checked('pnpm', ['exec', 'vite', 'build', '--config', join(root, 'vite.config.ts')], { cwd: uiRoot, env: environment({ GO_ADMIN_IAM_ADMIN_E2E_OUT_DIR: staticRoot }), stdio: 'pipe' })
  for (const profile of ['sqlite', 'postgres']) {
    const ready = join(temporary, `${profile}.ready`)
    const host = track(spawn('go', ['test', './test/iam/authorization', '-run', '^TestIAMAdministrationBrowserHarnessServer$', '-count=1', '-v'], { cwd: backend, env: environment({ GO_ADMIN_IAM_ADMIN_E2E_SERVE: '1', GO_ADMIN_IAM_ADMIN_E2E_PROFILE: profile, GO_ADMIN_IAM_ADMIN_E2E_READY_FILE: ready, GO_ADMIN_IAM_ADMIN_E2E_STATIC_DIR: staticRoot }, profile === 'postgres'), stdio: ['ignore', 'ignore', 'ignore'] }))
    try { const url = await waitReady(ready, host); const output = checked(chromium, ['--headless=new', '--disable-gpu', '--ignore-certificate-errors', '--no-first-run', '--no-default-browser-check', '--dump-dom', '--virtual-time-budget=45000', url], { env: environment(), stdio: 'pipe' }); if (!output.includes('IAM_ADMIN_E2E_PASS')) throw new Error(`${profile} browser scenario failed: ${browserDiagnostic(output)}`) } finally { await stop(host) }
  }
} catch (error) {
  failure = safeRunnerDiagnostic(error)
} finally {
  try { rmSync(temporary, { recursive: true, force: true }) } catch { if (!failure) failure = 'temporary cleanup failed' }
}
if (failure) { console.error(runnerFailureLine(failure)); process.exitCode = 1 } else console.log('IAM_ADMIN_E2E_PASS')
