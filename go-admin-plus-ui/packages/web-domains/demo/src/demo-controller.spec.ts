import { describe, expect, it, vi } from 'vitest'
import { DemoRequestError, type DemoClient, type Product, type ProductInput, type ProductQuery } from '@go-admin/domain-demo'
import { createDemoController } from './demo-controller'

const product = (id = '00000000-0000-4000-8000-000000000001', revision = 1): Product => ({ id, sku: 'DEMO-01', name: 'Demo product', description: '', priceCents: 10, status: 'active', revision, createdAt: '2026-08-27T00:00:00Z', updatedAt: '2026-08-27T00:00:00Z' })
const input: ProductInput = { sku: 'DEMO-01', name: 'Demo product', description: '', priceCents: 10, status: 'active' }

const fixture = () => {
  let rows = [product()]
  const client: DemoClient = {
    list: vi.fn(async (_query: ProductQuery) => ({ rows, total: rows.length })),
    get: vi.fn(async () => rows[0]!),
    create: vi.fn(async value => { const created = { ...product(), ...value }; rows = [created]; return created }),
    update: vi.fn(async (_id, value) => { const updated = { ...product(undefined, value.revision + 1), ...value }; rows = [updated]; return updated }),
    delete: vi.fn(async () => { rows = [] }),
  }
  return { client, controller: createDemoController(client, vi.fn(async () => true)) }
}

describe('demo controller', () => {
  it('searches, resets, pages, sorts and selects through the shared list state', async () => {
    const { client, controller } = fixture()
    await controller.list.search({ search: 'demo' })
    await controller.list.setPageSize(10)
    await controller.list.setSort({ key: 'priceCents', direction: 'ascending' })
    controller.list.select(controller.list.snapshot().rows)
    expect(controller.list.snapshot().selectedKeys).toEqual([product().id])
    expect(client.list).toHaveBeenLastCalledWith(expect.objectContaining({ search: 'demo', pageSize: 10, sort: 'priceCents', direction: 'ascending' }))
    await controller.list.reset()
    expect(controller.list.snapshot().filters.search).toBe('')
  })

  it('fences new writes and repairs repeated refresh failures without duplicating create', async () => {
    const { client, controller } = fixture()
    vi.mocked(client.list).mockRejectedValueOnce(new DemoRequestError('unavailable')).mockRejectedValueOnce(new DemoRequestError('unavailable')).mockResolvedValue({ rows: [product()], total: 1 })
    expect(await controller.save(input)).toBe('refresh-failed')
    expect(controller.pendingRepair).toBe(true)
    expect(await controller.save({ ...input, sku: 'DEMO-02' })).toBe('refresh-failed')
    expect(await controller.repairProjection()).toBe('refresh-failed')
    expect(await controller.repairProjection()).toBe('completed')
    expect(client.create).toHaveBeenCalledTimes(1)
    expect(controller.pendingRepair).toBe(false)
  })

  it('confirms removal once and never repeats a successful delete while repairing', async () => {
    const { client } = fixture()
    const confirm = vi.fn(async () => true)
    const managed = createDemoController(client, confirm)
    vi.mocked(client.list).mockRejectedValueOnce(new DemoRequestError('unavailable')).mockResolvedValue({ rows: [], total: 0 })
    expect(await managed.remove([product()])).toBe('refresh-failed')
    expect(await managed.remove([product('00000000-0000-4000-8000-000000000002')])).toBe('refresh-failed')
    expect(await managed.repairProjection()).toBe('completed')
    expect(client.delete).toHaveBeenCalledTimes(1)
    expect(confirm).toHaveBeenCalledTimes(1)
  })

  it('classifies list failures and returns stable state', async () => {
    const { client, controller } = fixture()
    vi.mocked(client.list).mockRejectedValue(new DemoRequestError('forbidden'))
    await expect(controller.list.refresh()).rejects.toMatchObject({ category: 'forbidden' })
    expect(controller.failure()).toBe('forbidden')
  })
})
