import { createApp, h, type Component } from 'vue'
import { createCapabilityController } from '@go-admin/domain-iam/administration'
import { createSessionController } from '@go-admin/domain-iam/session'
import { FilesRequestError, filesPermissions } from '@go-admin/domain-files'
import { createBrowserFilesClient, createBrowserSessionFetch } from '@go-admin/adapter-browser'
import { createWebAdministrationClient } from '@go-admin/web-domain-iam/administration'
import { createWebSessionClient } from '@go-admin/web-domain-iam/session'
import { createFilesController, FilesPage } from '@go-admin/web-domain-files'

const assert: (condition: unknown, message: string) => asserts condition = (condition, message) => { if (!condition) throw new Error(message) }
const waitUntil = async (condition: () => boolean, message: string, timeout = 15_000) => {
  const deadline = Date.now() + timeout
  while (Date.now() < deadline) { if (condition()) return; await new Promise(resolve => setTimeout(resolve, 25)) }
  throw new Error(message)
}
const element = <T extends Element>(selector: string) => { const value = document.querySelector<T>(selector); assert(value, `missing ${selector}`); return value }
const control = async (path: string, body?: unknown) => {
  const response = await fetch(`/__test/${path}`, body === undefined ? {} : { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
  assert(response.status === (body === undefined ? 200 : 204), `control ${path} failed`)
  return response
}
const setSearch = async (value: string) => {
  const input = element<HTMLInputElement>('[name="search"]')
  input.value = value
  input.dispatchEvent(new Event('input', { bubbles: true }))
  element<HTMLFormElement>('.files-page__search').requestSubmit()
  await Promise.resolve()
}
const uploadThroughPage = async (name: string, content: string) => {
  const input = element<HTMLInputElement>('.files-page__upload [name="file"]')
  const transfer = new DataTransfer()
  transfer.items.add(new File([content], name, { type: 'text/plain' }))
  input.files = transfer.files
  input.dispatchEvent(new Event('change', { bubbles: true }))
  await Promise.resolve()
  element<HTMLButtonElement>('[data-testid="files-upload"] button').click()
  await waitUntil(() => [...document.querySelectorAll('tbody tr')].some(row => row.textContent?.includes(name)), `upload ${name} did not render`)
}
const expectFailure = async (operation: () => Promise<unknown>, category: 'forbidden' | 'relogin') => {
  try { await operation() } catch (error) {
    assert(error instanceof FilesRequestError && error.category === category, `expected ${category}`)
    return
  }
  throw new Error(`expected ${category}`)
}

const scenario = async () => {
  const sessionFetch = createBrowserSessionFetch(fetch)
  const session = createSessionController(createWebSessionClient(sessionFetch, '/api'))
  await session.login({ username: 'files-admin', password: 'files contract password' })
  assert(session.state().status === 'authenticated', 'administrator login failed')
  assert(!document.cookie.includes('__Host-go-admin-session'), 'HttpOnly cookie became script readable')

  const capabilities = createCapabilityController(createWebAdministrationClient(sessionFetch, '/api'))
  await capabilities.refresh()
  assert(capabilities.can(filesPermissions.read) && capabilities.can(filesPermissions.write) && capabilities.can(filesPermissions.delete), 'files capability manifest incomplete')
  const client = createBrowserFilesClient(sessionFetch, '/api')

  await control('scope', { scope: 'self' })
  await capabilities.refresh()
  assert(capabilities.state().manifest?.dataScope === 'self', 'self scope not projected')
  const selfPage = await client.list({ search: '', page: 1, pageSize: 20, sort: 'createdAt', direction: 'descending' })
  assert(selfPage.total === 0, 'self scope leaked foreign metadata')
  await control('scope', { scope: 'all' })
  await capabilities.refresh()

  const controller = createFilesController(client, async () => true, { can: permission => capabilities.can(permission) })
  let sessionRequired = false
  const app = createApp({ render: () => h(FilesPage as Component, { controller, onSessionRequired: () => { sessionRequired = true } }) })
  app.mount('#app')
  await waitUntil(() => controller.list.snapshot().rows.some(row => row.originalName === 'foreign.txt'), 'foreign fixture did not load')
  const foreign = controller.list.snapshot().rows.find(row => row.originalName === 'foreign.txt')!

  for (const [name, content] of [['% literal.txt', 'percent'], ['_ literal.txt', 'underscore'], ['<:@ collision.txt', 'ascii'], ['ä collision.txt', 'unicode']] as const) {
    await uploadThroughPage(name, content)
  }
  for (const [search, expected] of [['%', '% literal.txt'], ['_', '_ literal.txt'], ['ä', 'ä collision.txt'], ['<:@', '<:@ collision.txt']] as const) {
    await setSearch(search)
    await waitUntil(() => controller.list.snapshot().rows.length === 1 && controller.list.snapshot().rows[0]?.originalName === expected, `literal search ${search} failed`)
  }
  const reset = [...document.querySelectorAll<HTMLButtonElement>('.files-page__search button')].find(button => button.textContent === 'Reset')
  assert(reset, 'search reset missing')
  reset.click()
  await waitUntil(() => controller.list.snapshot().total === 5, 'search reset failed')

  const ascii = controller.list.snapshot().rows.find(row => row.originalName === '<:@ collision.txt')
  assert(ascii, 'download fixture missing')
  assert(await (await client.download(ascii.id)).text() === 'ascii', 'download bytes differ')
  const beforeRestart = await (await control('snapshot')).json() as { metadata: number; ready: number; objects: number }
  await control('restart', {})
  await controller.list.refresh()
  assert(controller.list.snapshot().total === 5, 'restart lost metadata')
  assert(await (await client.download(ascii.id)).text() === 'ascii', 'restart lost content')
  const afterRestart = await (await control('snapshot')).json() as typeof beforeRestart
  assert(afterRestart.metadata === beforeRestart.metadata && afterRestart.ready === beforeRestart.ready, 'restart changed durable state')

  const csrfBody = new FormData()
  csrfBody.append('file', new Blob(['denied'], { type: 'text/plain' }), 'csrf-denied.txt')
  const csrfResponse = await fetch('/api/files/objects', { method: 'POST', body: csrfBody })
  const csrfProblem = await csrfResponse.json() as { code?: string }
  assert(csrfResponse.status === 403 && csrfProblem.code === 'CSRF_REJECTED', 'missing CSRF was not rejected')
  const afterCSRF = await (await control('snapshot')).json() as typeof beforeRestart
  assert(afterCSRF.metadata === beforeRestart.metadata, 'CSRF rejection changed metadata')

  const owned = controller.list.snapshot().rows.filter(row => row.originalName !== 'foreign.txt')
  for (const row of owned) element<HTMLInputElement>(`tr[data-file-id="${row.id}"] input[type="checkbox"]`).click()
  await waitUntil(() => !element<HTMLButtonElement>('[data-testid="files-delete-selected"]').disabled, 'batch selection did not render')
  element<HTMLButtonElement>('[data-testid="files-delete-selected"]').click()
  await waitUntil(() => controller.list.snapshot().total === 1, 'batch delete did not complete')
  assert(controller.list.snapshot().rows[0]?.id === foreign.id, 'batch delete removed foreign object')

  await control('scope', { scope: 'self' })
  await capabilities.refresh()
  await controller.list.refresh()
  assert(controller.list.snapshot().total === 0, 'self scope did not hide foreign row')
  await expectFailure(() => client.download(foreign.id), 'forbidden')
  await control('scope', { scope: 'all' })
  await capabilities.refresh()
  await controller.list.refresh()

  await control('permissions', { enabled: false })
  await capabilities.refresh()
  await setSearch('')
  assert(!capabilities.can(filesPermissions.write) && !capabilities.can(filesPermissions.delete), 'revoked capabilities remained')
  await waitUntil(() => document.querySelector('[data-testid="files-upload"]') === null && document.querySelector('[data-testid="files-delete-selected"]') === null, 'revoked mutation UI remained')
  await expectFailure(() => client.upload({ name: 'revoked.txt', type: 'text/plain', size: 7, body: new Blob(['revoked'], { type: 'text/plain' }) }), 'forbidden')

  await control('revoke-session', {})
  await setSearch('')
  await waitUntil(() => sessionRequired && controller.failure() === 'relogin', 'revoked session did not request relogin')
  assert(controller.list.snapshot().rows.length === 0 && document.querySelector('tbody tr') === null, 'revoked session retained projection')
  app.unmount()
  document.body.innerHTML = '<pre id="result">FILES_E2E_PASS</pre>'
  await control('shutdown', {})
}

await scenario().catch(async (error) => {
  const allowed = ['administrator login failed', 'files capability manifest incomplete', 'self scope not projected', 'self scope leaked foreign metadata', 'foreign fixture did not load', 'search reset failed', 'download bytes differ', 'restart lost metadata', 'restart lost content', 'restart changed durable state', 'missing CSRF was not rejected', 'CSRF rejection changed metadata', 'batch delete did not complete', 'batch delete removed foreign object', 'self scope did not hide foreign row', 'revoked capabilities remained', 'revoked mutation UI remained', 'revoked session did not request relogin', 'revoked session retained projection']
  const message = error instanceof Error && allowed.includes(error.message) ? error.message : 'browser assertion failed'
  document.body.replaceChildren()
  const result = document.createElement('pre')
  result.id = 'result'
  result.textContent = `FILES_E2E_FAIL|ASSERTION:${message}`
  document.body.append(result)
  await control('shutdown', {}).catch(() => undefined)
})
