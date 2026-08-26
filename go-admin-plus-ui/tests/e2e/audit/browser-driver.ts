import { createAuditController, createWebAuditClient, mountAuditPage } from '@go-admin/web-domain-audit'

const assert: (condition: unknown, message: string) => asserts condition = (condition, message) => {
  if (!condition) throw new Error(message)
}

let confirmCleanup = false
const controller = createAuditController(createWebAuditClient(fetch, '/api'), async () => confirmCleanup)
mountAuditPage('#app', controller)

const waitFor = async (predicate: () => boolean, message: string) => {
  for (let attempt = 0; attempt < 200; attempt += 1) {
    if (predicate()) return
    await new Promise((resolvePromise) => setTimeout(resolvePromise, 25))
  }
  throw new Error(message)
}

const select = (testID: string, value: string) => {
  const element = document.querySelector<HTMLSelectElement>(`[data-testid="${testID}"]`)
  assert(element, `${testID} is unavailable`)
  element.value = value
  element.dispatchEvent(new Event('change', { bubbles: true }))
}

const click = (testID: string) => {
  const element = document.querySelector<HTMLButtonElement>(`[data-testid="${testID}"]`)
  assert(element, `${testID} is unavailable`)
  element.click()
}

const snapshot = async (): Promise<{ count: number }> => {
  const response = await fetch('/__test/snapshot')
  assert(response.ok, 'Audit browser snapshot failed')
  return response.json() as Promise<{ count: number }>
}

const driver = {
  async run() {
    await waitFor(() => document.querySelectorAll('[data-testid="audit-row"]').length === 1, 'initial Audit page load failed')
    select('audit-source', 'web')
    select('audit-action', 'update')
    click('audit-search')
    await waitFor(() => controller.list.snapshot().filters.source === 'web' && controller.list.snapshot().total === 1, 'Audit browser filter failed')

    click('audit-view')
    await waitFor(() => Boolean(document.querySelector('dialog[open]')), 'Audit browser detail did not open')
    const detail = document.querySelector('dialog')?.textContent ?? ''
    assert(detail.includes('demo:ui-record') && detail.includes('Succeeded'), 'Audit browser detail content failed')

    const before = document.querySelector<HTMLInputElement>('[data-testid="audit-cleanup-before"]')
    assert(before, 'Audit cleanup boundary is unavailable')
    before.value = '2026-06-01'
    before.dispatchEvent(new Event('input', { bubbles: true }))
		await new Promise((resolvePromise) => setTimeout(resolvePromise, 0))

    confirmCleanup = false
    click('audit-cleanup')
    await new Promise((resolvePromise) => setTimeout(resolvePromise, 50))
    assert((await snapshot()).count === 1, 'cancelled Audit cleanup changed state')

    confirmCleanup = true
    click('audit-cleanup')
    await waitFor(() => (document.querySelector('[data-testid="audit-cleanup-status"]')?.textContent ?? '').includes('Deleted 1 records'), 'Audit cleanup status failed')
    assert(document.querySelectorAll('[data-testid="audit-row"]').length === 0, 'Audit list did not refresh after cleanup')
    assert((await snapshot()).count === 0, 'Audit browser cleanup did not persist')
    return true
  },
  async shutdown() {
    await fetch('/__test/shutdown', { method: 'POST' })
    return true
  },
}

declare global {
  interface Window { __auditE2E: typeof driver }
}

window.__auditE2E = driver
