import { spawn, spawnSync } from 'node:child_process'
import { existsSync, mkdtempSync, readFileSync, rmSync } from 'node:fs'
import { request as httpsRequest } from 'node:https'
import { tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const testRoot = dirname(fileURLToPath(import.meta.url))
const uiRoot = resolve(testRoot, '../../../..')
const repositoryRoot = resolve(uiRoot, '..')
const backendRoot = join(repositoryRoot, 'go-admin-plus')
const requireE2E = process.env.GO_ADMIN_REQUIRE_IAM_E2E === '1'
const chromium = process.env.GO_ADMIN_TEST_CHROMIUM_EXECUTABLE
const postgresDSN = process.env.GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN
const temporaryRoot = mkdtempSync(join(tmpdir(), 'go-admin-iam-e2e-'))
const staticRoot = join(temporaryRoot, 'static')

const fail = (message) => { throw new Error(`IAM session E2E: ${message}`) }
const assert = (condition, message) => { if (!condition) fail(message) }
const delay = (milliseconds) => new Promise((resolvePromise) => setTimeout(resolvePromise, milliseconds))

const runChecked = (command, args, options = {}) => {
  const result = spawnSync(command, args, {
    cwd: options.cwd,
    env: options.env ?? process.env,
    encoding: 'utf8',
    stdio: options.quiet ? 'pipe' : 'inherit',
  })
  if (result.error || result.status !== 0) fail(options.failure ?? `${command} failed`)
}

class CDPClient {
  constructor(socket) {
    this.socket = socket
    this.nextID = 1
    this.pending = new Map()
    socket.addEventListener('message', (event) => {
      const message = JSON.parse(String(event.data))
      if (!message.id) return
      const pending = this.pending.get(message.id)
      if (!pending) return
      this.pending.delete(message.id)
      if (message.error) pending.reject(new Error('CDP command failed'))
      else pending.resolve(message.result)
    })
    socket.addEventListener('close', () => {
      for (const pending of this.pending.values()) pending.reject(new Error('CDP connection closed'))
      this.pending.clear()
    })
  }

  send(method, params = {}, sessionId) {
    return new Promise((resolvePromise, reject) => {
      const id = this.nextID++
      this.pending.set(id, { resolve: resolvePromise, reject })
      this.socket.send(JSON.stringify({ id, method, params, ...(sessionId ? { sessionId } : {}) }))
    })
  }
}

const connectWebSocket = (url) => new Promise((resolvePromise, reject) => {
  const socket = new WebSocket(url)
  const timer = setTimeout(() => reject(new Error('CDP connection timed out')), 15000)
  socket.addEventListener('open', () => { clearTimeout(timer); resolvePromise(socket) }, { once: true })
  socket.addEventListener('error', () => { clearTimeout(timer); reject(new Error('CDP connection failed')) }, { once: true })
})

const waitForFile = async (path, child) => {
  for (let attempt = 0; attempt < 600; attempt += 1) {
    if (existsSync(path)) return readFileSync(path, 'utf8').trim()
    if (child.exitCode !== null) fail('HTTPS test host exited before readiness')
    await delay(100)
  }
  fail('HTTPS test host readiness timed out')
}

const waitForDevTools = (child) => new Promise((resolvePromise, reject) => {
  let buffered = ''
  const timer = setTimeout(() => reject(new Error('Chromium DevTools endpoint timed out')), 30000)
  child.stderr.setEncoding('utf8')
  child.stderr.on('data', (chunk) => {
    buffered += chunk
    const match = buffered.match(/DevTools listening on (ws:\/\/[^\s]+)/)
    if (!match) return
    clearTimeout(timer)
    resolvePromise(match[1])
  })
  child.once('exit', () => {
    clearTimeout(timer)
    reject(new Error('Chromium exited before DevTools readiness'))
  })
})

const waitForTarget = async (client, baseURL) => {
  for (let attempt = 0; attempt < 300; attempt += 1) {
    const { targetInfos } = await client.send('Target.getTargets')
    const target = targetInfos.find((candidate) => candidate.type === 'page' && candidate.url.startsWith(baseURL))
    if (target) return target.targetId
    await delay(100)
  }
  fail('browser page target did not load')
}

const evaluate = async (client, sessionId, expression) => {
  const outcome = await client.send('Runtime.evaluate', {
    expression,
    awaitPromise: true,
    returnByValue: true,
  }, sessionId)
  if (outcome.exceptionDetails) fail('browser scenario raised an exception')
  return outcome.result.value
}

const currentCookie = async (client, sessionId) => {
  const { cookies } = await client.send('Network.getAllCookies', {}, sessionId)
  const matches = cookies.filter((cookie) => cookie.name === '__Host-go-admin-session')
  assert(matches.length === 1, 'expected exactly one host session cookie')
  return matches[0]
}

const shutdownHost = (baseURL) => new Promise((resolvePromise) => {
  if (!baseURL) return resolvePromise()
  const endpoint = new URL('/__test/shutdown', baseURL)
  const request = httpsRequest(endpoint, { method: 'POST', rejectUnauthorized: false }, (response) => {
    response.resume()
    response.once('end', resolvePromise)
  })
  request.once('error', () => resolvePromise())
  request.end()
})

const waitForExit = (child) => new Promise((resolvePromise) => {
  if (child.exitCode !== null) return resolvePromise(child.exitCode)
  child.once('exit', (code) => resolvePromise(code))
})

const runProfile = async (profile) => {
  const profileRoot = join(temporaryRoot, profile)
  const readyFile = join(profileRoot, 'ready')
  const browserRoot = join(profileRoot, 'chromium')
  const host = spawn('go', ['test', './test/iam/session', '-run', '^TestIAMBrowserHarnessServer$', '-count=1', '-v'], {
    cwd: backendRoot,
    env: {
      ...process.env,
      GO_ADMIN_IAM_E2E_SERVE: '1',
      GO_ADMIN_IAM_E2E_PROFILE: profile,
      GO_ADMIN_IAM_E2E_READY_FILE: readyFile,
      GO_ADMIN_IAM_E2E_STATIC_DIR: staticRoot,
    },
    stdio: ['ignore', 'pipe', 'pipe'],
  })
  let browser
  let baseURL
  try {
    baseURL = await waitForFile(readyFile, host)
    browser = spawn(chromium, [
      '--headless=new',
      '--disable-gpu',
      '--ignore-certificate-errors',
      '--no-first-run',
      '--no-default-browser-check',
      '--remote-debugging-port=0',
      `--user-data-dir=${browserRoot}`,
      baseURL,
    ], { stdio: ['ignore', 'ignore', 'pipe'] })
    const devToolsURL = await waitForDevTools(browser)
    const socket = await connectWebSocket(devToolsURL)
    const cdp = new CDPClient(socket)
    const targetID = await waitForTarget(cdp, baseURL)
    const { sessionId } = await cdp.send('Target.attachToTarget', { targetId: targetID, flatten: true })
    await cdp.send('Runtime.enable', {}, sessionId)
    await cdp.send('Network.enable', {}, sessionId)
    for (let attempt = 0; attempt < 300; attempt += 1) {
      if (await evaluate(cdp, sessionId, 'Boolean(globalThis.__iamE2E)')) break
      if (attempt === 299) fail('browser driver did not initialize')
      await delay(100)
    }

    assert(await evaluate(cdp, sessionId, 'globalThis.__iamE2E.loginAndCheckState()') === true, 'login scenario failed')
    const initialCookie = await currentCookie(cdp, sessionId)
    assert(initialCookie.secure === true, 'session cookie is not Secure')
    assert(initialCookie.httpOnly === true, 'session cookie is not HttpOnly')
    assert(initialCookie.sameSite === 'Strict', 'session cookie is not SameSite=Strict')
    assert(initialCookie.path === '/', 'session cookie path is not host-wide')
    assert(!initialCookie.domain.startsWith('.'), 'session cookie set a Domain attribute')
    const readableCookies = await evaluate(cdp, sessionId, 'document.cookie')
    assert(!String(readableCookies).includes('__Host-go-admin-session'), 'session cookie is readable by document.cookie')

    assert(await evaluate(cdp, sessionId, 'globalThis.__iamE2E.verifyCSRF()') === true, 'CSRF scenarios failed')
    const beforeRotation = await currentCookie(cdp, sessionId)
    assert(await evaluate(cdp, sessionId, 'globalThis.__iamE2E.rotate()') === true, 'rotation scenario failed')
    const replacement = await currentCookie(cdp, sessionId)
    assert(beforeRotation.value !== replacement.value, 'rotation did not replace the opaque cookie')
    await cdp.send('Network.setCookie', {
      name: beforeRotation.name,
      value: beforeRotation.value,
      url: baseURL,
      path: '/',
      secure: true,
      httpOnly: true,
      sameSite: 'Strict',
    }, sessionId)
    const oldStatus = await evaluate(cdp, sessionId, "fetch('/api/iam/session/current').then((response) => response.status)")
    assert(oldStatus === 401, 'rotated cookie recovered')
    await cdp.send('Network.setCookie', {
      name: replacement.name,
      value: replacement.value,
      url: baseURL,
      path: '/',
      secure: true,
      httpOnly: true,
      sameSite: 'Strict',
    }, sessionId)
    assert(await evaluate(cdp, sessionId, 'globalThis.__iamE2E.recoverReplacementCookie()') === true, 'replacement cookie recovery failed')
    assert(await evaluate(cdp, sessionId, 'globalThis.__iamE2E.updateProfile()') === true, 'profile scenario failed')
    assert(await evaluate(cdp, sessionId, 'globalThis.__iamE2E.verifyLogoutRetry()') === true, 'logout retry scenario failed')
    assert(await evaluate(cdp, sessionId, 'globalThis.__iamE2E.verifyIdleTimeout()') === true, 'idle timeout scenario failed')
    assert(await evaluate(cdp, sessionId, 'globalThis.__iamE2E.verifyAbsoluteTimeout()') === true, 'absolute timeout scenario failed')
    assert(await evaluate(cdp, sessionId, 'globalThis.__iamE2E.verifyPasswordChange()') === true, 'password scenario failed')
    await evaluate(cdp, sessionId, 'globalThis.__iamE2E.shutdown()')
    socket.close()
    const hostExit = await waitForExit(host)
    assert(hostExit === 0, 'HTTPS test host reported failure')
  } finally {
    if (browser && browser.exitCode === null) browser.kill('SIGTERM')
    await shutdownHost(baseURL)
    if (host.exitCode === null) {
      const exited = await Promise.race([waitForExit(host), delay(5000).then(() => false)])
      if (exited === false) host.kill('SIGTERM')
    }
  }
}

try {
  runChecked('pnpm', ['exec', 'vite', 'build', '--config', join(testRoot, 'vite.config.ts')], {
    cwd: uiRoot,
    env: { ...process.env, GO_ADMIN_IAM_E2E_OUT_DIR: staticRoot },
    failure: 'browser fixture build failed',
  })
  runChecked('go', ['test', './test/iam/session', '-run', '^TestIAMBrowserHarnessServer$', '-count=1'], {
    cwd: backendRoot,
    env: { ...process.env, GO_ADMIN_IAM_E2E_SERVE: '0' },
    quiet: true,
    failure: 'HTTPS test host compile self-check failed',
  })

  const missing = []
  if (!chromium) missing.push('GO_ADMIN_TEST_CHROMIUM_EXECUTABLE')
  if (!postgresDSN) missing.push('GO_ADMIN_TEST_POSTGRES_DISPOSABLE_DSN')
  if (missing.length) {
    if (requireE2E) fail(`required environment is missing: ${missing.join(', ')}`)
    console.log(`IAM_SESSION_E2E_SKIP missing environment: ${missing.join(', ')}`)
  } else {
    await runProfile('sqlite')
    await runProfile('postgres')
    console.log('IAM_SESSION_E2E_PASS profiles=sqlite,postgres')
  }
} finally {
  rmSync(temporaryRoot, { recursive: true, force: true })
}
