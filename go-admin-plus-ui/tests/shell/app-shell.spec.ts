import { describe, expect, it, vi } from 'vitest'

import { createShellNavigator, resolveShellState } from '@go-admin-plus/app-shell'
import type { RuntimeIdentity, RuntimeRequest, ShellRuntimePort } from '@go-admin-plus/platform'

const deferred = <T>() => {
  let resolve!: (value: T) => void
  const promise = new Promise<T>(done => { resolve = done })
  return { promise, resolve }
}

const runtime = (overrides: Partial<ShellRuntimePort> = {}): ShellRuntimePort => ({
  loadIdentity: async () => ({
    kind: 'authenticated',
    subjectId: 'user-1',
    permissions: ['demo.products.read'],
    dataScope: 'all'
  }),
  loadNavigation: async () => [{ path: '/demo', permission: 'demo.products.read' }],
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
        permissions: [],
        dataScope: 'self'
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

  it('commits only the latest navigation when requests resolve out of order', async () => {
    const first = deferred<RuntimeIdentity>()
    const second = deferred<RuntimeIdentity>()
    const requests: AbortSignal[] = []
    const loadIdentity = vi.fn(({ signal }: RuntimeRequest = {}) => {
      requests.push(signal as AbortSignal)
      return requests.length === 1 ? first.promise : second.promise
    })
    const commit = vi.fn()
    const setLoading = vi.fn()
    const navigator = createShellNavigator(runtime({ loadIdentity }), { commit, setLoading })

    const staleNavigation = navigator.navigate('/old')
    const latestNavigation = navigator.navigate('/demo')
    expect(requests[0]?.aborted).toBe(true)
    expect(requests[1]?.aborted).toBe(false)

    second.resolve({
      kind: 'authenticated',
      subjectId: 'user-2',
      permissions: ['demo.products.read'],
      dataScope: 'all'
    })
    await latestNavigation
    first.resolve({
      kind: 'authenticated',
      subjectId: 'user-1',
      permissions: ['demo.products.read'],
      dataScope: 'self'
    })
    await staleNavigation

    expect(commit).toHaveBeenCalledTimes(1)
    expect(commit).toHaveBeenCalledWith('/demo', {
      kind: 'authenticated',
      path: '/demo',
      subjectId: 'user-2'
    })
    expect(setLoading.mock.calls).toEqual([[true], [true], [false]])
  })

  it('invalidates pending navigation without committing after unmount', async () => {
    const identity = deferred<RuntimeIdentity>()
    const commit = vi.fn()
    const setLoading = vi.fn()
    const navigator = createShellNavigator(runtime({ loadIdentity: () => identity.promise }), {
      commit,
      setLoading
    })

    const navigation = navigator.navigate('/demo')
    navigator.invalidate()
    identity.resolve({
      kind: 'authenticated',
      subjectId: 'user-1',
      permissions: ['demo.products.read'],
      dataScope: 'all'
    })
    await navigation

    expect(commit).not.toHaveBeenCalled()
    expect(setLoading).toHaveBeenCalledTimes(1)
  })
})
