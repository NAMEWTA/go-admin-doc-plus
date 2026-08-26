import { createApp, h, type Component } from 'vue'
import { createCapabilityController } from '@go-admin/domain-iam/administration'
import { createSessionController } from '@go-admin/domain-iam/session'
import { createOrganizationController, createWebOrganizationClient, OrganizationPage } from '@go-admin/web-domain-organization'
import { createWebAdministrationClient } from '@go-admin/web-domain-iam/administration'
import { createWebSessionClient } from '@go-admin/web-domain-iam/session'

const assert: (condition: unknown, message: string) => asserts condition = (condition, message) => { if (!condition) throw new Error(message) }
const waitUntil = async (condition: () => boolean, message: string, timeout = 10_000) => {
  const deadline = Date.now() + timeout
  while (Date.now() < deadline) {
    if (condition()) return
    await new Promise((resolve) => setTimeout(resolve, 25))
  }
  throw new Error(message)
}
const element = <T extends Element>(selector: string) => {
  const value = document.querySelector<T>(selector)
  assert(value, `missing element: ${selector}`)
  return value
}
const input = async (selector: string, value: string | boolean) => {
  const target = element<HTMLInputElement | HTMLSelectElement>(selector)
  if (target instanceof HTMLInputElement && target.type === 'checkbox') target.checked = Boolean(value)
  else target.value = String(value)
  target.dispatchEvent(new Event(target instanceof HTMLSelectElement || target.type === 'checkbox' ? 'change' : 'input', { bubbles: true }))
  await Promise.resolve()
}
const submit = (testID: string) => element<HTMLFormElement>(`[data-testid="${testID}"]`).requestSubmit()
const row = (key: string) => document.querySelector<HTMLElement>(`[data-row-key="${key}"]`)
const rowAction = (key: string, action: string) => {
  const button = row(key)?.querySelector<HTMLButtonElement>(`[data-action="${action}"]`)
  assert(button, `missing ${action} for ${key}`)
  button.click()
}
const openTab = async (label: string, heading: string) => {
  const button = [...document.querySelectorAll<HTMLButtonElement>('.tabs button')].find((candidate) => candidate.textContent === label)
  assert(button, `missing tab: ${label}`)
  button.click()
  await waitUntil(() => document.querySelector(`#${heading}`) !== null, `${label} tab did not render`)
}
const control = async (path: string, method = 'GET') => {
  const response = await fetch(`/__test/${path}`, { method })
  assert(response.ok, 'test control failed')
  return response
}

const scenario = async () => {
  const session = createSessionController(createWebSessionClient(fetch, '/api'))
  await session.login({ username: 'organization-admin', password: 'organization administrator password' })
  assert(session.state().status === 'authenticated', 'organization administrator login failed')
  assert(!document.cookie.includes('__Host-go-admin-session'), 'HttpOnly session became script-readable')
  const attributes = await (await control('cookie-attributes')).json() as Record<string, boolean>
  assert(attributes.secure && attributes.httpOnly && attributes.strict, 'host cookie attributes failed')

  const api = createWebOrganizationClient(fetch, '/api')
  const capability = createCapabilityController(createWebAdministrationClient(fetch, '/api'))
  await capability.refresh()
  assert(capability.can('organization.departments.read') && capability.can('organization.positions.read'), 'organization capability manifest is incomplete')
  const controller = createOrganizationController(api, (permission) => capability.can(permission), async () => true)
  document.body.innerHTML = '<div id="app"></div>'
  const app = createApp({ render: () => h(OrganizationPage as Component, { controller }) })
  app.mount('#app')
  await waitUntil(() => row('root') !== null, 'department root did not render')

  await input('[data-testid="create-department"] [name="key"]', 'browser-operations')
  await input('[data-testid="create-department"] [name="name"]', 'Browser Operations')
  await input('[data-testid="create-department"] [name="parentId"]', 'department-root-001')
  submit('create-department')
  await waitUntil(() => row('browser-operations') !== null, 'department create did not render')

  await openTab('Positions', 'positions-heading')
  const department = controller.departments().find((item) => item.key === 'browser-operations')
  assert(department, 'created department projection missing')
  await input('[data-testid="create-position"] [name="key"]', 'browser-lead')
  await input('[data-testid="create-position"] [name="name"]', 'Browser Lead')
  await input('[data-testid="create-position"] [name="departmentId"]', department.id)
  submit('create-position')
  await waitUntil(() => row('browser-lead') !== null, 'position create did not render')

  await openTab('Departments', 'departments-heading')
  rowAction('browser-operations', 'delete')
  await waitUntil(() => document.querySelector('[role="alert"]')?.textContent?.includes('referenced') === true, 'referenced department conflict was not visible')
  assert(row('browser-operations') !== null, 'referenced department was deleted')
  let protectedDenied = false
  try { await api.deleteDepartment('department-root-001') } catch { protectedDenied = true }
  assert(protectedDenied, 'protected root direct API deletion succeeded')

  await openTab('Positions', 'positions-heading')
  rowAction('browser-lead', 'delete')
  await waitUntil(() => row('browser-lead') === null, 'position delete did not render')
  await openTab('Departments', 'departments-heading')
  rowAction('browser-operations', 'delete')
  await waitUntil(() => row('browser-operations') === null, 'department delete did not render')

  const before = await (await control('snapshot')).json() as { departments: number; positions: number }
  assert(before.departments === 1 && before.positions === 0, 'cleanup did not restore initial organization state')
  await control('revoke-position-read', 'POST')
  let revoked = false
  try { await api.listPositions('', 1, 20) } catch { revoked = true }
  assert(revoked, 'permission revocation was not immediate')

  await capability.refresh()
  assert(!capability.can('organization.positions.read'), 'revoked capability remained visible')
  app.unmount()
  document.body.innerHTML = '<div id="app"></div>'
  const revokedController = createOrganizationController(api, (permission) => capability.can(permission), async () => true)
  const revokedApp = createApp({ render: () => h(OrganizationPage as Component, { controller: revokedController }) })
  revokedApp.mount('#app')
  await waitUntil(() => row('root') !== null, 'authorized department projection did not recover')
  assert(![...document.querySelectorAll('.tabs button')].some((item) => item.textContent === 'Positions'), 'revoked organization navigation remained visible')

  revokedApp.unmount()
  document.body.innerHTML = '<pre id="result">ORGANIZATION_E2E_PASS</pre>'
  await control('shutdown', 'POST')
}

await scenario().catch(async () => {
  document.body.replaceChildren()
  const result = document.createElement('pre')
  result.id = 'result'
  result.textContent = 'ORGANIZATION_E2E_FAIL|ASSERTION'
  document.body.append(result)
  await control('shutdown', 'POST').catch(() => undefined)
})
