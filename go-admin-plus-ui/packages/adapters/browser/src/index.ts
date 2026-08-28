import type {
  NavigationEntry,
  PermissionCode,
  RuntimeIdentity,
  RuntimeRequest,
  ShellRuntimePort
} from '@go-admin-plus/platform'

export { createBrowserFilesClient } from './files'
export { createBrowserSessionFetch } from './session-fetch'

type Fetch = typeof globalThis.fetch
const permissionPattern = /^[a-z][a-z0-9-]*(?:\.[a-z][a-z0-9-]*){1,2}$/

const hasExactKeys = (record: Record<string, unknown>, expected: ReadonlyArray<string>) => {
  const keys = Object.keys(record).sort()
  const expectedKeys = [...expected].sort()
  return keys.length === expectedKeys.length && keys.every((key, index) => key === expectedKeys[index])
}

const parseIdentity = (value: unknown): RuntimeIdentity => {
  if (!value || typeof value !== 'object') throw new Error('invalid identity response')
  const record = value as Record<string, unknown>
  if (record.kind === 'unauthenticated') {
    if (!hasExactKeys(record, ['kind'])) throw new Error('invalid identity response')
    return { kind: 'unauthenticated' }
  }
  if (
    record.kind !== 'authenticated'
    || !hasExactKeys(record, ['dataScope', 'kind', 'permissions', 'subjectId'])
    || typeof record.subjectId !== 'string'
    || record.subjectId.length === 0
    || (record.dataScope !== 'self' && record.dataScope !== 'all')
    || !Array.isArray(record.permissions)
    || !record.permissions.every(permission =>
      typeof permission === 'string' && permissionPattern.test(permission)
    )
  ) {
    throw new Error('invalid identity response')
  }
  return {
    kind: 'authenticated',
    subjectId: record.subjectId,
    permissions: record.permissions as PermissionCode[],
    dataScope: record.dataScope
  }
}

const parseNavigation = (value: unknown): ReadonlyArray<NavigationEntry> => {
  if (!Array.isArray(value)) throw new Error('invalid navigation response')
  const entries = value.map(entry => {
    if (!entry || typeof entry !== 'object') throw new Error('invalid navigation entry')
    const record = entry as Record<string, unknown>
    if (
      typeof record.path !== 'string'
      || !record.path.startsWith('/')
      || record.path.startsWith('//')
      || record.path.includes('\\')
    ) {
      throw new Error('invalid navigation entry')
    }
    if (
      record.permission !== undefined
      && (typeof record.permission !== 'string' || !permissionPattern.test(record.permission))
    ) {
      throw new Error('invalid navigation entry')
    }
    return {
      path: record.path as `/${string}`,
      ...(record.permission ? { permission: record.permission as PermissionCode } : {})
    }
  })
  if (new Set(entries.map(entry => entry.path)).size !== entries.length) {
    throw new Error('duplicate navigation path')
  }
  return entries
}

const readJson = async (fetch: Fetch, path: string, request?: RuntimeRequest): Promise<unknown> => {
  const response = await fetch(path, {
    credentials: 'include',
    headers: { accept: 'application/json' },
    ...(request?.signal ? { signal: request.signal } : {})
  })
  if (!response.ok) throw new Error(`runtime request failed with ${response.status}`)
  return response.json()
}

/** Creates the browser adapter; cookie credentials remain inside fetch and never cross the port. */
export const createWebRuntime = (fetch: Fetch = globalThis.fetch): ShellRuntimePort => ({
  async loadIdentity(request) {
    const response = await fetch('/api/runtime/identity', {
      credentials: 'include',
      headers: { accept: 'application/json' },
      ...(request?.signal ? { signal: request.signal } : {})
    })
    if (response.status === 401) return { kind: 'unauthenticated' }
    if (!response.ok) throw new Error(`identity request failed with ${response.status}`)
    return parseIdentity(await response.json())
  },
  async loadNavigation(request) {
    return parseNavigation(await readJson(fetch, '/api/runtime/navigation', request))
  }
})
