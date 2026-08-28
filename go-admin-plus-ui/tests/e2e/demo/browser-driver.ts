import { createApp, h, type App, type Component } from 'vue'
import { createDemoController, createWebDemoClient, DemoProductsPage } from '@go-admin-plus/web-domain-demo'
import { createCapabilityController } from '@go-admin-plus/domain-iam/administration'
import { createSessionController } from '@go-admin-plus/domain-iam/session'
import { createWebAdministrationClient } from '@go-admin-plus/web-domain-iam/administration'
import { createWebSessionClient } from '@go-admin-plus/web-domain-iam/session'
import { DemoRequestError } from '@go-admin-plus/domain-demo'

const result = document.querySelector<HTMLElement>('#result')!
const rawClient = createWebDemoClient(fetch, '/api')
let listSettled = 0
const client = { ...rawClient, list: async (...arguments_: Parameters<typeof rawClient.list>) => {
  try { return await rawClient.list(...arguments_) } finally { listSettled += 1 }
} }
let loginCSRFLength = 0
const sessionFetcher: typeof fetch = async input => {
  const request = input instanceof Request ? input : new Request(input)
  const response = await fetch(request)
  if (new URL(request.url).pathname === '/api/iam/session/login') loginCSRFLength = response.headers.get('X-CSRF-Token')?.length ?? 0
  return response
}
const session = createSessionController(createWebSessionClient(sessionFetcher, '/api'))
const capabilities = createCapabilityController(createWebAdministrationClient(fetch, '/api'))
const controller = createDemoController(client, async () => true, { can: code => capabilities.can(code) })
let app: App<Element> | null = null
const mount = () => {
  app?.unmount()
  document.querySelector('#app')!.textContent = ''
  app = createApp({ render: () => h(DemoProductsPage as Component, { controller }) })
  app.mount('#app')
}

const delay = (milliseconds: number) => new Promise(resolve => setTimeout(resolve, milliseconds))
const waitUntil = async (condition: () => boolean, message: string) => {
  const deadline = Date.now() + 15_000
  while (Date.now() < deadline) { if (condition()) return; await delay(25) }
  throw new Error(message)
}
const postControl = async (path: string, body?: unknown) => {
  const response = await fetch(path, { method: 'POST', headers: body === undefined ? undefined : { 'Content-Type': 'application/json' }, body: body === undefined ? undefined : JSON.stringify(body) })
  if (response.status !== 204) throw new Error('test control failed')
}
const element = <T extends Element>(selector: string): T => {
  const value = document.querySelector<T>(selector); if (!value) throw new Error('expected interface element is missing'); return value
}
const input = (name: string, value: string) => {
  const control = element<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>(`[name="${name}"]`)
  control.value = value; control.dispatchEvent(new Event(control instanceof HTMLSelectElement ? 'change' : 'input', { bubbles: true }))
}
const create = async (sku: string, name: string) => {
  element<HTMLButtonElement>('[data-testid="open-product-form"]').click()
  await waitUntil(() => document.querySelector('.demo-products__form') !== null, 'product create form did not open')
  input('sku', sku); input('name', name); input('description', 'Browser tracer'); input('priceCents', '1250'); input('status', 'active')
  element<HTMLButtonElement>('.demo-products__form button[type="submit"]').click()
  await waitUntil(() => controller.list.snapshot().rows.some(row => row.sku === sku), 'created product did not appear')
}

try {
  await session.login({ username: 'admin', password: 'administrator password' })
  if (session.state().status !== 'authenticated' || loginCSRFLength !== 43 || document.cookie.includes('__Host-go-admin-session')) throw new Error('administrator login contract failed')
  await capabilities.refresh()
  if (!capabilities.can('demo.products.read') || !capabilities.can('demo.products.write') || !capabilities.can('demo.products.delete')) throw new Error('demo capability registration failed')
  if (capabilities.state().manifest?.menus.some(menu => menu.path === '/demo/products' && menu.permissionCode === 'demo.products.read') !== true) throw new Error('demo navigation capability missing')
  mount()
  await waitUntil(() => document.querySelector('#demo-products-title') !== null && listSettled >= 1 && !controller.list.snapshot().loading, 'product page did not load')
  await postControl('/__test/scope', { scope: 'self' })
  await controller.list.refresh()
  if (controller.list.snapshot().total !== 0) throw new Error('self scope exposed foreign product')
  await postControl('/__test/scope', { scope: 'all' })
  await controller.list.refresh()
  if (!controller.list.snapshot().rows.some(row => row.sku === 'FOREIGN-01')) throw new Error('all scope omitted foreign product')
  await create('TRACE-01', 'Tracer product one')
  await postControl('/__test/restart')
  const listBeforeRestart = listSettled
  element<HTMLButtonElement>('.demo-products__search button[type="submit"]').click()
  await waitUntil(() => listSettled > listBeforeRestart && controller.list.snapshot().rows.some(row => row.sku === 'TRACE-01'), 'product did not survive database restart')
  element<HTMLButtonElement>('tbody tr button').click()
  await waitUntil(() => element<HTMLInputElement>('[name="name"]').value === 'Tracer product one', 'edit form did not open')
  input('name', 'Tracer product updated'); element<HTMLButtonElement>('.demo-products__form button[type="submit"]').click()
  await waitUntil(() => controller.list.snapshot().rows.some(row => row.name === 'Tracer product updated'), 'updated product did not appear')
  await create('TRACE-02', 'Tracer product two')
  const selections = [...document.querySelectorAll<HTMLInputElement>('tbody input[type="checkbox"]')]
  for (const selection of selections) { selection.click(); await Promise.resolve() }
  const batch = element<HTMLButtonElement>('[data-testid="delete-selected-products"]'); await waitUntil(() => !batch.disabled, 'batch delete did not enable'); batch.click()
  await waitUntil(() => !controller.list.snapshot().rows.some(row => row.sku.startsWith('TRACE-')), 'batch delete did not refresh')
  const csrfBaseline = controller.list.snapshot().total
  const csrfRejected = await fetch('/api/demo/products', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ sku: 'CSRF-01', name: 'CSRF rejected', description: '', priceCents: 1, status: 'active' }) })
  const csrfProblem = await csrfRejected.json() as { code?: string }
  if (csrfRejected.status !== 403 || csrfProblem.code !== 'CSRF_REJECTED') throw new Error('missing csrf was not rejected')
  await controller.list.refresh()
  if (controller.list.snapshot().total !== csrfBaseline) throw new Error('csrf rejection changed state')
  await postControl('/__test/permissions', { enabled: false })
  await capabilities.refresh()
  mount()
  await waitUntil(() => document.querySelector('.demo-products__form') === null && document.querySelector('.demo-products__actions') === null, 'revoked controls remained visible')
  const beforeDenied = controller.list.snapshot().total
  try { await rawClient.create({ sku: 'DENIED-01', name: 'Denied product', description: '', priceCents: 1, status: 'active' }); throw new Error('revoked direct write succeeded') } catch (error) {
    if (!(error instanceof DemoRequestError) || error.category !== 'forbidden') throw error
  }
  await controller.list.refresh()
  if (controller.list.snapshot().total !== beforeDenied) throw new Error('revoked write changed state')
  await postControl('/__test/permissions', { enabled: true })
  await capabilities.refresh()
  await postControl('/__test/revoke-session')
  try { await rawClient.list({ search: '', page: 1, pageSize: 20, sort: 'updatedAt', direction: 'descending' }); throw new Error('revoked session remained active') } catch (error) {
    if (!(error instanceof DemoRequestError) || error.category !== 'relogin') throw error
  }
  result.textContent = 'DEMO_E2E_PASS'
} catch (error) {
  const message = error instanceof Error && /^[a-zA-Z0-9 .,:'-]{1,160}$/.test(error.message) ? error.message : 'browser assertion failed'
  result.textContent = `DEMO_E2E_FAIL|${message}`
}
