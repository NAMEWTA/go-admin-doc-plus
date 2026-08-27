#!/usr/bin/env node

import { spawn } from 'node:child_process'
import { writeFile } from 'node:fs/promises'
import process from 'node:process'

const argumentsByName = new Map()
for (let index = 2; index < process.argv.length; index += 2) argumentsByName.set(process.argv[index], process.argv[index + 1])
const application = argumentsByName.get('--application')
const evidenceFile = argumentsByName.get('--evidence')
if (process.platform !== 'win32' || process.env.CI !== 'true' || process.env.GITHUB_ACTIONS !== 'true' ||
    !application || !evidenceFile || process.argv.length !== 6) {
  throw new Error('installed Windows tracer requires an ephemeral CI runner and exact paths')
}

const endpoint = 'http://127.0.0.1:4444'
const delay = milliseconds => new Promise(resolve => setTimeout(resolve, milliseconds))
const driver = spawn('tauri-driver.exe', [], { stdio: ['ignore', 'ignore', 'pipe'], windowsHide: true })
let driverError = ''
driver.stderr.on('data', chunk => { driverError = `${driverError}${chunk}`.slice(-4096) })

const request = async (path, method = 'GET', body) => {
  const response = await fetch(`${endpoint}${path}`, {
    method,
    headers: body === undefined ? undefined : { 'content-type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body)
  })
  const payload = await response.json().catch(() => ({}))
  if (!response.ok || payload.value?.error) throw new Error('WebDriver command failed')
  return payload.value
}
const poll = async (description, operation, timeout = 90_000) => {
  const end = Date.now() + timeout
  while (Date.now() < end) {
    try { if (await operation()) return } catch { /* application may still be starting */ }
    await delay(200)
  }
  throw new Error(`${description} timed out`)
}
const waitForDriver = () => poll('tauri-driver readiness', async () => {
  if (driver.exitCode !== null) throw new Error('tauri-driver exited before accepting a session')
  await request('/status')
  return true
}, 30_000)
const createSession = async () => {
  const value = await request('/session', 'POST', {
    capabilities: { alwaysMatch: { browserName: 'wry', 'tauri:options': { application } } }
  })
  if (typeof value.sessionId !== 'string' || value.sessionId.length === 0) throw new Error('WebDriver session identifier is invalid')
  return value.sessionId
}
const execute = (session, script, args = []) => request(`/session/${session}/execute/sync`, 'POST', { script, args })
const closeSession = async session => {
  await request(`/session/${session}`, 'DELETE')
  await delay(1000)
}
const loginIfRequired = async session => {
  await poll('login or restored workspace', () => execute(session, `
    return Boolean(document.querySelector('form[aria-label="Sign in"]') || document.querySelector('nav[aria-label="主导航"]'))
  `))
  const loginVisible = await execute(session, 'return Boolean(document.querySelector(\'form[aria-label="Sign in"]\'))')
  if (!loginVisible) return false
  await execute(session, `
    const form = document.querySelector('form[aria-label="Sign in"]')
    const username = form?.querySelector('input[autocomplete="username"]')
    const password = form?.querySelector('input[autocomplete="current-password"]')
    if (!form || !username || !password) return false
    username.value = arguments[0]
    password.value = arguments[1]
    username.dispatchEvent(new Event('input', { bubbles: true }))
    password.dispatchEvent(new Event('input', { bubbles: true }))
    form.requestSubmit()
    return true
  `, ['admin', 'administrator password'])
  await poll('authenticated workspace', () => execute(session, 'return Boolean(document.querySelector(\'nav[aria-label="主导航"]\'))'))
  return true
}
const openProducts = async session => {
  await poll('Demo products navigation', () => execute(session, `
    const button = [...document.querySelectorAll('nav[aria-label="主导航"] button')].find(value => value.textContent?.trim() === 'Demo products')
    if (!button) return false
    button.click()
    return true
  `))
  await poll('Products page', () => execute(session, 'return Boolean(document.querySelector(\'#demo-products-title\'))'))
}

let activeSession
try {
  await waitForDriver()
  const sku = `WIN-${Date.now()}`
  activeSession = await createSession()
  if (!await loginIfRequired(activeSession)) throw new Error('first installed launch unexpectedly restored a prior session')
  await openProducts(activeSession)
  const submitted = await execute(activeSession, `
    const form = document.querySelector('.demo-products__form')
    const values = { sku: arguments[0], name: 'Windows release tracer', description: 'Installed Tauri NSIS', priceCents: '1250', status: 'active' }
    if (!form) return false
    for (const [name, value] of Object.entries(values)) {
      const control = form.querySelector('[name="' + name + '"]')
      if (!control) return false
      control.value = value
      control.dispatchEvent(new Event(control.tagName === 'SELECT' ? 'change' : 'input', { bubbles: true }))
    }
    form.requestSubmit()
    return true
  `, [sku])
  if (!submitted) throw new Error('installed product form is incomplete')
  await poll('created product', () => execute(activeSession, 'return document.querySelector(\'tbody\')?.textContent?.includes(arguments[0]) === true', [sku]))
  await closeSession(activeSession)
  activeSession = undefined

  activeSession = await createSession()
  await loginIfRequired(activeSession)
  await openProducts(activeSession)
  await poll('persisted product after restart', () => execute(activeSession, 'return document.querySelector(\'tbody\')?.textContent?.includes(arguments[0]) === true', [sku]))
  const deleted = await execute(activeSession, `
    window.confirm = () => true
    const row = [...document.querySelectorAll('tbody tr')].find(value => value.textContent?.includes(arguments[0]))
    const button = row && [...row.querySelectorAll('button')].find(value => value.textContent?.trim() === 'Delete')
    if (!button) return false
    button.click()
    return true
  `, [sku])
  if (!deleted) throw new Error('persisted product could not be deleted')
  await poll('deleted product', () => execute(activeSession, 'return document.querySelector(\'tbody\')?.textContent?.includes(arguments[0]) !== true', [sku]))
  await closeSession(activeSession)
  activeSession = undefined

  await writeFile(evidenceFile, `${JSON.stringify({
    schemaVersion: 1, driver: 'tauri-driver', firstLaunchLogin: 'passed', create: 'passed',
    restart: 'passed', persistence: 'passed', delete: 'passed'
  }, null, 2)}\n`, { encoding: 'utf8', flag: 'wx' })
  process.stdout.write('GO_ADMIN_WINDOWS_INSTALLED_TRACER_PASS\n')
} catch (error) {
  if (activeSession) await closeSession(activeSession).catch(() => {})
  if (driverError) process.stderr.write('tauri-driver diagnostics were captured\n')
  throw error
} finally {
  driver.kill()
}
