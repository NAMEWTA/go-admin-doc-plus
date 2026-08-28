import { invoke as tauriInvoke } from '@tauri-apps/api/core'

import { DemoRequestError, type DemoClient, type DemoFailure, type Product, type ProductPage } from '@go-admin-plus/domain-demo'
import { SessionRequestError, type AccountProfile, type SessionClient } from '@go-admin-plus/domain-iam/session'
import type { DataScope, HostFile, HostFileSaveResult, NavigationEntry, PermissionCode, PlatformPort, RuntimeIdentity, ShellRuntimePort } from '@go-admin-plus/platform'

type Invoke = <T>(command: string, args?: Record<string, unknown>) => Promise<T>
type Method = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
const hostCapabilities = Object.freeze(['clipboard-write', 'file-open', 'file-save', 'notification'] as const)
const maximumHostFileBytes = 10 * 1024 * 1024
const hostMediaTypes = new Set(['application/pdf', 'image/jpeg', 'image/png', 'text/plain'])
const validHostFileName = (name: string) => name.length > 0 && name.length <= 255 && !/[\\/\0]/.test(name)

const parseHostFile = (value: unknown): HostFile => {
  const record = asRecord(value)
  if (!exactKeys(record, ['name', 'mediaType', 'sizeBytes', 'data']) ||
    typeof record.name !== 'string' || !validHostFileName(record.name) ||
    typeof record.mediaType !== 'string' || !hostMediaTypes.has(record.mediaType) ||
    !Number.isSafeInteger(record.sizeBytes) || Number(record.sizeBytes) < 0 || Number(record.sizeBytes) > maximumHostFileBytes ||
    typeof record.data !== 'string' || !/^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$/.test(record.data)) {
    throw new Error('invalid desktop host file')
  }
  const bytes = base64ToBytes(record.data)
  if (bytes.length !== record.sizeBytes || bytesToBase64(bytes) !== record.data) throw new Error('invalid desktop host file')
  return { name: record.name, mediaType: record.mediaType, bytes }
}

const encodeHostFile = (file: HostFile) => {
  if (!validHostFileName(file.name) || !hostMediaTypes.has(file.mediaType) || file.bytes.length > maximumHostFileBytes) {
    throw new Error('invalid desktop host file')
  }
  return { name: file.name, mediaType: file.mediaType, data: bytesToBase64(file.bytes) }
}

export const createDesktopPlatform = (invoke: Invoke = tauriInvoke): PlatformPort => ({
  runtime: 'desktop',
  listCapabilities: () => new Set(hostCapabilities),
  pickFile: async () => {
    const value = await invoke<unknown>('desktop_pick_file')
    return value === null ? null : parseHostFile(value)
  },
  saveFile: async file => {
    const value = asRecord(await invoke<unknown>('desktop_save_file', { file: encodeHostFile(file) }))
    if (!exactKeys(value, ['status']) || (value.status !== 'saved' && value.status !== 'cancelled')) {
      throw new Error('invalid desktop save result')
    }
    return value.status as HostFileSaveResult
  },
  notify: message => invoke<void>('desktop_notify', { message }),
  writeClipboard: text => invoke<void>('desktop_write_clipboard', { text })
})

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

export type DesktopDataScope = DataScope
export type DesktopIdentity = RuntimeIdentity
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

interface BinaryBody {
  readonly encoding: 'base64'
  readonly name: string
  readonly mediaType: string
  readonly data: string
}

const bytesToBase64 = (bytes: Uint8Array): string => {
  let binary = ''
  for (let offset = 0; offset < bytes.length; offset += 0x8000) {
    binary += String.fromCharCode(...bytes.subarray(offset, offset + 0x8000))
  }
  return btoa(binary)
}

const base64ToBytes = (value: string): Uint8Array => {
  const binary = atob(value)
  return Uint8Array.from(binary, character => character.charCodeAt(0))
}

const requestBody = async (request: Request): Promise<unknown> => {
  if (request.method === 'GET' || request.method === 'HEAD') return undefined
  const contentType = request.headers.get('content-type') ?? ''
  if (contentType.startsWith('multipart/form-data')) {
    const form = await request.clone().formData()
    const file = form.get('file')
    if (!(file instanceof File) || [...form.keys()].some(key => key !== 'file')) throw new Error('invalid desktop upload')
    return {
      encoding: 'base64',
      name: file.name,
      mediaType: file.type,
      data: bytesToBase64(new Uint8Array(await file.arrayBuffer()))
    } satisfies BinaryBody
  }
  const text = await request.clone().text()
  if (text === '') return undefined
  if (!contentType.includes('json')) throw new Error('desktop request content type is not allowed')
  return JSON.parse(text) as unknown
}

/** Fetch-compatible product bridge. Cookies, session tokens, and CSRF remain in Rust Stronghold. */
export const createDesktopFetch = (invoke: Invoke = tauriInvoke): typeof globalThis.fetch => async (input, init) => {
  const request = input instanceof Request ? new Request(input, init) : new Request(new URL(String(input), window.location.origin), init)
  const url = new URL(request.url)
  if (url.origin !== window.location.origin || !url.pathname.startsWith('/api/')) throw new Error('desktop request path is not allowed')
  const method = request.method.toUpperCase() as Method
  if (!['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].includes(method)) throw new Error('desktop request method is not allowed')
  const response = await invokeRequest(invoke, `${url.pathname.slice(4)}${url.search}`, method, await requestBody(request), request.signal)
  const binary = response.body && typeof response.body === 'object' && !Array.isArray(response.body)
    ? response.body as Record<string, unknown>
    : null
  if (binary?.encoding === 'base64' && typeof binary.data === 'string' && typeof binary.mediaType === 'string') {
    return new Response(base64ToBytes(binary.data).buffer as ArrayBuffer, { status: response.status, headers: { 'content-type': binary.mediaType } })
  }
  return new Response(response.body === null ? null : JSON.stringify(response.body), {
    status: response.status,
    headers: response.body === null ? undefined : { 'content-type': 'application/json' }
  })
}

const sessionFailure = (status: number): SessionRequestError => new SessionRequestError(
  status === 401 ? 'authentication' : status === 403 ? 'authorization' : status === 400 || status === 422
    ? 'validation' : status === 409 ? 'conflict' : 'unavailable'
)

/** Session port backed by native commands and the controlled product bridge. */
export const createDesktopSessionClient = (invoke: Invoke = tauriInvoke): SessionClient => {
  const fetcher = createDesktopFetch(invoke)
  const identityProfile = async (): Promise<AccountProfile> => {
    const result = asRecord(await invoke<unknown>('desktop_identity'))
    if (result.kind !== 'authenticated') throw new SessionRequestError('authentication')
    return parseProfile(result.profile) as AccountProfile
  }
  const json = async <T>(path: string, init?: RequestInit): Promise<T> => {
    const response = await fetcher(`/api${path}`, init)
    if (!response.ok) throw sessionFailure(response.status)
    return response.status === 204 ? undefined as T : await response.json() as T
  }
  return {
    login: async credentials => parseProfile(await invoke<unknown>('desktop_login', credentials)) as AccountProfile,
    current: identityProfile,
    logout: async () => { await createDesktopSession(invoke).logout() },
    profile: identityProfile,
    updateProfile: update => json<AccountProfile>('/iam/account/profile', {
      method: 'PATCH', headers: { 'content-type': 'application/json' }, body: JSON.stringify(update)
    }),
    changePassword: async (currentPassword, newPassword) => {
      await json<void>('/iam/account/password', {
        method: 'PUT', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ currentPassword, newPassword })
      })
    }
  }
}

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
