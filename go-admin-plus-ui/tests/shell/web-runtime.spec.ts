import { describe, expect, it, vi } from 'vitest'

import { createWebRuntime } from '@go-admin/adapter-browser'

const response = (status: number, value?: unknown): Response => ({
  ok: status >= 200 && status < 300,
  status,
  json: async () => value
} as Response)

describe('web runtime adapter', () => {
  it('uses cookie credentials without accepting or returning a raw session', async () => {
    const fetch = vi.fn(async () => response(200, {
      kind: 'authenticated',
      subjectId: 'user-1',
      permissions: ['demo:read']
    }))
    const runtime = createWebRuntime(fetch as typeof globalThis.fetch)

    await expect(runtime.loadIdentity()).resolves.toEqual({
      kind: 'authenticated',
      subjectId: 'user-1',
      permissions: ['demo:read']
    })
    expect(fetch).toHaveBeenCalledWith('/api/runtime/identity', {
      credentials: 'include',
      headers: { accept: 'application/json' }
    })
    expect(JSON.stringify(await runtime.loadIdentity())).not.toMatch(/session|secret|token/i)
  })

  it('maps 401 to unauthenticated and rejects malformed runtime data', async () => {
    await expect(createWebRuntime(async () => response(401) as never).loadIdentity())
      .resolves.toEqual({ kind: 'unauthenticated' })
    await expect(createWebRuntime(async () => response(200, { kind: 'authenticated' }) as never).loadIdentity())
      .rejects.toThrow('invalid identity response')
  })

  it('rejects unsafe, duplicate, and malformed navigation entries', async () => {
    const unsafe = createWebRuntime(async () => response(200, [{ path: '//outside.example' }]) as never)
    await expect(unsafe.loadNavigation()).rejects.toThrow('invalid navigation entry')

    const duplicate = createWebRuntime(async () => response(200, [
      { path: '/demo', permission: 'demo:read' },
      { path: '/demo', permission: 'demo:write' }
    ]) as never)
    await expect(duplicate.loadNavigation()).rejects.toThrow('duplicate navigation path')
  })
})
