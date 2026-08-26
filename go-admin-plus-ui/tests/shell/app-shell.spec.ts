import { describe, expect, it, vi } from 'vitest'

import { resolveShellState } from '@go-admin/app-shell'
import type { ShellRuntimePort } from '@go-admin/platform'

const runtime = (overrides: Partial<ShellRuntimePort> = {}): ShellRuntimePort => ({
  loadIdentity: async () => ({
    kind: 'authenticated',
    subjectId: 'user-1',
    permissions: ['demo:read']
  }),
  loadNavigation: async () => [{ path: '/demo', permission: 'demo:read' }],
  ...overrides
})

describe('app shell state', () => {
  it('returns a stable authenticated state for an allowed route', async () => {
    await expect(resolveShellState(runtime(), '/demo')).resolves.toEqual({
      kind: 'authenticated',
      path: '/demo',
      subjectId: 'user-1'
    })
  })

  it('returns login without requesting navigation for an unauthenticated identity', async () => {
    const loadNavigation = vi.fn()
    const result = await resolveShellState(runtime({
      loadIdentity: async () => ({ kind: 'unauthenticated' }),
      loadNavigation
    }), '/demo')

    expect(result).toEqual({ kind: 'unauthenticated', redirectTo: '/login' })
    expect(loadNavigation).not.toHaveBeenCalled()
  })

  it('distinguishes unauthorized and unknown routes', async () => {
    await expect(resolveShellState(runtime({
      loadIdentity: async () => ({
        kind: 'authenticated',
        subjectId: 'user-2',
        permissions: []
      })
    }), '/demo')).resolves.toEqual({ kind: 'unauthorized', path: '/demo' })

    await expect(resolveShellState(runtime(), '/missing')).resolves.toEqual({
      kind: 'not-found',
      path: '/missing'
    })
  })

  it('contains no adapter error or credential material in failure state', async () => {
    const result = await resolveShellState(runtime({
      loadIdentity: async () => { throw new Error('secret=session-value') }
    }), '/demo')

    expect(result).toEqual({ kind: 'adapter-failed', retryable: true })
    expect(JSON.stringify(result)).not.toContain('session-value')
  })
})
