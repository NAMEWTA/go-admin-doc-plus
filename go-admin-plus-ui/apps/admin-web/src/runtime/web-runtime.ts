import type {
  NavigationEntry,
  PermissionCode,
  RuntimeIdentity,
  ShellRuntimePort
} from '@go-admin/platform'

type Fetch = typeof globalThis.fetch
const permissionPattern = /^[a-z][a-z0-9-]*(?::[a-z][a-z0-9-]*){1,2}$/

const parseIdentity = (value: unknown): RuntimeIdentity => {
  if (!value || typeof value !== 'object') throw new Error('invalid identity response')
  const record = value as Record<string, unknown>
  if (record.kind === 'unauthenticated') return { kind: 'unauthenticated' }
  if (
    record.kind !== 'authenticated'
    || typeof record.subjectId !== 'string'
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
    permissions: record.permissions as PermissionCode[]
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

const readJson = async (fetch: Fetch, path: string): Promise<unknown> => {
  const response = await fetch(path, {
    credentials: 'include',
    headers: { accept: 'application/json' }
  })
  if (!response.ok) throw new Error(`runtime request failed with ${response.status}`)
  return response.json()
}

export const createWebRuntime = (fetch: Fetch = globalThis.fetch): ShellRuntimePort => ({
  async loadIdentity() {
    const response = await fetch('/api/runtime/identity', {
      credentials: 'include',
      headers: { accept: 'application/json' }
    })
    if (response.status === 401) return { kind: 'unauthenticated' }
    if (!response.ok) throw new Error(`identity request failed with ${response.status}`)
    return parseIdentity(await response.json())
  },
  async loadNavigation() {
    return parseNavigation(await readJson(fetch, '/api/runtime/navigation'))
  }
})
