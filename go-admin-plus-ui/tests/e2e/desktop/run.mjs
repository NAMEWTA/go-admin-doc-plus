#!/usr/bin/env node

import { mkdir, mkdtemp, open, readFile, readdir, realpath, rm } from 'node:fs/promises'
import { spawn } from 'node:child_process'
import { createHash, randomBytes } from 'node:crypto'
import { createConnection } from 'node:net'
import { networkInterfaces, tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { execute, reapNewSidecars, sidecarProcesses } from './processes.mjs'

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

const fileContains = async (path, needle) => {
  const file = await open(path, 'r')
  const target = Buffer.from(needle)
  const chunk = Buffer.alloc(64 * 1024 + target.length)
  let overlap = 0
  try {
    for (;;) {
      const { bytesRead } = await file.read(chunk, overlap, 64 * 1024, null)
      if (bytesRead === 0) return false
      const length = overlap + bytesRead
      if (chunk.subarray(0, length).includes(target)) return true
      overlap = Math.min(target.length - 1, length)
      chunk.copy(chunk, 0, length - overlap, length)
    }
  } finally {
    await file.close()
  }
}

const directoryContains = async (directory, needle) => {
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const path = join(directory, entry.name)
    if (entry.isDirectory() ? await directoryContains(path, needle) : entry.isFile() && await fileContains(path, needle)) return true
  }
  return false
}

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
  await execute('cargo', ['build', '--locked', '--quiet', '--release'], { cwd: rustRoot, timeout: 300_000 })
  if (await fileContains(sidecarBinary, '/__desktop/test-control') ||
    await fileContains(hostBinary, '/__desktop/test-control') ||
    await directoryContains(join(appRoot, 'dist'), 'E2E scope self')) {
    throw new Error('production desktop artifacts retained native test controls')
  }
}

const assertSafeDiagnostics = (output, protectedRoots) => {
  const lower = output.toLowerCase()
  const forbidden = ['__host-go-admin-session', 'csrf', 'cookie', 'readiness', 'controltoken', 'bearer ', 'session token']
  if (output.includes(fixturePassword) || protectedRoots.some(root => output.includes(root)) || forbidden.some(value => lower.includes(value))) {
    throw new Error('native diagnostics leaked protected material')
  }
}

const runAppleScript = script => execute('/usr/bin/osascript', ['-'], { input: script, timeout: 10_000 })
const quoteAppleScript = value => `"${value.replaceAll('\\', '\\\\').replaceAll('"', '\\"')}"`

const windowCount = async pid => {
  const output = await runAppleScript(`tell application "System Events"
if exists (first process whose unix id is ${pid}) then
  tell (first process whose unix id is ${pid}) to return count of windows
end if
return 0
end tell`)
  return Number.parseInt(output.trim(), 10) || 0
}

const windowContains = async (pid, value) => {
  const output = await runAppleScript(`tell application "System Events"
if not (exists (first process whose unix id is ${pid})) then return "false"
tell (first process whose unix id is ${pid})
  if (count of windows) is 0 then return "false"
  return ((name of every UI element of entire contents of window 1) as text) contains ${quoteAppleScript(value)}
end tell
end tell`)
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

const login = (pid, username, password) => runAppleScript(`tell application "System Events"
tell (first process whose unix id is ${pid})
  tell window 1
    set value of text field 1 to ${quoteAppleScript(username)}
    set value of text field 2 to ${quoteAppleScript(password)}
    click button "登录"
  end tell
end tell
end tell`)

const createProduct = pid => runAppleScript(`tell application "System Events"
tell (first process whose unix id is ${pid})
  tell window 1
    set value of text field 2 to "E2E-001"
    set value of text field 3 to "Native product"
    set value of text area 1 to "created through the native window"
    set value of text field 4 to "1250"
    click button "Save"
  end tell
end tell
end tell`)

const updateProduct = pid => runAppleScript(`tell application "System Events"
tell (first process whose unix id is ${pid})
  tell window 1
    click first button whose name is "Edit"
    set value of text field 3 to "Native product updated"
    click button "Save"
  end tell
end tell
end tell`)

const deleteProduct = pid => runAppleScript(`tell application "System Events"
tell (first process whose unix id is ${pid})
  tell window 1 to click first button whose name is "Delete"
  repeat 100 times
    if (count of sheets of window 1) > 0 then
      tell sheet 1 of window 1 to click button "OK"
      return
    end if
    delay 0.05
  end repeat
  error "native delete confirmation unavailable"
end tell
end tell`)

const logout = pid => runAppleScript(`tell application "System Events"
tell (first process whose unix id is ${pid}) to tell window 1 to click button "退出"
end tell`)

const clickButton = (pid, name) => runAppleScript(`tell application "System Events"
tell (first process whose unix id is ${pid}) to tell window 1 to click first button whose name is ${quoteAppleScript(name)}
end tell`)

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
  try {
    if (await keyringExists(productionKeyringService, productionKeyringAccount)) {
      throw new Error('production desktop credential pre-existed; native E2E refuses to touch it')
    }
    if (await keyringExists(testKeyringService, failedKeyring) || await keyringExists(testKeyringService, liveKeyring)) {
      throw new Error('native test credential identity collision')
    }
    await execute(process.execPath, [join(root, 'release/shared/sidecar/build.mjs'), '--native-e2e', '--target', 'aarch64-apple-darwin'], { cwd: root })
    await execute(join(appRoot, 'node_modules/.bin/vite'), ['build', '--config', 'vite.config.ts'], {
      cwd: appRoot,
      env: { PATH: process.env.PATH ?? '', HOME: process.env.HOME ?? '', VITE_GO_ADMIN_NATIVE_E2E: '1' }
    })
    await execute('cargo', ['build', '--locked', '--quiet', '--release', '--features', 'native-e2e'], { cwd: rustRoot, timeout: 600_000 })
    const binary = hostBinary

    const failedRoot = join(workspace, 'failed')
    await mkdir(failedRoot, { recursive: true, mode: 0o700 })
    await execute('go', ['run', './test/desktop/fixture', '--root', failedRoot, '--mode', 'migration-failure'], { cwd: goRoot })
    const database = join(failedRoot, 'data/go-admin-plus.db')
    const beforeFailure = await hashFile(database)
    const failed = startTracked(binary, failedRoot, failedKeyring)
    const failureDeadline = Date.now() + 15_000
    while (Date.now() < failureDeadline && failed.child.exitCode === null) {
      if (await windowCount(failed.child.pid) !== 0) throw new Error('migration failure opened the native window')
      await delay(100)
    }
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

    const liveRoot = join(workspace, 'live')
    await mkdir(liveRoot, { recursive: true, mode: 0o700 })
    await execute('go', ['run', './test/desktop/fixture', '--root', liveRoot, '--mode', 'previous'], { cwd: goRoot })
    app = startTracked(binary, liveRoot, liveKeyring)
    await poll('native login window', () => windowContains(app.child.pid, '登录'), 90_000)
    await login(app.child.pid, 'admin', fixturePassword)
    await poll('native Demo page', () => windowContains(app.child.pid, 'Products'))
    await poll('authenticated WebView storage and URL boundary', () => windowContains(app.child.pid, 'E2E authenticated boundary verified'))
    await clickButton(app.child.pid, 'E2E scope self')
    await poll('self scope request denied', () => windowContains(app.child.pid, 'E2E authorization denied'))
    await poll('self scope capability hidden', () => windowContains(app.child.pid, '无权访问'))
    await clickButton(app.child.pid, 'E2E scope all')
    await poll('all scope capability restored', () => windowContains(app.child.pid, 'Products'))
    await clickButton(app.child.pid, 'E2E permissions off')
    await poll('revoked permission request denied', () => windowContains(app.child.pid, 'E2E authorization denied'))
    await poll('revoked permission capability hidden', () => windowContains(app.child.pid, '无权访问'))
    await clickButton(app.child.pid, 'E2E permissions on')
    await poll('permission capability restored', () => windowContains(app.child.pid, 'Products'))
    await clickButton(app.child.pid, 'E2E revoke session')
    await poll('session revoke requires login', () => windowContains(app.child.pid, '登录'))
    await login(app.child.pid, 'admin', fixturePassword)
    await poll('native Demo page after relogin', () => windowContains(app.child.pid, 'Products'))
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
    if (!(await windowContains(app.child.pid, 'Products'))) throw new Error('first native instance stopped serving after duplicate launch')
    if (await newSidecarPid(sidecarBaseline) !== firstSidecar) throw new Error('second native instance spawned another sidecar')
    await createProduct(app.child.pid)
    await poll('native product create', () => windowContains(app.child.pid, 'E2E-001'))
    await updateProduct(app.child.pid)
    await poll('native product update', () => windowContains(app.child.pid, 'Native product updated'))
    await stopTracked(app)
    assertSafeDiagnostics(app.output(), [workspace, liveRoot])
    await assertNoNewSidecars(sidecarBaseline)

    app = startTracked(binary, liveRoot, liveKeyring)
    await poll('Stronghold session restart', () => windowContains(app.child.pid, 'Products'))
    await poll('restarted authenticated WebView boundary', () => windowContains(app.child.pid, 'E2E authenticated boundary verified'))
    await poll('SQLite product restart', () => windowContains(app.child.pid, 'Native product updated'))
    await execute('go', [
      'run', './test/desktop/fixture', '--root', liveRoot, '--mode', 'verify', '--expected-product', 'Native product updated'
    ], { cwd: goRoot })
    await deleteProduct(app.child.pid)
    await poll('native product delete', async () => !(await windowContains(app.child.pid, 'E2E-001')))
    await logout(app.child.pid)
    await poll('native logout', () => windowContains(app.child.pid, '登录'))
    await stopTracked(app)
    assertSafeDiagnostics(app.output(), [workspace, liveRoot])
    app = startTracked(binary, liveRoot, liveKeyring)
    await poll('logout persistence after restart', () => windowContains(app.child.pid, '登录'), 90_000)
    if (await windowContains(app.child.pid, 'Products')) throw new Error('logout left a restartable desktop session')
    await stopTracked(app)
    assertSafeDiagnostics(app.output(), [workspace, liveRoot])
    await deleteTestKeyring(liveKeyring)
    await assertNoNewSidecars(sidecarBaseline)
    if (await keyringExists(productionKeyringService, productionKeyringAccount)) {
      throw new Error('native E2E created a production credential')
    }
  } catch (error) {
    failure = error
  } finally {
    const cleanups = [
      async () => {
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
      () => deleteTestKeyring(failedKeyring),
      () => deleteTestKeyring(liveKeyring),
      async () => {
        const leaked = await reapNewSidecars(sidecarBaseline)
        if (leaked.size !== 0) throw new Error('desktop sidecar leak was recovered')
      },
      () => restoreProductionArtifacts(),
      async () => {
        await assertNoNewSidecars(sidecarBaseline)
        if (await keyringExists(testKeyringService, failedKeyring) || await keyringExists(testKeyringService, liveKeyring)) {
          throw new Error('native E2E left a test credential')
        }
      },
      () => rm(workspace, { recursive: true, force: true })
    ]
    for (const cleanup of cleanups) {
      try {
        await cleanup()
      } catch (error) {
        failure ??= error
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
