import { describe, expect, it } from 'vitest'

import { DemoRequestError } from '@go-admin/domain-demo'

import { createDesktopDemoClient, createDesktopRuntime, createDesktopSession, createDesktopTransport } from './index'

describe('desktop adapter security boundary', () => {
  it('accepts only the public identity projection and rejects hidden session material', async () => {
    const runtime = createDesktopRuntime(async <T>(command: string) => {
      if (command === 'desktop_identity') return {
        kind: 'authenticated',
        profile: { id: 'account-1', username: 'admin', displayName: 'Admin', email: 'admin@example.test' },
        permissions: ['demo.products.read'],
        dataScope: 'all'
      } as T
      throw new Error('unexpected command')
    })
    await expect(runtime.loadIdentity()).resolves.toEqual({
      kind: 'authenticated', subjectId: 'account-1', permissions: ['demo.products.read'], dataScope: 'all'
    })

    const poisoned = createDesktopRuntime(async <T>() => ({
      kind: 'authenticated',
      profile: { id: 'account-1', username: 'admin', displayName: 'Admin', email: 'admin@example.test' },
      permissions: ['demo.products.read'],
      dataScope: 'all',
      csrfToken: 'abcdefghijklmnopqrstuvwxyzABCDEFGH123456789'
    }) as T)
    await expect(poisoned.loadIdentity()).rejects.toThrow('invalid desktop identity')
  })

  it('keeps self-scoped identities fail closed for global Demo capabilities', async () => {
    const runtime = createDesktopRuntime(async <T>() => ({
      kind: 'authenticated',
      profile: { id: 'account-1', username: 'admin', displayName: 'Admin', email: 'admin@example.test' },
      permissions: [],
      dataScope: 'self'
    }) as T)
    await expect(runtime.loadIdentity()).resolves.toMatchObject({ dataScope: 'self', permissions: [] })
  })

  it('never adds headers, origins, or session data to a business invocation', async () => {
    const calls: Array<readonly [string, Record<string, unknown> | undefined]> = []
    const invoke = async <T>(command: string, args?: Record<string, unknown>) => {
      calls.push([command, args])
      return { status: 200, body: { rows: [], total: 0 } } as T
    }
    const transport = createDesktopTransport(invoke)
    await transport.request('/demo/products?page=1', 'GET')
    expect(calls).toEqual([['desktop_request', {
      request: { path: '/demo/products?page=1', method: 'GET', body: undefined }
    }]])
  })

  it('maps stable host statuses and serializes Demo operations', async () => {
    const calls: string[] = []
    const client = createDesktopDemoClient({
      async request(path) {
        calls.push(path)
        return calls.length === 1 ? { status: 403, body: {} } : { status: 200, body: { rows: [], total: 0 } }
      }
    })
    const query = { search: '%_', page: 1, pageSize: 20, sort: 'sku', direction: 'ascending' } as const
    await expect(client.list(query)).rejects.toMatchObject({ category: 'forbidden' } satisfies Partial<DemoRequestError>)
    await expect(client.list(query)).resolves.toEqual({ rows: [], total: 0 })
    expect(calls).toEqual([
      '/demo/products?search=%25_&page=1&pageSize=20&sort=sku&direction=ascending',
      '/demo/products?search=%25_&page=1&pageSize=20&sort=sku&direction=ascending'
    ])
  })

  it('uses dedicated typed commands for login and logout', async () => {
    const commands: string[] = []
    const invoke = async <T>(command: string) => {
      commands.push(command)
      return (command === 'desktop_login'
        ? { id: 'account-1', username: 'admin', displayName: 'Admin', email: 'admin@example.test' }
        : { localCleared: true, remoteRevoked: true }) as T
    }
    const session = createDesktopSession(invoke)
    await expect(session.login('admin', 'administrator password')).resolves.toMatchObject({ id: 'account-1' })
    await session.logout()
    expect(commands).toEqual(['desktop_login', 'desktop_logout'])
  })

  it('rejects logout responses that do not confirm local vault clearing', async () => {
    const session = createDesktopSession(async <T>() => ({ localCleared: false, remoteRevoked: true }) as T)
    await expect(session.logout()).rejects.toThrow('invalid desktop logout result')
  })
})
