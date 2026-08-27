import { invoke as tauriInvoke } from '@tauri-apps/api/core'

import { DemoRequestError, type DemoClient, type DemoFailure, type Product, type ProductPage } from '@go-admin/domain-demo'
import type { NavigationEntry, PermissionCode, RuntimeIdentity, ShellRuntimePort } from '@go-admin/platform'

type Invoke = <T>(command: string, args?: Record<string, unknown>) => Promise<T>
type Method = 'GET' | 'POST' | 'PATCH' | 'DELETE'

interface HostResponse {
  readonly status: number
  readonly body: unknown
}

interface PublicProfile {
  readonly id: string
  readonly username: string
  readonly displayName: string
  readonly email: string
  readonly avatarRef?: string | null
}

export type DesktopDataScope = 'self' | 'all'
export type DesktopIdentity = RuntimeIdentity & { readonly dataScope?: DesktopDataScope }
export interface DesktopRuntime extends Omit<ShellRuntimePort, 'loadIdentity'> {
  loadIdentity(request?: { readonly signal?: AbortSignal }): Promise<DesktopIdentity>
}

const permissionPattern = /^[a-z][a-z0-9-]*(?:\.[a-z][a-z0-9-]*){1,2}$/

const asRecord = (value: unknown): Record<string, unknown> => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('invalid desktop response')
  return value as Record<string, unknown>
}

const exactKeys = (value: Record<string, unknown>, required: ReadonlyArray<string>, optional: ReadonlyArray<string> = []) => {
  const allowed = new Set([...required, ...optional])
  return required.every(key => Object.hasOwn(value, key)) && Object.keys(value).every(key => allowed.has(key))
}

const parseProfile = (value: unknown): PublicProfile => {
  const profile = asRecord(value)
  if (!exactKeys(profile, ['id', 'username', 'displayName', 'email'], ['avatarRef']) ||
    !['id', 'username', 'displayName', 'email'].every(key => typeof profile[key] === 'string') ||
    (profile.avatarRef !== undefined && profile.avatarRef !== null && typeof profile.avatarRef !== 'string')) {
    throw new Error('invalid desktop profile')
  }
  return profile as unknown as PublicProfile
}

const invokeRequest = async (
  invoke: Invoke,
  path: string,
  method: Method,
  body?: unknown,
  signal?: AbortSignal
): Promise<HostResponse> => {
  if (signal?.aborted) throw new DOMException('Request aborted', 'AbortError')
  const response = await invoke<HostResponse>('desktop_request', { request: { path, method, body } })
  if (signal?.aborted) throw new DOMException('Request aborted', 'AbortError')
  if (!Number.isInteger(response.status) || response.status < 100 || response.status > 599) {
    throw new Error('invalid desktop response')
  }
  return response
}

const parseIdentity = (sessionValue: unknown, manifestValue: unknown): DesktopIdentity => {
  const session = asRecord(sessionValue)
  if (!exactKeys(session, ['kind', 'profile', 'permissions', 'dataScope']) || session.kind !== 'authenticated') throw new Error('invalid desktop identity')
  const profile = parseProfile(session.profile)
  const manifest = asRecord(manifestValue)
  if (profile.id.length === 0 || !Array.isArray(manifest.permissions) ||
    (session.dataScope !== 'self' && session.dataScope !== 'all') ||
    !manifest.permissions.every(value => typeof value === 'string' && permissionPattern.test(value))) {
    throw new Error('invalid desktop identity')
  }
  return {
    kind: 'authenticated',
    subjectId: profile.id,
    permissions: manifest.permissions as PermissionCode[],
    dataScope: session.dataScope
  }
}

const parseNavigation = (value: unknown): ReadonlyArray<NavigationEntry> => {
  const record = asRecord(value)
  if (!Array.isArray(record.menus)) throw new Error('invalid desktop navigation')
  const entries = record.menus.map(value => {
    const menu = asRecord(value)
    if (typeof menu.path !== 'string' || !menu.path.startsWith('/') || menu.path.startsWith('//') ||
      typeof menu.permissionCode !== 'string' || !permissionPattern.test(menu.permissionCode)) {
      throw new Error('invalid desktop navigation')
    }
    return { path: menu.path as `/${string}`, permission: menu.permissionCode as PermissionCode }
  })
  if (new Set(entries.map(entry => entry.path)).size !== entries.length) throw new Error('invalid desktop navigation')
  return entries
}

/** Creates the Tauri-backed runtime. Session and CSRF material never enter this adapter. */
export const createDesktopRuntime = (invoke: Invoke = tauriInvoke): DesktopRuntime => ({
  async loadIdentity(request) {
	if (request?.signal?.aborted) throw new DOMException('Request aborted', 'AbortError')
	const result = await invoke<unknown>('desktop_identity')
	if (request?.signal?.aborted) throw new DOMException('Request aborted', 'AbortError')
	if (result && typeof result === 'object' && (result as Record<string, unknown>).kind === 'unauthenticated') {
	  return { kind: 'unauthenticated' }
	}
	const record = asRecord(result)
	return parseIdentity(record, { permissions: record.permissions })
  },
  async loadNavigation(request) {
	if (request?.signal?.aborted) throw new DOMException('Request aborted', 'AbortError')
	const result = await invoke<unknown>('desktop_navigation')
	if (request?.signal?.aborted) throw new DOMException('Request aborted', 'AbortError')
	return parseNavigation({ menus: result })
  }
})

export interface DesktopTransport {
  request(path: string, method: Method, body?: unknown, signal?: AbortSignal): Promise<HostResponse>
}

/** Business clients use this controlled bridge instead of fetch or a loopback URL. */
export const createDesktopTransport = (invoke: Invoke = tauriInvoke): DesktopTransport => ({
  request: (path, method, body, signal) => invokeRequest(invoke, path, method, body, signal)
})

export interface DesktopSession {
  login(username: string, password: string): Promise<PublicProfile>
  logout(): Promise<{ readonly localCleared: true, readonly remoteRevoked: boolean }>
}

export const createDesktopSession = (invoke: Invoke = tauriInvoke): DesktopSession => ({
  login: async (username, password) => parseProfile(await invoke<unknown>('desktop_login', { username, password })),
  logout: async () => {
    const value = asRecord(await invoke<unknown>('desktop_logout'))
    if (!exactKeys(value, ['localCleared', 'remoteRevoked']) || value.localCleared !== true || typeof value.remoteRevoked !== 'boolean') {
      throw new Error('invalid desktop logout result')
    }
    return { localCleared: true, remoteRevoked: value.remoteRevoked }
  }
})

const failure = (status: number): DemoFailure => {
  if (status === 401) return 'relogin'
  if (status === 403) return 'forbidden'
  if (status === 400 || status === 422) return 'validation'
  if (status === 404) return 'not-found'
  if (status === 409) return 'conflict'
  return 'unavailable'
}

/** Demo client backed by the Rust allowlisted bridge. Host response schemas are rechecked in Rust. */
export const createDesktopDemoClient = (transport: DesktopTransport = createDesktopTransport()): DemoClient => {
  let tail: Promise<void> = Promise.resolve()
  const serialized = <T>(operation: () => Promise<T>): Promise<T> => {
    const result = tail.then(operation, operation)
    tail = result.then(() => undefined, () => undefined)
    return result
  }
  const required = async <T>(path: string, method: Method, body?: unknown): Promise<T> => {
    const response = await transport.request(path, method, body)
    if (response.status < 200 || response.status >= 300) throw new DemoRequestError(failure(response.status))
    return response.body as T
  }
  return {
    list: query => serialized(() => required<ProductPage>(`/demo/products?${new URLSearchParams({
      search: query.search, page: String(query.page), pageSize: String(query.pageSize),
      sort: query.sort, direction: query.direction
    })}`, 'GET')),
    get: id => serialized(() => required<Product>(`/demo/products/${encodeURIComponent(id)}`, 'GET')),
    create: input => serialized(() => required<Product>('/demo/products', 'POST', input)),
    update: (id, input) => serialized(() => required<Product>(`/demo/products/${encodeURIComponent(id)}`, 'PATCH', input)),
    delete: products => serialized(async () => { await required<null>('/demo/products/batch-delete', 'POST', { products: [...products] }) })
  }
}
