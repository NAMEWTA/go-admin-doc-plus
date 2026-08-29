#!/usr/bin/env node

import { mkdir, mkdtemp, readFile, readdir, realpath, rm } from 'node:fs/promises'
import { spawn } from 'node:child_process'
import { createHash, randomBytes } from 'node:crypto'
import { createConnection } from 'node:net'
import { networkInterfaces, tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { clickButtonScript, fillAndClickScript, windowContainsScript, windowValueScript } from './accessibility.mjs'
import { nativeAccessibilityFailure, nativePhaseFailure } from './diagnostics.mjs'
import { execute, reapNewSidecars, sidecarProcesses } from './processes.mjs'
import { verifyDesktopProductionAssets, verifyDesktopProductionFiles } from '../../../apps/admin-desktop/scripts/verify-production.mjs'

const enabled = 'GO_ADMIN_DESKTOP_NATIVE_E2E'
const maxOutput = 16 * 1024
const fixturePassword = 'administrator password'
const productionKeyringService = 'com.goadmin.plus.stronghold'
const productionKeyringAccount = 'desktop-session-vault'
const testKeyringService = 'com.goadmin.plus.stronghold.native-e2e'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '../../../..')
const goRoot = join(root, 'go-admin-plus')
const uiRoot = join(root, 'go-admin-plus-ui')
const appRoot = join(uiRoot, 'apps/admin-desktop')
const rustRoot = join(appRoot, 'src-tauri')
const sidecarBinary = join(rustRoot, 'binaries/go-admin-sidecar-aarch64-apple-darwin')
const hostBinary = join(rustRoot, 'target/release/go-admin-plus-desktop')

if (process.env[enabled] !== '1') {
  process.stdout.write(`${JSON.stringify({ state: 'skipped', reason: `${enabled} is not enabled` })}\n`)
  process.exit(0)
}
if (process.platform !== 'darwin') {
  process.stderr.write('desktop native E2E requires the macOS native runner\n')
  process.exit(1)
}

const delay = milliseconds => new Promise(resolveDelay => setTimeout(resolveDelay, milliseconds))
const safeText = value => value
  .replaceAll(fixturePassword, '[redacted]')
  .replaceAll(/(?:csrf|token|cookie|password)[^\s]*/gi, '[redacted]')
const qualifyCommandFailure = (error, phase) => error instanceof Error && error.message.startsWith('desktop native command ')
  ? new Error(`desktop native ${phase} failed`)
  : error instanceof Error && error.message.startsWith('desktop native accessibility ')
    ? new Error(`desktop native ${phase} failed: ${error.message}`)
  : error
const keyringExists = (service, account) => new Promise((resolveKeyring, rejectKeyring) => {
  const child = spawn('/usr/bin/security', ['find-generic-password', '-s', service, '-a', account], { stdio: 'ignore' })
  child.once('error', () => rejectKeyring(new Error('macOS credential store unavailable')))
  child.once('exit', code => {
    if (code === 0) resolveKeyring(true)
    else if (code === 44) resolveKeyring(false)
    else rejectKeyring(new Error('macOS credential store probe failed'))
  })
})

const deleteTestKeyring = async account => {
  if (await keyringExists(testKeyringService, account)) {
    await execute('/usr/bin/security', ['delete-generic-password', '-s', testKeyringService, '-a', account])
  }
  if (await keyringExists(testKeyringService, account)) throw new Error('native test credential cleanup failed')
}

const assertNoNewSidecars = async baseline => {
  const current = await sidecarProcesses()
  const leaked = [...current].filter(pid => !baseline.has(pid))
  if (leaked.length !== 0) throw new Error('desktop sidecar process was not reaped')
}

const newSidecarPid = async baseline => {
  const current = await sidecarProcesses()
  const added = [...current].filter(pid => !baseline.has(pid))
  if (added.length !== 1) throw new Error('native host did not own exactly one sidecar process')
  return added[0]
}

const assertLoopbackOnly = async pid => {
  const output = await execute('/usr/sbin/lsof', ['-Pan', '-p', String(pid), '-iTCP', '-sTCP:LISTEN'])
  const match = output.match(/TCP 127\.0\.0\.1:(\d+) \(LISTEN\)/)
  if (!match || /TCP (?:\*|0\.0\.0\.0|\[::\]):/.test(output)) throw new Error('desktop sidecar listener was not loopback-only')
  const address = Object.values(networkInterfaces()).flat().find(value => value?.family === 'IPv4' && !value.internal)?.address
  if (!address) return
  await new Promise((resolveProbe, rejectProbe) => {
    const socket = createConnection({ host: address, port: Number(match[1]) })
    const timer = setTimeout(() => { socket.destroy(); resolveProbe() }, 1000)
    socket.once('connect', () => { clearTimeout(timer); socket.destroy(); rejectProbe(new Error('desktop sidecar accepted a LAN connection')) })
    socket.once('error', () => { clearTimeout(timer); resolveProbe() })
  })
}

const hashFile = async path => createHash('sha256').update(await readFile(path)).digest('hex')

const restoreProductionArtifacts = async () => {
  await Promise.all([
    rm(sidecarBinary, { force: true }),
    rm(hostBinary, { force: true }),
    rm(join(appRoot, 'dist'), { recursive: true, force: true })
  ])
  await execute(process.execPath, [join(root, 'release/shared/sidecar/build.mjs'), '--target', 'aarch64-apple-darwin'], { cwd: root })
  await execute(join(appRoot, 'node_modules/.bin/vite'), ['build', '--config', 'vite.config.ts'], {
    cwd: appRoot, env: { PATH: process.env.PATH ?? '', HOME: process.env.HOME ?? '' }
  })
  await execute('cargo', ['build', '--locked', '--quiet', '--release', '--features', 'custom-protocol'], { cwd: rustRoot, timeout: 300_000 })
  await verifyDesktopProductionAssets(join(appRoot, 'dist'))
  await verifyDesktopProductionFiles([sidecarBinary, hostBinary])
}

const assertSafeDiagnostics = (output, protectedRoots) => {
  const lower = output.toLowerCase()
  const forbidden = ['__host-go-admin-session', 'csrf', 'cookie', 'readiness', 'controltoken', 'bearer ', 'session token']
  if (output.includes(fixturePassword) || protectedRoots.some(root => output.includes(root)) || forbidden.some(value => lower.includes(value))) {
    throw new Error('native diagnostics leaked protected material')
  }
}

const runAppleScript = script => new Promise((resolveScript, rejectScript) => {
  const child = spawn('/usr/bin/osascript', ['-'], { stdio: ['pipe', 'pipe', 'pipe'] })
  const stdout = []
  const stderr = []
  let size = 0
  let settled = false
  const finish = (error, output = '') => {
    if (settled) return
    settled = true
    clearTimeout(timer)
    if (error) rejectScript(error)
    else resolveScript(output)
  }
  const collect = target => chunk => {
    size += chunk.length
    if (size > 4096) {
      child.kill('SIGKILL')
      finish(new Error('desktop native accessibility output exceeded the limit'))
      return
    }
    target.push(chunk)
  }
  const timer = setTimeout(() => {
    child.kill('SIGKILL')
    finish(new Error('desktop native accessibility command timed out'))
  }, 20_000)
  child.stdout.on('data', collect(stdout))
  child.stderr.on('data', collect(stderr))
  child.once('error', () => finish(new Error('desktop native accessibility command unavailable')))
  child.once('close', (code, signal) => {
    if (settled) return
    if (code === 0 && signal === null) {
      finish(undefined, Buffer.concat(stdout).toString('utf8'))
      return
    }
    const failure = Buffer.concat(stderr).toString('utf8')
    const controlled = nativeAccessibilityFailure(failure)
    const codeMatch = failure.match(/\(-?\d+\)/)?.[0]
    finish(new Error(controlled ?? `desktop native accessibility command failed${codeMatch ? ` ${codeMatch}` : ''}`))
  })
  child.stdin.end(script)
})
const processIsAlive = pid => {
  try {
    process.kill(pid, 0)
    return true
  } catch (error) {
    if (error?.code === 'ESRCH') return false
    throw error
  }
}

const windowCount = async pid => {
  let output
  try {
    output = await runAppleScript(`tell application "System Events"
if exists (first process whose unix id is ${pid}) then
  tell (first process whose unix id is ${pid}) to return count of windows
end if
return 0
end tell`)
  } catch (error) {
    if (!processIsAlive(pid)) return 0
    throw error
  }
  return Number.parseInt(output.trim(), 10) || 0
}

const windowContains = async (pid, value) => {
  if (!processIsAlive(pid)) throw new Error('native host exited before UI observation')
  const output = await runAppleScript(windowContainsScript(pid, value))
  return output.trim() === 'true'
}

const poll = async (description, condition, timeout = 30_000) => {
  const deadline = Date.now() + timeout
  while (Date.now() < deadline) {
    if (await condition()) return
    await delay(100)
  }
  throw new Error(`${description} timed out`)
}

const pollBoundary = async (pid, timeout = 30_000) => {
  const deadline = Date.now() + timeout
  while (Date.now() < deadline) {
    if (await windowContains(pid, 'E2E authenticated boundary verified')) return
    const blocked = (await runAppleScript(windowValueScript(pid, 'E2E boundary blocked:'))).trim()
    if (blocked) throw new Error(blocked)
    await delay(100)
  }
  throw new Error('authenticated WebView storage and URL boundary timed out')
}

const pollControl = async (pid, description, success, timeout = 30_000) => {
  const deadline = Date.now() + timeout
  while (Date.now() < deadline) {
    if (await windowContains(pid, success)) return
    const failed = (await runAppleScript(windowValueScript(pid, 'E2E control failed:'))).trim()
    if (failed) throw new Error(failed)
    await delay(100)
  }
  throw new Error(`${description} timed out`)
}

const pollRestoredIdentity = async (pid, timeout = 90_000) => {
  const deadline = Date.now() + timeout
  while (Date.now() < deadline) {
    if (await windowContains(pid, 'Administrator')) return
    if (await windowContains(pid, '使用管理员账号登录控制台')) throw new Error('Stronghold session was not restored')
    if (await windowContains(pid, '服务暂不可用')) throw new Error('Stronghold identity restore was unavailable')
    await delay(100)
  }
  throw new Error('Stronghold authenticated workspace timed out')
}

const startApp = (binary, isolatedRoot, keyringAccount) => {
  const env = {}
  for (const key of ['HOME', 'PATH', 'TMPDIR', 'LANG', 'LC_ALL']) {
    if (process.env[key]) env[key] = process.env[key]
  }
  env.GO_ADMIN_DESKTOP_E2E_ROOT = isolatedRoot
  env.GO_ADMIN_DESKTOP_E2E_KEYRING_ACCOUNT = keyringAccount
  const child = spawn(binary, [], { env, stdio: ['ignore', 'pipe', 'pipe'] })
  let output = Buffer.alloc(0)
  const collect = chunk => {
    if (output.length >= maxOutput) return
    output = Buffer.concat([output, chunk]).subarray(0, maxOutput)
    if (output.length === maxOutput) child.kill('SIGKILL')
  }
  child.stdout.on('data', collect)
  child.stderr.on('data', collect)
  const exited = new Promise(resolveExit => {
    let settled = false
    const finish = result => {
      if (settled) return
      settled = true
      resolveExit(result)
    }
    child.once('exit', (code, signal) => finish({ code, signal, spawnError: false }))
    child.once('error', () => finish({ code: null, signal: null, spawnError: true }))
  })
  return { child, exited, output: () => output.toString('utf8') }
}

const stopApp = async owner => {
  if (owner.child.exitCode === null && owner.child.signalCode === null) owner.child.kill('SIGTERM')
  let result = await Promise.race([owner.exited, delay(8_000).then(() => null)])
  if (result === null) {
    owner.child.kill('SIGKILL')
    result = await Promise.race([owner.exited, delay(5_000).then(() => null)])
    if (result === null) throw new Error('native host cleanup failed')
  }
  if (owner.child.exitCode === null && owner.child.signalCode === null && !result.spawnError) throw new Error('native host cleanup was not observed')
}

const login = (pid, username, password) => runAppleScript(fillAndClickScript(pid, [
  { name: '账号', value: username },
  { name: '密码', value: password }
], '登录'))

const createProduct = pid => runAppleScript(fillAndClickScript(pid, [
  { name: 'SKU', value: 'E2E-001' },
  { name: '名称', value: 'Native product' },
  { name: '描述', role: 'AXTextArea', value: 'created through the native window' },
  { name: '价格（分）', value: '1250' }
], '保存'))

const updateProduct = async pid => {
  await runAppleScript(clickButtonScript(pid, '修改'))
  await runAppleScript(fillAndClickScript(pid, [
    { name: '名称', value: 'Native product updated' }
  ], '保存'))
}

const deleteProduct = pid => clickButton(pid, '删除 E2E-001')

const logout = pid => runAppleScript(clickButtonScript(pid, '退出登录'))

const clickButton = (pid, name) => runAppleScript(clickButtonScript(pid, name))

const main = async () => {
  const workspace = await realpath(await mkdtemp(join(tmpdir(), 'go-admin-desktop-native-')))
  const failedKeyring = `go-admin-plus-native-e2e-${randomBytes(16).toString('hex')}`
  const liveKeyring = `go-admin-plus-native-e2e-${randomBytes(16).toString('hex')}`
  const sidecarBaseline = await sidecarProcesses()
  const owners = new Set()
  const startTracked = (binary, isolatedRoot, keyringAccount) => {
    const owner = startApp(binary, isolatedRoot, keyringAccount)
    owners.add(owner)
    return owner
  }
  const stopTracked = async owner => {
    await stopApp(owner)
    owners.delete(owner)
  }
  let app
  let failure
  let phase = 'preflight'
  try {
    if (await keyringExists(productionKeyringService, productionKeyringAccount)) {
      throw new Error('production desktop credential pre-existed; native E2E refuses to touch it')
    }
    if (await keyringExists(testKeyringService, failedKeyring) || await keyringExists(testKeyringService, liveKeyring)) {
      throw new Error('native test credential identity collision')
    }
    phase = 'native-sidecar-build'
    await execute(process.execPath, [join(root, 'release/shared/sidecar/build.mjs'), '--native-e2e', '--target', 'aarch64-apple-darwin'], { cwd: root })
    phase = 'native-ui-build'
    await execute(join(appRoot, 'node_modules/.bin/vite'), ['build', '--config', 'vite.config.ts', '--mode', 'native-e2e'], {
      cwd: appRoot,
      env: { PATH: process.env.PATH ?? '', HOME: process.env.HOME ?? '' }
    })
    phase = 'native-host-build'
    await execute('cargo', ['build', '--locked', '--quiet', '--release', '--features', 'native-e2e'], { cwd: rustRoot, timeout: 600_000 })
    const binary = hostBinary

    phase = 'migration-failure-fixture'
    const failedRoot = join(workspace, 'failed')
    await mkdir(failedRoot, { recursive: true, mode: 0o700 })
    await execute('go', ['run', './test/desktop/fixture', '--root', failedRoot, '--mode', 'migration-failure'], { cwd: goRoot })
    const database = join(failedRoot, 'data/go-admin-plus.db')
    const beforeFailure = await hashFile(database)
    phase = 'migration-failure-host'
    const failed = startTracked(binary, failedRoot, failedKeyring)
    const failureDeadline = Date.now() + 15_000
    while (Date.now() < failureDeadline && failed.child.exitCode === null) {
      if (await windowCount(failed.child.pid) !== 0) throw new Error('migration failure opened the native window')
      await delay(100)
    }
    phase = 'migration-failure-verification'
    if (failed.child.exitCode === null) throw new Error('migration failure did not terminate the native host')
    const failedExit = await failed.exited
    if (failedExit.spawnError || failedExit.code === 0 || failedExit.code === null) throw new Error('migration failure did not produce a nonzero exit')
    await stopTracked(failed)
    if (await windowCount(failed.child.pid) !== 0) throw new Error('migration failure left a native window')
    if (await hashFile(database) !== beforeFailure) throw new Error('migration failure changed the source database fixture')
    const backups = await readdir(join(failedRoot, 'data/backups')).catch(error => error?.code === 'ENOENT' ? [] : Promise.reject(error))
    if (backups.length !== 1) throw new Error('migration failure did not preserve exactly one recovery snapshot')
    if (await hashFile(join(failedRoot, 'data/backups', backups[0], 'go-admin-plus.db')) !== beforeFailure) {
      throw new Error('migration failure backup does not match the original database')
    }
    assertSafeDiagnostics(failed.output(), [workspace, failedRoot])
    await deleteTestKeyring(failedKeyring)
    await assertNoNewSidecars(sidecarBaseline)

    phase = 'live-fixture'
    const liveRoot = join(workspace, 'live')
    await mkdir(liveRoot, { recursive: true, mode: 0o700 })
    await execute('go', ['run', './test/desktop/fixture', '--root', liveRoot, '--mode', 'previous'], { cwd: goRoot })
    phase = 'login-window'
    app = startTracked(binary, liveRoot, liveKeyring)
    await poll('native login window', () => windowContains(app.child.pid, '使用管理员账号登录控制台'), 90_000)
    phase = 'login-submit'
    await login(app.child.pid, 'admin', fixturePassword)
    phase = 'login-workspace'
    await poll('native authenticated workspace', () => windowContains(app.child.pid, 'Administrator'))
    phase = 'login-navigation'
    await poll('native Demo navigation', () => windowContains(app.child.pid, '产品示例'))
    await clickButton(app.child.pid, '产品示例')
    phase = 'login-demo'
    await poll('native Demo page', () => windowContains(app.child.pid, '产品搜索'))
    phase = 'login-boundary'
    await pollBoundary(app.child.pid)
    phase = 'scope-authorization'
    await clickButton(app.child.pid, 'E2E scope self')
    await pollControl(app.child.pid, 'self scope ownership enforced', 'E2E self scope enforced')
    await poll('self scope capability retained', () => windowContains(app.child.pid, '产品搜索'))
    await clickButton(app.child.pid, 'E2E scope all')
    await pollControl(app.child.pid, 'all scope visibility restored', 'E2E all scope restored')
    await poll('all scope capability restored', () => windowContains(app.child.pid, '产品搜索'))
    phase = 'permission-authorization'
    await clickButton(app.child.pid, 'E2E permissions off')
    await pollControl(app.child.pid, 'revoked permission request denied', 'E2E authorization denied')
    await poll('revoked permission capability hidden', () => windowContains(app.child.pid, '无权访问'))
    await clickButton(app.child.pid, 'E2E permissions on')
    await poll('permission capability restored', () => windowContains(app.child.pid, '产品搜索'))
    phase = 'session-revocation'
    await clickButton(app.child.pid, 'E2E revoke session')
    await poll('session revoke requires login', () => windowContains(app.child.pid, '使用管理员账号登录控制台'))
    await login(app.child.pid, 'admin', fixturePassword)
    await poll('native authenticated workspace after relogin', () => windowContains(app.child.pid, 'Administrator'))
    await poll('native Demo navigation after relogin', () => windowContains(app.child.pid, '产品示例'))
    await clickButton(app.child.pid, '产品示例')
    await poll('native Demo page after relogin', () => windowContains(app.child.pid, '产品搜索'))
    phase = 'single-instance'
    const firstSidecar = await newSidecarPid(sidecarBaseline)
    await assertLoopbackOnly(firstSidecar)
    const firstWindowCount = await windowCount(app.child.pid)
    const second = startTracked(binary, liveRoot, liveKeyring)
    const secondExit = await Promise.race([second.exited, delay(10_000).then(() => null)])
    if (secondExit === null) {
      await stopTracked(second)
      throw new Error('second native instance did not exit')
    }
    await stopTracked(second)
    if (await windowCount(second.child.pid) !== 0 || await windowCount(app.child.pid) !== firstWindowCount) {
      throw new Error('second native instance created an additional window')
    }
    if (!(await windowContains(app.child.pid, '产品搜索'))) throw new Error('first native instance stopped serving after duplicate launch')
    if (await newSidecarPid(sidecarBaseline) !== firstSidecar) throw new Error('second native instance spawned another sidecar')
    phase = 'product-create'
    await clickButton(app.child.pid, '新增')
    await poll('native product form', () => windowContains(app.child.pid, '新增产品'))
    await poll('native product save control', () => windowContains(app.child.pid, '保存'))
    await createProduct(app.child.pid)
    await poll('native product create', () => windowContains(app.child.pid, 'E2E-001'))
    phase = 'product-update'
    await updateProduct(app.child.pid)
    await poll('native product update', () => windowContains(app.child.pid, 'Native product updated'))
    await stopTracked(app)
    assertSafeDiagnostics(app.output(), [workspace, liveRoot])
    await assertNoNewSidecars(sidecarBaseline)
    phase = 'persistence-verification'
    await execute('go', [
      'run', './test/desktop/fixture', '--root', liveRoot, '--mode', 'verify', '--expected-product', 'Native product updated'
    ], { cwd: goRoot })

    phase = 'stronghold-restart'
    app = startTracked(binary, liveRoot, liveKeyring)
    await pollRestoredIdentity(app.child.pid)
    await poll('Stronghold Demo navigation', () => windowContains(app.child.pid, '产品示例'))
    await clickButton(app.child.pid, '产品示例')
    await poll('Stronghold session restart', () => windowContains(app.child.pid, '产品搜索'))
    await pollBoundary(app.child.pid)
    await poll('SQLite product restart', () => windowContains(app.child.pid, 'Native product updated'))
    phase = 'product-delete'
    await deleteProduct(app.child.pid)
    await poll('native product delete', async () => !(await windowContains(app.child.pid, 'E2E-001')))
    phase = 'logout-navigation'
    await clickButton(app.child.pid, '账户菜单')
    await poll('native account page', () => windowContains(app.child.pid, '退出登录'))
    phase = 'logout'
    await logout(app.child.pid)
    await poll('native logout', () => windowContains(app.child.pid, '使用管理员账号登录控制台'), 90_000)
    await stopTracked(app)
    assertSafeDiagnostics(app.output(), [workspace, liveRoot])
    phase = 'logout-restart'
    app = startTracked(binary, liveRoot, liveKeyring)
    await poll('logout persistence after restart', () => windowContains(app.child.pid, '使用管理员账号登录控制台'), 90_000)
    if (await windowContains(app.child.pid, '产品搜索')) throw new Error('logout left a restartable desktop session')
    await stopTracked(app)
    assertSafeDiagnostics(app.output(), [workspace, liveRoot])
    await deleteTestKeyring(liveKeyring)
    await assertNoNewSidecars(sidecarBaseline)
    if (await keyringExists(productionKeyringService, productionKeyringAccount)) {
      throw new Error('native E2E created a production credential')
    }
  } catch (error) {
    failure = error instanceof Error && (error.message.startsWith('desktop native command ') || error.message.startsWith('desktop native accessibility '))
      ? qualifyCommandFailure(error, phase)
      : nativePhaseFailure(phase, app?.output() ?? '')
  } finally {
    const cleanups = [
      async () => {
        phase = 'cleanup-hosts'
        let cleanupFailure
        for (const owner of owners) {
          try {
            await stopTracked(owner)
          } catch (error) {
            cleanupFailure ??= error
          }
        }
        if (cleanupFailure) throw cleanupFailure
      },
      () => {
        phase = 'cleanup-failed-credential'
        return deleteTestKeyring(failedKeyring)
      },
      () => {
        phase = 'cleanup-live-credential'
        return deleteTestKeyring(liveKeyring)
      },
      async () => {
        phase = 'cleanup-sidecars'
        const leaked = await reapNewSidecars(sidecarBaseline)
        if (leaked.size !== 0) throw new Error('desktop sidecar leak was recovered')
      },
      () => {
        phase = 'cleanup-production-artifacts'
        return restoreProductionArtifacts()
      },
      async () => {
        phase = 'cleanup-verification'
        await assertNoNewSidecars(sidecarBaseline)
        if (await keyringExists(testKeyringService, failedKeyring) || await keyringExists(testKeyringService, liveKeyring)) {
          throw new Error('native E2E left a test credential')
        }
      },
      () => {
        phase = 'cleanup-workspace'
        return rm(workspace, { recursive: true, force: true })
      }
    ]
    for (const cleanup of cleanups) {
      try {
        await cleanup()
      } catch (error) {
        failure ??= qualifyCommandFailure(error, phase)
      }
    }
  }
  if (failure) throw failure
  process.stdout.write(`${JSON.stringify({ state: 'passed', runtime: 'tauri-native', profile: 'sqlite' })}\n`)
}

main().catch(error => {
  process.stderr.write(`${safeText(error instanceof Error ? error.message : 'desktop native E2E failed')}\n`)
  process.exitCode = 1
})
