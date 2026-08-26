import { describe, expect, it, vi } from 'vitest'
import { createWebDemoClient } from './web-demo-client'

const json = (status: number, body: unknown, headers: Record<string, string> = {}) => new Response(JSON.stringify(body), { status, headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json', ...headers } })

describe('web demo client', () => {
  it('carries CSRF only on mutations and classifies ordinary authorization separately', async () => {
    const fetcher = vi.fn<typeof fetch>()
      .mockResolvedValueOnce(json(200, { rows: [], total: 0 }, { 'X-CSRF-Token': 'c'.repeat(32) }))
      .mockResolvedValueOnce(json(403, { category: 'authorization', code: 'PERMISSION_DENIED' }))
    const client = createWebDemoClient(fetcher, 'https://demo.test/api')
    await client.list({ search: '', page: 1, pageSize: 20, sort: 'updatedAt', direction: 'descending' })
    await expect(client.create({ sku: 'DEMO-01', name: 'Demo product', description: '', priceCents: 1, status: 'active' })).rejects.toMatchObject({ category: 'forbidden' })
    expect(new Request(fetcher.mock.calls[0]![0]).headers.has('X-CSRF-Token')).toBe(false)
    expect(new Request(fetcher.mock.calls[1]![0]).headers.get('X-CSRF-Token')).toBe('c'.repeat(32))
  })

  it('turns CSRF rejection into a re-login signal', async () => {
    const client = createWebDemoClient(vi.fn<typeof fetch>().mockResolvedValue(json(403, { category: 'authorization', code: 'CSRF_REJECTED' })), 'https://demo.test/api')
    await expect(client.delete([{ id: '00000000-0000-4000-8000-000000000001', revision: 1 }])).rejects.toMatchObject({ category: 'relogin' })
  })
})
