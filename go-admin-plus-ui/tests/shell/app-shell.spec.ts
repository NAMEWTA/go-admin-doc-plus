import { describe, expect, it, vi } from 'vitest'
import { createShellNavigator, resolveShellState } from '@go-admin-plus/app-shell'
import {
  createProductMemoryHistory,
  createProductRouter,
  productBreadcrumbs,
  productHistoryMode,
  productRoutesFor,
  resolveAuthorizedProductRoutes
} from '@go-admin-plus/app-shell/product'
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

describe('product router', () => {
  const productRuntime = (
    permissions: Extract<RuntimeIdentity, { kind: 'authenticated' }>['permissions'] = ['demo.products.read']
  ): ShellRuntimePort => ({
    loadIdentity: async () => ({
      kind: 'authenticated',
      subjectId: 'user-1',
      permissions,
      dataScope: 'all'
    }),
    loadNavigation: async () => [
      { path: '/demo/products', permission: 'demo.products.read' },
      { path: '/files', permission: 'files.objects.read' }
    ]
  })

  it('uses host-specific history with one compiled business manifest', () => {
    expect(productHistoryMode('web')).toBe('html5')
    expect(productHistoryMode('desktop')).toBe('hash')
    expect(productRoutesFor('web').map(route => route.name))
      .toEqual(productRoutesFor('desktop').map(route => route.name))
  })

  it('resolves deep links and derives title hierarchy from the current route', async () => {
    const router = createProductRouter('web', productRuntime(), { history: createProductMemoryHistory() })
    await router.push('/demo/products')
    await router.isReady()

    expect(router.currentRoute.value.name).toBe('demo-products')
    expect(productBreadcrumbs('web', router.currentRoute.value)).toEqual([
      { title: '工作台', path: '/' },
      { title: '示例业务' },
      { title: '产品示例' }
    ])
  })

  it('keeps forbidden and unknown routes distinct', async () => {
    const forbidden = createProductRouter('web', productRuntime([]), { history: createProductMemoryHistory() })
    await forbidden.push('/demo/products')
    await forbidden.isReady()
    expect(forbidden.currentRoute.value.name).toBe('forbidden')

    const unknown = createProductRouter('web', productRuntime(), { history: createProductMemoryHistory() })
    await unknown.push('/missing')
    await unknown.isReady()
    expect(unknown.currentRoute.value.name).toBe('not-found')
  })

  it('uses a recoverable unavailable route for bounded runtime failures', async () => {
    const runtimeFailure: ShellRuntimePort = {
      loadIdentity: async () => { throw new Error('secret=adapter-detail') },
      loadNavigation: async () => []
    }
    const router = createProductRouter('web', runtimeFailure, { history: createProductMemoryHistory() })
    await router.push('/demo/products')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('unavailable')
    expect(JSON.stringify(router.currentRoute.value)).not.toContain('adapter-detail')
  })

  it('restores the same route truth through history navigation', async () => {
    const router = createProductRouter(
      'web',
      productRuntime(['demo.products.read', 'files.objects.read']),
      { history: createProductMemoryHistory() }
    )
    await router.push('/demo/products')
    await router.isReady()
    await router.push('/files')

    const restored = new Promise<void>(resolve => {
      const remove = router.afterEach(to => {
        if (to.name === 'demo-products') {
          remove()
          resolve()
        }
      })
    })
    router.back()
    await restored
    expect(router.currentRoute.value.path).toBe('/demo/products')
  })

  it('intersects server navigation with compiled routes and ignores component payloads', () => {
    const identity = {
      kind: 'authenticated' as const,
      subjectId: 'user-1',
      permissions: ['demo.products.read' as const],
      dataScope: 'all' as const
    }
    const navigation = [{
      path: '/demo/products' as const,
      permission: 'demo.products.read' as const,
      component: 'https://outside.example/remote.js'
    }]

    const allowed = resolveAuthorizedProductRoutes('web', identity, navigation)
    expect(allowed.map(route => route.name)).toEqual(['demo-products'])
    expect(allowed[0]?.component).toBe(productRoutesFor('web').find(route => route.name === 'demo-products')?.component)
  })
})
