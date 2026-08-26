import { createApp, h, type App, type Component } from 'vue'
import { createSessionController } from '@go-admin/domain-iam/session'
import { createAdministrationController, createWebAdministrationClient, AdministrationPage, type AdministrationController } from '@go-admin/web-domain-iam/administration'
import { createWebSessionClient } from '@go-admin/web-domain-iam/session'
import { administrationMountDiagnostic, safeBrowserDiagnostic } from './diagnostics.mjs'

const assert: (condition: unknown, message: string) => asserts condition = (condition, message) => { if (!condition) throw new Error(message) }
const renderFailure = (error: unknown) => {
  const diagnostic = safeBrowserDiagnostic(error)
  const marker = `IAM_ADMIN_E2E_FAIL|ASSERTION|${diagnostic}`
  document.body.replaceChildren()
  const result = document.createElement('pre')
  result.id = 'result'
  result.textContent = marker
  document.body.append(result)
  console.error(`IAM_ADMIN_E2E_DIAGNOSTIC|ASSERTION|${diagnostic}`)
}
const session = createSessionController(createWebSessionClient(fetch, '/api'))
const control = async (path: string, method = 'GET') => { const response = await fetch(`/__test/${path}`, { method }); assert(response.ok, 'test control failed'); return response }
const waitUntil = async (condition: () => boolean, message: string | (() => string), timeout = 10_000) => {
  const deadline = Date.now() + timeout
  while (Date.now() < deadline) {
    if (condition()) return
    await new Promise((resolve) => setTimeout(resolve, 25))
  }
  throw new Error(typeof message === 'function' ? message() : message)
}
const element = <T extends Element>(selector: string) => {
  const result = document.querySelector<T>(selector)
  assert(result, `missing element: ${selector}`)
  return result
}
const row = (key: string) => document.querySelector<HTMLTableRowElement>(`[data-row-key="${key}"]`)
const input = (selector: string, value: string | boolean) => {
  const target = element<HTMLInputElement | HTMLSelectElement>(selector)
  if (target instanceof HTMLInputElement && target.type === 'checkbox') target.checked = Boolean(value)
  else target.value = String(value)
  target.dispatchEvent(new Event(target instanceof HTMLSelectElement || target.type === 'checkbox' ? 'change' : 'input', { bubbles: true }))
}
const click = (selector: string) => element<HTMLElement>(selector).click()
const clickRow = (key: string, action: string) => {
  const target = row(key)?.querySelector<HTMLButtonElement>(`[data-action="${action}"]`)
  assert(target, `missing ${action} for ${key}`)
  target.click()
}
const submit = (testID: string) => element<HTMLFormElement>(`[data-testid="${testID}"]`).requestSubmit()
const openTab = (name: string) => {
  const button = [...document.querySelectorAll<HTMLButtonElement>('.tabs button')].find((candidate) => candidate.textContent === name)
  assert(button, `missing tab: ${name}`)
  button.click()
}

interface MountedAdministration { app: App; controller: AdministrationController; api: ReturnType<typeof createWebAdministrationClient>; confirmations(): number }
const mountAdministration = async (expectedUser = 'admin'): Promise<MountedAdministration> => {
  document.body.innerHTML = '<div id="app"></div>'
  const api = createWebAdministrationClient(fetch, '/api')
  let confirmations = 0
  const controller = createAdministrationController(api, async () => { confirmations += 1; return true })
  const app = createApp({ render: () => h(AdministrationPage as Component, { controller }) })
  app.mount('#app')
  await waitUntil(() => row(expectedUser) !== null, () => {
    const snapshot = controller.users.snapshot()
    return administrationMountDiagnostic({
      failure: controller.failure(),
      canUsersRead: controller.can('iam.users.read'),
      rows: snapshot.rows.length,
      total: snapshot.total,
      loading: snapshot.loading,
      alertText: document.querySelector('[role="alert"]')?.textContent?.trim() ?? null,
    })
  })
  return { app, controller, api, confirmations: () => confirmations }
}

const fillCreateUser = async (username: string, displayName: string) => {
  input('[data-testid="create-user"] [name="username"]', username)
  input('[data-testid="create-user"] [name="displayName"]', displayName)
  input('[data-testid="create-user"] [name="email"]', `${username}@example.test`)
  input('[data-testid="create-user"] [name="password"]', `${username} browser password`)
  submit('create-user')
  await waitUntil(() => row(username) !== null, `user ${username} was not rendered`)
  assert(element<HTMLInputElement>('[data-testid="create-user"] [name="password"]').value === '', 'create password was retained')
}

const scenario = async () => {
  await session.login({ username: 'admin', password: 'administrator password' })
  assert(session.state().status === 'authenticated', 'administrator login failed')
  assert(!document.cookie.includes('__Host-go-admin-session'), 'HttpOnly session became script-readable')
  const attributes = await (await control('cookie-attributes')).json() as Record<string, boolean>
  assert(attributes.secure && attributes.httpOnly && attributes.strict, 'host cookie attributes failed')
  let mounted = await mountAdministration()

  input('[data-testid="user-search"] input', 'admin')
  submit('user-search')
  await waitUntil(() => row('admin') !== null, 'search result missing')
  click('[data-testid="user-search-reset"]')
  await waitUntil(() => element<HTMLInputElement>('[data-testid="user-search"] input').value === '', 'search reset failed')

  openTab('roles')
  input('[data-testid="create-role"] [name="key"]', 'browser-reader')
  input('[data-testid="create-role"] [name="name"]', 'Browser Reader')
  input('[data-testid="create-role"] [name="dataScope"]', 'self')
  submit('create-role')
  await waitUntil(() => row('browser-reader') !== null, 'role create was not rendered')
  clickRow('browser-reader', 'edit')
  input('[data-testid="edit-role"] [name="name"]', 'Browser Reader Updated')
  submit('edit-role')
  await waitUntil(() => row('browser-reader')?.textContent?.includes('Browser Reader Updated') === true, 'role edit was not rendered')
  input('[data-testid="edit-role"] [name="enabled"]', false)
  submit('edit-role')
  await waitUntil(() => row('browser-reader')?.textContent?.includes('Disabled') === true, 'role disable was not rendered')
  input('[data-testid="edit-role"] [name="enabled"]', true)
  submit('edit-role')
  await waitUntil(() => row('browser-reader')?.textContent?.includes('Enabled') === true, 'role enable was not rendered')

  openTab('menus')
  input('[data-testid="create-menu"] [name="key"]', 'browser-menu')
  input('[data-testid="create-menu"] [name="label"]', 'Browser Menu')
  input('[data-testid="create-menu"] [name="path"]', '/iam/browser')
  input('[data-testid="create-menu"] [name="permissionCode"]', 'iam.users.read')
  input('[data-testid="create-menu"] [name="sortOrder"]', '90')
  submit('create-menu')
  await waitUntil(() => row('browser-menu') !== null, 'menu create was not rendered')
  clickRow('browser-menu', 'edit')
  input('[data-testid="edit-menu"] [name="label"]', 'Browser Menu Updated')
  submit('edit-menu')
  await waitUntil(() => row('browser-menu')?.textContent?.includes('Browser Menu Updated') === true, 'menu edit was not rendered')

  openTab('roles')
  clickRow('browser-reader', 'edit')
  input('[data-testid="assign-role-grants"] [data-permission-code="iam.users.read"]', true)
  input('[data-testid="assign-role-grants"] [data-permission-code="iam.manifest.read"]', true)
  input('[data-testid="assign-role-grants"] [data-menu-key="browser-menu"]', true)
  submit('assign-role-grants')
  await waitUntil(() => mounted.controller.roles().find((value) => value.key === 'browser-reader')?.permissionCodes.includes('iam.manifest.read') === true, 'grants did not refresh')

  openTab('users')
  await fillCreateUser('browser-user', 'Browser User')
  input('[data-testid="create-user"] [name="username"]', 'browser-user')
  input('[data-testid="create-user"] [name="displayName"]', 'Duplicate Browser User')
  input('[data-testid="create-user"] [name="email"]', 'browser-user@example.test')
  input('[data-testid="create-user"] [name="password"]', 'duplicate browser password')
  submit('create-user')
  await waitUntil(() => document.querySelector('[role="alert"]')?.textContent?.includes('protected') === true, 'conflict page state was not visible')
  clickRow('browser-user', 'edit')
  input('[data-testid="edit-user"] [name="displayName"]', 'Browser Updated')
  submit('edit-user')
  await waitUntil(() => row('browser-user')?.textContent?.includes('Browser Updated') === true, 'user edit was not rendered')
  input('[data-testid="assign-user-roles"] [data-role-key="browser-reader"]', true)
  submit('assign-user-roles')
  await waitUntil(() => mounted.controller.users.snapshot().rows.find((value) => value.username === 'browser-user')?.roleIds.length === 1, 'role assignment did not refresh')
  clickRow('browser-user', 'toggle')
  await waitUntil(() => row('browser-user')?.textContent?.includes('Disabled') === true, 'user disable was not rendered')
  clickRow('browser-user', 'toggle')
  await waitUntil(() => row('browser-user')?.textContent?.includes('Enabled') === true, 'user enable was not rendered')
  clickRow('browser-user', 'edit')
  input('[data-testid="reset-user-password"] [name="password"]', 'browser replacement password')
  submit('reset-user-password')
  await waitUntil(() => element<HTMLInputElement>('[data-testid="reset-user-password"] [name="password"]').value === '', 'reset password was retained')

  await fillCreateUser('browser-single', 'Browser Single')
  clickRow('browser-single', 'delete')
  await waitUntil(() => row('browser-single') === null, 'single delete was not rendered')
  await fillCreateUser('browser-batch-a', 'Browser Batch A')
  await fillCreateUser('browser-batch-b', 'Browser Batch B')
  input('[aria-label="Select browser-batch-a"]', true)
  input('[aria-label="Select browser-batch-b"]', true)
  click('[data-testid="delete-selected-users"]')
  await waitUntil(() => row('browser-batch-a') === null && row('browser-batch-b') === null, 'batch delete was not rendered')
  assert(mounted.confirmations() >= 5, 'destructive page commands bypassed confirmation')

  mounted.app.unmount()
  await session.logout()
  await session.login({ username: 'browser-user', password: 'browser replacement password' })
  assert(session.state().status === 'authenticated', 'reset password login failed')
  mounted = await mountAdministration('browser-user')
  const manifest = await mounted.api.manifest()
  assert(manifest.menus.some((value) => value.path === '/iam/browser'), 'assigned menu not projected')
  assert(![...document.querySelectorAll('.tabs button')].some((value) => value.textContent === 'roles' || value.textContent === 'menus'), 'unauthorized administration tabs were visible')
  assert(document.querySelector('[data-testid="create-user"]') === null, 'unauthorized create action was visible')
  assert(row('browser-user')?.querySelector('[data-action]') === null, 'unauthorized row actions were visible')
  const before = await (await control('snapshot')).json() as { users: number }
  let denied = false
  try { await mounted.api.createUser({ username: 'direct-forbidden', displayName: 'Forbidden', email: 'direct-forbidden@example.test', password: 'forbidden password' }) } catch { denied = true }
  assert(denied, 'direct unauthorized API succeeded')
  const after = await (await control('snapshot')).json() as { users: number }
  assert(after.users === before.users, 'direct denial changed state')

  mounted.app.unmount()
  await session.logout()
  await session.login({ username: 'admin', password: 'administrator password' })
  mounted = await mountAdministration()
  clickRow('browser-user', 'delete')
  await waitUntil(() => row('browser-user') === null, 'assigned user cleanup failed')
  openTab('roles')
  clickRow('browser-reader', 'delete')
  await waitUntil(() => row('browser-reader') === null, 'role cleanup failed')
  openTab('menus')
  clickRow('browser-menu', 'delete')
  await waitUntil(() => row('browser-menu') === null, 'menu cleanup failed')
  await control('revoke-role-read', 'POST')
  let revoked = false
  try { await mounted.api.listRoles() } catch { revoked = true }
  assert(revoked, 'permission revoke was not immediate')
  mounted.app.unmount()
  document.body.innerHTML = '<pre id="result">IAM_ADMIN_E2E_PASS</pre>'
  await control('shutdown', 'POST')
}

await scenario().catch(async (error: unknown) => {
  renderFailure(error)
  await control('shutdown', 'POST').catch(() => undefined)
})
