import { createApp, h, type Component } from 'vue'
import { createDemoController, createWebDemoClient, DemoProductsPage } from '@go-admin/web-domain-demo'

const result = document.querySelector<HTMLElement>('#result')!
const rawClient = createWebDemoClient(fetch, '/api')
let listSettled = 0
const client = { ...rawClient, list: async (...arguments_: Parameters<typeof rawClient.list>) => {
  try { return await rawClient.list(...arguments_) } finally { listSettled += 1 }
} }
const controller = createDemoController(client, async () => true)
createApp({ render: () => h(DemoProductsPage as Component, { controller }) }).mount('#app')

const delay = (milliseconds: number) => new Promise(resolve => setTimeout(resolve, milliseconds))
const waitUntil = async (condition: () => boolean, message: string) => {
  const deadline = Date.now() + 15_000
  while (Date.now() < deadline) { if (condition()) return; await delay(25) }
  throw new Error(message)
}
const element = <T extends Element>(selector: string): T => {
  const value = document.querySelector<T>(selector); if (!value) throw new Error('expected interface element is missing'); return value
}
const input = (name: string, value: string) => {
  const control = element<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>(`[name="${name}"]`)
  control.value = value; control.dispatchEvent(new Event(control instanceof HTMLSelectElement ? 'change' : 'input', { bubbles: true }))
}
const create = async (sku: string, name: string) => {
  input('sku', sku); input('name', name); input('description', 'Browser tracer'); input('priceCents', '1250'); input('status', 'active')
  element<HTMLButtonElement>('.demo-products__form button[type="submit"]').click()
  await waitUntil(() => controller.list.snapshot().rows.some(row => row.sku === sku), 'created product did not appear')
}

try {
  await waitUntil(() => document.querySelector('#demo-products-title') !== null && listSettled >= 1 && !controller.list.snapshot().loading, 'product page did not load')
  await create('TRACE-01', 'Tracer product one')
  await fetch('/__test/restart', { method: 'POST' })
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
  const batch = element<HTMLButtonElement>('.demo-products__actions button'); await waitUntil(() => !batch.disabled, 'batch delete did not enable'); batch.click()
  await waitUntil(() => controller.list.snapshot().total === 0, 'batch delete did not refresh')
  result.textContent = 'DEMO_E2E_PASS'
} catch (error) {
  const message = error instanceof Error && /^[a-zA-Z0-9 .,:'-]{1,160}$/.test(error.message) ? error.message : 'browser assertion failed'
  result.textContent = `DEMO_E2E_FAIL|${message}`
}
