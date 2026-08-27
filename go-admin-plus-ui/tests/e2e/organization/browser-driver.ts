import { createApp, h, type Component } from 'vue'
import { createCapabilityController } from '@go-admin/domain-iam/administration'
import { createSessionController } from '@go-admin/domain-iam/session'
import { OrganizationRequestError } from '@go-admin/domain-organization'
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
const control = async (path: string, method: 'GET' | 'POST' = 'GET', body?: unknown) => {
  const response = body === undefined
    ? await fetch(`/__test/${path}`, { method })
    : await fetch(`/__test/${path}`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
  assert(response.status === (method === 'GET' ? 200 : 204), 'test control failed')
  return response
}

const expectOrganizationFailure = async (operation: () => Promise<unknown>, category: 'conflict' | 'forbidden' | 'relogin') => {
  try {
    await operation()
  } catch (error) {
    assert(error instanceof OrganizationRequestError && error.category === category, `expected ${category} organization failure`)
    return
  }
  throw new Error(`expected ${category} organization failure`)
}

const scenario = async () => {
  let loginCSRFLength = 0
  const sessionFetcher: typeof fetch = async (input) => {
    const request = input instanceof Request ? input : new Request(input)
    const response = await fetch(request)
    if (new URL(request.url).pathname === '/api/iam/session/login') loginCSRFLength = response.headers.get('X-CSRF-Token')?.length ?? 0
    return response
  }
  const session = createSessionController(createWebSessionClient(sessionFetcher, '/api'))
  await session.login({ username: 'organization-admin', password: 'organization administrator password' })
  assert(session.state().status === 'authenticated' && loginCSRFLength === 43, 'organization administrator login failed')
  assert(!document.cookie.includes('__Host-go-admin-session'), 'HttpOnly session became script-readable')
  const attributes = await (await control('cookie-attributes')).json() as Record<string, boolean>
  assert(attributes.secure && attributes.httpOnly && attributes.strict, 'host cookie attributes failed')

  const api = createWebOrganizationClient(fetch, '/api')
  const capability = createCapabilityController(createWebAdministrationClient(fetch, '/api'))
  await capability.refresh()
  assert(capability.can('organization.departments.read') && capability.can('organization.positions.read'), 'organization capability manifest is incomplete')

  const beforeSelf = await (await control('snapshot')).json() as { departments: number; positions: number }
  await control('scope', 'POST', { scope: 'self' })
  await capability.refresh()
  assert(capability.state().manifest?.dataScope === 'self', 'self scope was not projected')
  await expectOrganizationFailure(() => api.listDepartments(), 'forbidden')
  await expectOrganizationFailure(() => api.createPosition({ key: 'self-denied', name: 'Self denied', departmentId: 'department-root-001', enabled: true }), 'forbidden')
  document.body.innerHTML = '<div id="app"></div>'
  const selfController = createOrganizationController(api, { can: (permission) => capability.can(permission), scope: () => capability.state().manifest?.dataScope ?? null }, async () => true)
  const selfApp = createApp({ render: () => h(OrganizationPage as Component, { controller: selfController }) })
  selfApp.mount('#app')
  await Promise.resolve()
  await Promise.resolve()
  assert(document.querySelector('.tabs button') === null && document.querySelector('.editor') === null, 'self scope exposed organization management controls')
  const afterSelf = await (await control('snapshot')).json() as { departments: number; positions: number }
  assert(afterSelf.departments === beforeSelf.departments && afterSelf.positions === beforeSelf.positions, 'self-scoped requests changed organization state')
  selfApp.unmount()
  await control('scope', 'POST', { scope: 'all' })
  await capability.refresh()
  assert(capability.state().manifest?.dataScope === 'all', 'all scope did not recover')

  const controller = createOrganizationController(api, { can: (permission) => capability.can(permission), scope: () => capability.state().manifest?.dataScope ?? null }, async () => true)
  document.body.innerHTML = '<div id="app"></div>'
  const app = createApp({ render: () => h(OrganizationPage as Component, { controller }) })
  app.mount('#app')
  await waitUntil(() => row('root') !== null, 'department root did not render')

  await input('[data-testid="create-department"] [name="key"]', 'browser-operations')
  await input('[data-testid="create-department"] [name="name"]', 'Browser Operations')
  await input('[data-testid="create-department"] [name="parentId"]', 'department-root-001')
  submit('create-department')
  await waitUntil(() => row('browser-operations') !== null, 'department create did not render')

  const browserOperations = controller.departments().find((item) => item.key === 'browser-operations')
  assert(browserOperations, 'created department projection missing')
  await input('[data-testid="create-department"] [name="key"]', 'browser-team')
  await input('[data-testid="create-department"] [name="name"]', 'Browser Team')
  await input('[data-testid="create-department"] [name="parentId"]', browserOperations.id)
  submit('create-department')
  await waitUntil(() => row('browser-team') !== null, 'child department create did not render')

  rowAction('browser-operations', 'edit')
  await waitUntil(() => document.querySelector('[data-testid="edit-department"]') !== null, 'department editor did not render')
  await input('[data-testid="edit-department"] input', 'Browser Operations Updated')
  submit('edit-department')
  await waitUntil(() => row('browser-operations')?.textContent?.includes('Browser Operations Updated') === true, 'department edit did not render')
  const browserTeam = controller.departments().find((item) => item.key === 'browser-team')
  assert(browserTeam, 'child department projection missing')
  await input('[data-testid="edit-department"] select', browserTeam.id)
  submit('edit-department')
  await waitUntil(() => document.querySelector('[role="alert"]')?.textContent?.includes('protected or still referenced') === true, 'department cycle conflict was not visible')
  assert(controller.departments().find((item) => item.key === 'browser-operations')?.parentId === 'department-root-001', 'department cycle changed state')

  await openTab('Positions', 'positions-heading')
  const createPosition = async (key: string, name: string) => {
    await input('[data-testid="create-position"] [name="key"]', key)
    await input('[data-testid="create-position"] [name="name"]', name)
    await input('[data-testid="create-position"] [name="departmentId"]', browserOperations.id)
    submit('create-position')
    await waitUntil(() => row(key) !== null, `position ${key} create did not render`)
  }
  await createPosition('browser-lead', 'Browser Lead')
  await createPosition('browser-ascii', '<:@ collision')
  await createPosition('browser-percent', '% literal')
  await createPosition('browser-under', '_ literal')
  await createPosition('browser-unicode', 'ä collision')

  rowAction('browser-lead', 'edit')
  await waitUntil(() => document.querySelector('[data-testid="edit-position"]') !== null, 'position editor did not render')
  await input('[data-testid="edit-position"] input', 'Browser Lead Updated')
  submit('edit-position')
  await waitUntil(() => row('browser-lead')?.textContent?.includes('Browser Lead Updated') === true, 'position edit did not render')

  for (const [search, expected] of [['%', 'browser-percent'], ['_', 'browser-under'], ['ä', 'browser-unicode']] as const) {
    await input('[data-testid="position-search"] [name="search"]', search)
    submit('position-search')
    await waitUntil(() => row(expected) !== null && document.querySelectorAll('[data-row-key]').length === 1, `literal search ${expected} was not exact`)
  }
  await input('[data-testid="position-search"] [name="search"]', '<:@')
  submit('position-search')
  await waitUntil(() => row('browser-ascii') !== null && row('browser-unicode') === null, 'Unicode byte-boundary collision was not isolated')
  element<HTMLButtonElement>('[data-testid="position-search"] button[type="button"]').click()
  await waitUntil(() => ['browser-lead', 'browser-ascii', 'browser-percent', 'browser-under', 'browser-unicode'].every((key) => row(key) !== null), 'position search reset did not restore projection')

  await openTab('Departments', 'departments-heading')
  rowAction('browser-operations', 'delete')
  await waitUntil(() => document.querySelector('[role="alert"]')?.textContent?.includes('referenced') === true, 'referenced department conflict was not visible')
  assert(row('browser-operations') !== null, 'referenced department was deleted')
  await expectOrganizationFailure(() => api.deleteDepartment('department-root-001'), 'conflict')

  await openTab('Positions', 'positions-heading')
  for (const key of ['browser-lead', 'browser-ascii', 'browser-percent', 'browser-under', 'browser-unicode']) {
    rowAction(key, 'delete')
    await waitUntil(() => row(key) === null, `position ${key} delete did not render`)
  }
  await openTab('Departments', 'departments-heading')
  rowAction('browser-team', 'delete')
  await waitUntil(() => row('browser-team') === null, 'child department delete did not render')
  rowAction('browser-operations', 'delete')
  await waitUntil(() => row('browser-operations') === null, 'department delete did not render')

  const before = await (await control('snapshot')).json() as { departments: number; positions: number }
  assert(before.departments === 1 && before.positions === 0, 'cleanup did not restore initial organization state')

  const csrfRejected = await fetch('/api/organization/positions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key: 'csrf-denied', name: 'CSRF denied', departmentId: 'department-root-001', enabled: true }),
  })
  const csrfProblem = await csrfRejected.json() as { code?: string }
  assert(csrfRejected.status === 403 && csrfProblem.code === 'CSRF_REJECTED', 'missing CSRF was not rejected')
  const afterCSRF = await (await control('snapshot')).json() as { departments: number; positions: number }
  assert(afterCSRF.departments === before.departments && afterCSRF.positions === before.positions, 'CSRF rejection changed organization state')

  await control('revoke-position-read', 'POST')
  await expectOrganizationFailure(() => api.listPositions('', 1, 20), 'forbidden')

  await capability.refresh()
  assert(!capability.can('organization.positions.read'), 'revoked capability remained visible')
  app.unmount()
  document.body.innerHTML = '<div id="app"></div>'
  const revokedController = createOrganizationController(api, { can: (permission) => capability.can(permission), scope: () => capability.state().manifest?.dataScope ?? null }, async () => true)
  const revokedApp = createApp({ render: () => h(OrganizationPage as Component, { controller: revokedController }) })
  revokedApp.mount('#app')
  await waitUntil(() => row('root') !== null, 'authorized department projection did not recover')
  assert(![...document.querySelectorAll('.tabs button')].some((item) => item.textContent === 'Positions'), 'revoked organization navigation remained visible')

  revokedApp.unmount()
  await control('revoke-session', 'POST')
  await expectOrganizationFailure(() => api.listDepartments(), 'relogin')
  document.body.innerHTML = '<div id="app"></div>'
  let sessionRequired = false
  const expiredController = createOrganizationController(api, { can: (permission) => capability.can(permission), scope: () => capability.state().manifest?.dataScope ?? null }, async () => true)
  const expiredApp = createApp({ render: () => h(OrganizationPage as Component, { controller: expiredController, onSessionRequired: () => { sessionRequired = true } }) })
  expiredApp.mount('#app')
  await waitUntil(() => sessionRequired && expiredController.failure() === 'relogin', 'revoked session did not enter relogin state')
  assert(document.querySelector('[data-row-key]') === null && document.querySelector('[role="alert"]')?.textContent?.includes('renewed') === true, 'revoked session kept organization data visible')
  await capability.refresh()
  assert(capability.state().manifest === null && capability.state().status === 'unauthorized', 'revoked session retained capability manifest')
  expiredApp.unmount()
  document.body.innerHTML = '<div id="app"></div>'
  const unauthorizedController = createOrganizationController(api, { can: (permission) => capability.can(permission), scope: () => capability.state().manifest?.dataScope ?? null }, async () => true)
  const unauthorizedApp = createApp({ render: () => h(OrganizationPage as Component, { controller: unauthorizedController }) })
  unauthorizedApp.mount('#app')
  await Promise.resolve()
  assert(document.querySelector('.tabs button') === null, 'revoked session retained organization navigation')
  unauthorizedApp.unmount()
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
