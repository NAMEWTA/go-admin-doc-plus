import { describe, expect, it, vi } from 'vitest'

import { createWebRuntime } from './index'

describe('browser product runtime', () => {
  it('accepts canonical dot-separated permissions and rejects legacy separators', async () => {
    const fetcher = vi.fn<typeof fetch>()
      .mockResolvedValueOnce(new Response(JSON.stringify({
        kind: 'authenticated', subjectId: 'account-1', permissions: ['iam.users.read', 'files.objects.write']
      }), { status: 200, headers: { 'content-type': 'application/json' } }))
      .mockResolvedValueOnce(new Response(JSON.stringify([
        { path: '/iam/users', permission: 'iam.users.read' },
        { path: '/files', permission: 'files.objects.write' }
      ]), { status: 200, headers: { 'content-type': 'application/json' } }))
    const runtime = createWebRuntime(fetcher)
    await expect(runtime.loadIdentity()).resolves.toMatchObject({ kind: 'authenticated', permissions: ['iam.users.read', 'files.objects.write'] })
    await expect(runtime.loadNavigation()).resolves.toHaveLength(2)

    const legacy = createWebRuntime(vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({
      kind: 'authenticated', subjectId: 'account-1', permissions: ['iam:users:read']
    }), { status: 200, headers: { 'content-type': 'application/json' } })))
    await expect(legacy.loadIdentity()).rejects.toThrow('invalid identity response')
  })
})
