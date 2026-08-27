import { createApp, h, type Component } from 'vue'
import { createCapabilityController } from '@go-admin/domain-iam/administration'
import { createSessionController } from '@go-admin/domain-iam/session'
import { SchedulerRequestError, type DefinitionInput } from '@go-admin/domain-scheduler'
import { createSchedulerController, createWebSchedulerClient, SchedulerPage } from '@go-admin/web-domain-scheduler'
import { createWebAdministrationClient } from '@go-admin/web-domain-iam/administration'
import { createWebSessionClient } from '@go-admin/web-domain-iam/session'
import { createBrowserSessionFetch } from '@go-admin/adapter-browser'

const assert: (condition: unknown, message: string) => asserts condition = (condition, message) => { if (!condition) throw new Error(message) }
const waitUntil = async (condition: () => boolean, message: string, timeout = 10_000) => { const deadline = Date.now() + timeout; while (Date.now() < deadline) { if (condition()) return; await new Promise(resolve => setTimeout(resolve, 25)) }; throw new Error(message) }
const element = <T extends Element>(selector: string) => { const value = document.querySelector<T>(selector); assert(value, `missing element: ${selector}`); return value }
const input = async (selector: string, value: string | boolean) => { const target = element<HTMLInputElement | HTMLSelectElement>(selector); if (target instanceof HTMLInputElement && target.type === 'checkbox') target.checked = Boolean(value); else target.value = String(value); target.dispatchEvent(new Event(target instanceof HTMLSelectElement || target.type === 'checkbox' ? 'change' : 'input', { bubbles: true })); await Promise.resolve() }
const control = async (path: string, method: 'GET' | 'POST' = 'GET', body?: unknown) => { const response = await fetch(`/__test/${path}`, body === undefined ? { method } : { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) }); assert(response.status === (method === 'GET' ? 200 : path === 'run' ? 200 : 204), `control ${path} failed`); return response }
const definitionRow = (name: string) => [...document.querySelectorAll<HTMLTableRowElement>('[data-row-key]')].find(row => row.cells[0]?.textContent === name)
const action = (name: string, actionName: string) => { const button = definitionRow(name)?.querySelector<HTMLButtonElement>(`[data-action="${actionName}"]`); assert(button, `missing ${actionName} for ${name}`); button.click() }
const expectFailure = async (operation: () => Promise<unknown>, category: string) => { try { await operation() } catch (error) { assert(error instanceof SchedulerRequestError && error.category === category, `expected ${category}`); return }; throw new Error(`expected ${category}`) }

const scenario = async () => {
  const sessionFetch = createBrowserSessionFetch(fetch)
  const session = createSessionController(createWebSessionClient(sessionFetch, '/api'))
  await session.login({ username: 'scheduler-admin', password: 'scheduler administrator password' })
  assert(session.state().status === 'authenticated' && !document.cookie.includes('__Host-go-admin-session'), 'scheduler login or HttpOnly boundary failed')
  const capability = createCapabilityController(createWebAdministrationClient(sessionFetch, '/api'))
  await capability.refresh()
  const api = createWebSchedulerClient(sessionFetch, '/api')
  assert(capability.can('scheduler.definitions.read') && capability.can('scheduler.executions.read'), 'scheduler manifest incomplete')
  await control('scope', 'POST', { scope: 'self' }); await capability.refresh()
  await expectFailure(() => api.taskTypes(), 'forbidden')
  const selfController = createSchedulerController(api, { can: permission => capability.can(permission), scope: () => capability.state().manifest?.dataScope ?? null }, async () => true)
  document.body.innerHTML = '<div id="app"></div>'
  const selfApp = createApp({ render: () => h(SchedulerPage as Component, { controller: selfController }) }); selfApp.mount('#app'); await Promise.resolve(); await Promise.resolve()
  assert(document.querySelector('.tabs button') === null && document.querySelector('.editor') === null, 'self scope exposed scheduler controls')
  selfApp.unmount(); await control('scope', 'POST', { scope: 'all' }); await capability.refresh()

  const controller = createSchedulerController(api, { can: permission => capability.can(permission), scope: () => capability.state().manifest?.dataScope ?? null }, async () => true)
  document.body.innerHTML = '<div id="app"></div>'
  const app = createApp({ render: () => h(SchedulerPage as Component, { controller }) }); app.mount('#app')
  await waitUntil(() => document.querySelector('[data-testid="scheduler-definition-form"]') !== null && controller.taskTypes().length === 1, 'scheduler editor did not load')
  const create = async (name: string, key: string, fail: boolean) => {
    await input('[data-testid="scheduler-definition-form"] [name="name"]', name)
    await input('[data-testid="scheduler-parameters"] [name="key"]', key)
    await input('[data-testid="scheduler-parameters"] [name="fail"]', fail)
    element<HTMLFormElement>('[data-testid="scheduler-definition-form"]').requestSubmit()
    await waitUntil(() => definitionRow(name) !== undefined, `${name} create did not render`)
    action(name, 'toggle')
    await waitUntil(() => definitionRow(name)?.textContent?.includes('Enabled') === true, `${name} enable did not render`)
  }
  await create('Browser success', 'success', false)
  await create('Browser failure', 'failure', true)
  await control('contender', 'POST')
  const run = await (await control('run', 'POST')).json() as Record<string, number>
  assert(run.triggered === 2 && run.succeeded === 1 && run.failed === 1 && run.delivered === 1, 'scheduler/outbox shared lease execution mismatch')
  const snapshot = await (await control('snapshot')).json() as Record<string, number>
  assert(snapshot.definitions === 2 && snapshot.executions === 2 && snapshot.effects === 1 && snapshot.events === 1 && snapshot.outbox === 1, 'scheduler transaction/savepoint snapshot mismatch')
  await controller.executions.refresh()
  const executionsTab = [...document.querySelectorAll<HTMLButtonElement>('.tabs button')].find(button => button.textContent === 'Executions'); assert(executionsTab, 'execution tab missing'); executionsTab.click()
  await waitUntil(() => document.body.textContent?.includes('browser_expected_failure') === true, 'failed execution history missing')
  const definitionsTab = [...document.querySelectorAll<HTMLButtonElement>('.tabs button')].find(button => button.textContent === 'Definitions'); assert(definitionsTab, 'definitions tab missing'); definitionsTab.click(); await Promise.resolve()
  for (const name of ['Browser success', 'Browser failure']) { action(name, 'toggle'); await waitUntil(() => definitionRow(name)?.textContent?.includes('Stopped') === true, `${name} stop did not linearize`) }
  const second = await (await control('run', 'POST')).json() as Record<string, number>
  assert(second.triggered === 0, 'stopped scheduler produced a new execution')
  action('Browser success', 'edit'); await waitUntil(() => document.querySelector('[data-testid="scheduler-definition-form"] [name="name"]') !== null, 'edit did not open'); await input('[data-testid="scheduler-definition-form"] [name="name"]', 'Browser success updated'); element<HTMLFormElement>('[data-testid="scheduler-definition-form"]').requestSubmit(); await waitUntil(() => definitionRow('Browser success updated') !== undefined, 'edit did not render')
  action('Browser failure', 'delete'); await waitUntil(() => definitionRow('Browser failure') === undefined, 'delete did not render')
  const invalid = { name: 'Invalid', taskType: 'browser.effect', schedule: { minutes: [0], hours: [0], daysOfMonth: [], months: [1], weekdays: [] }, parameters: { key: 'invalid', fail: false, extra: 'rejected' } } as unknown as DefinitionInput
  await expectFailure(() => api.createDefinition(invalid), 'validation')
  await control('revoke-read', 'POST'); await capability.refresh()
  assert(controller.definitions.snapshot().rows.length === 0, 'revoked capability retained definition projection')
  app.unmount(); document.body.innerHTML = '<div id="app"></div>'
  const revoked = createSchedulerController(api, { can: permission => capability.can(permission), scope: () => capability.state().manifest?.dataScope ?? null }, async () => true)
  const revokedApp = createApp({ render: () => h(SchedulerPage as Component, { controller: revoked }) }); revokedApp.mount('#app'); await Promise.resolve(); await Promise.resolve()
  assert(![...document.querySelectorAll('.tabs button')].some(button => button.textContent === 'Definitions'), 'revoked scheduler navigation remained visible')
  revokedApp.unmount(); await control('revoke-session', 'POST'); await expectFailure(() => api.taskTypes(), 'relogin')
  document.body.innerHTML = '<pre id="result">SCHEDULER_E2E_PASS</pre>'; await control('shutdown', 'POST')
}

await scenario().catch(async () => { document.body.replaceChildren(); const result = document.createElement('pre'); result.id = 'result'; result.textContent = 'SCHEDULER_E2E_FAIL|ASSERTION'; document.body.append(result); await control('shutdown', 'POST').catch(() => undefined) })
