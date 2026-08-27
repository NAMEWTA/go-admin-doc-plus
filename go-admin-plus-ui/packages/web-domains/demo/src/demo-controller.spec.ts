import { describe, expect, it, vi } from 'vitest'
import { DemoRequestError, demoPermissions, type DemoClient, type Product, type ProductInput, type ProductQuery } from '@go-admin/domain-demo'
import { createDemoController } from './demo-controller'
import pageSource from './DemoProductsPage.vue?raw'

const product = (id = '00000000-0000-4000-8000-000000000001', revision = 1): Product => ({ id, sku: 'DEMO-01', name: 'Demo product', description: '', priceCents: 10, status: 'active', revision, createdAt: '2026-08-27T00:00:00Z', updatedAt: '2026-08-27T00:00:00Z' })
const input: ProductInput = { sku: 'DEMO-01', name: 'Demo product', description: '', priceCents: 10, status: 'active' }
const allowAll = { can: () => true }
const deferred = <T>() => {
  let resolve!: (value: T) => void
  let reject!: (error: unknown) => void
  const promise = new Promise<T>((success, failure) => { resolve = success; reject = failure })
  return { promise, resolve, reject }
}

const fixture = () => {
  let rows = [product()]
  const client: DemoClient = {
    list: vi.fn(async (_query: ProductQuery) => ({ rows, total: rows.length })),
    get: vi.fn(async () => rows[0]!),
    create: vi.fn(async value => { const created = { ...product(), ...value }; rows = [created]; return created }),
    update: vi.fn(async (_id, value) => { const updated = { ...product(undefined, value.revision + 1), ...value }; rows = [updated]; return updated }),
    delete: vi.fn(async () => { rows = [] }),
  }
  return { client, controller: createDemoController(client, vi.fn(async () => true), allowAll) }
}

describe('demo controller', () => {
  it('binds the page data and action regions to the fail-closed projection', () => {
    expect(pageSource).toContain('v-if="projectionVisible && canRead" class="demo-products__search"')
    expect(pageSource).toContain('v-if="projectionVisible && canRead" class="demo-products__grid"')
  })

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
    await controller.list.refresh()
    vi.mocked(client.list).mockRejectedValueOnce(new DemoRequestError('unavailable')).mockRejectedValueOnce(new DemoRequestError('unavailable')).mockResolvedValue({ rows: [product()], total: 1 })
    expect(await controller.save(input)).toBe('refresh-failed')
    expect(controller.pendingRepair).toBe(true)
    expect(await controller.save({ ...input, sku: 'DEMO-02' })).toBe('refresh-failed')
    expect(await controller.repairProjection()).toBe('refresh-failed')
    expect(await controller.repairProjection()).toBe('completed')
    expect(client.create).toHaveBeenCalledTimes(1)
    expect(controller.pendingRepair).toBe(false)
    expect(controller.takeCompletion()).toBe('save')
    expect(controller.takeCompletion()).toBeNull()
  })

  it('confirms removal once and never repeats a successful delete while repairing', async () => {
    const { client } = fixture()
    const confirm = vi.fn(async () => true)
    const managed = createDemoController(client, confirm, allowAll)
    await managed.list.refresh()
    vi.mocked(client.list).mockRejectedValueOnce(new DemoRequestError('unavailable')).mockResolvedValue({ rows: [], total: 0 })
    expect(await managed.remove([product()])).toBe('refresh-failed')
    expect(await managed.remove([product('00000000-0000-4000-8000-000000000002')])).toBe('refresh-failed')
    expect(await managed.repairProjection()).toBe('completed')
    expect(client.delete).toHaveBeenCalledTimes(1)
    expect(confirm).toHaveBeenCalledTimes(1)
  })

  it('repairs an edited projection without repeating update and exposes one completion', async () => {
    const { client, controller } = fixture()
    await controller.list.refresh()
    vi.mocked(client.list).mockRejectedValueOnce(new DemoRequestError('unavailable')).mockRejectedValueOnce(new DemoRequestError('unavailable')).mockResolvedValue({ rows: [product(undefined, 2)], total: 1 })
    expect(await controller.save({ ...input, id: product().id, revision: 1 })).toBe('refresh-failed')
    expect(await controller.repairProjection()).toBe('refresh-failed')
    expect(controller.takeCompletion()).toBeNull()
    expect(await controller.repairProjection()).toBe('completed')
    expect(client.update).toHaveBeenCalledTimes(1)
    expect(controller.takeCompletion()).toBe('save')
    expect(controller.takeCompletion()).toBeNull()
  })

  it('fails closed when the host capability port withdraws permissions', async () => {
    const { client } = fixture()
    const granted = new Set<string>([demoPermissions.read, demoPermissions.write, demoPermissions.delete])
    const controller = createDemoController(client, vi.fn(async () => true), { can: code => granted.has(code) })
    await controller.list.refresh()
    granted.delete(demoPermissions.write)
    expect(await controller.save(input)).toBe('failed')
    granted.delete(demoPermissions.delete)
    expect(await controller.remove([product()])).toBe('failed')
    granted.delete(demoPermissions.read)
    await expect(controller.list.refresh()).rejects.toMatchObject({ category: 'forbidden' })
    expect(client.create).not.toHaveBeenCalled()
    expect(client.delete).not.toHaveBeenCalled()
    expect(controller.failure()).toBe('forbidden')
    expect(controller.projectionVisible).toBe(false)
    expect(controller.list.snapshot()).toMatchObject({ rows: [], total: 0, selectedKeys: [] })
  })

  it('classifies list failures and returns stable state', async () => {
    const { client, controller } = fixture()
    vi.mocked(client.list).mockRejectedValue(new DemoRequestError('forbidden'))
    await expect(controller.list.refresh()).rejects.toMatchObject({ category: 'forbidden' })
    expect(controller.failure()).toBe('forbidden')
  })

  for (const category of ['forbidden', 'relogin', 'unavailable'] as const) {
    it(`hides a previously successful projection after a current ${category} failure and recovers`, async () => {
      const { client, controller } = fixture()
      await controller.list.refresh()
      controller.list.select(controller.list.snapshot().rows)
      expect(controller.projectionVisible).toBe(true)
      vi.mocked(client.list).mockRejectedValueOnce(new DemoRequestError(category))
      await expect(controller.list.refresh()).rejects.toMatchObject({ category })
      expect(controller.failure()).toBe(category)
      expect(controller.projectionVisible).toBe(false)
      expect(controller.can(demoPermissions.write)).toBe(false)
      expect(controller.list.snapshot()).toMatchObject({ rows: [], total: 0, selectedKeys: [] })
      await controller.list.refresh()
      expect(controller.failure()).toBeNull()
      expect(controller.projectionVisible).toBe(true)
      expect(controller.list.snapshot().rows).toHaveLength(1)
    })
  }

  it('does not let stale success overwrite a current failure', async () => {
    const { client, controller } = fixture()
    const stale = deferred<{ rows: Product[]; total: number }>()
    const current = deferred<{ rows: Product[]; total: number }>()
    vi.mocked(client.list).mockImplementationOnce(() => stale.promise).mockImplementationOnce(() => current.promise)
    const first = controller.list.refresh()
    const second = controller.list.refresh()
    current.reject(new DemoRequestError('unavailable'))
    await expect(second).rejects.toMatchObject({ category: 'unavailable' })
    stale.resolve({ rows: [product()], total: 1 })
    await first
    expect(controller.failure()).toBe('unavailable')
    expect(controller.projectionVisible).toBe(false)
    expect(controller.list.snapshot().rows).toEqual([])
  })

  it('does not let stale failure overwrite a current success', async () => {
    const { client, controller } = fixture()
    const stale = deferred<{ rows: Product[]; total: number }>()
    const current = deferred<{ rows: Product[]; total: number }>()
    vi.mocked(client.list).mockImplementationOnce(() => stale.promise).mockImplementationOnce(() => current.promise)
    const first = controller.list.refresh()
    const second = controller.list.refresh()
    current.resolve({ rows: [product()], total: 1 })
    await second
    stale.reject(new DemoRequestError('forbidden'))
    await first
    expect(controller.failure()).toBeNull()
    expect(controller.projectionVisible).toBe(true)
    expect(controller.list.snapshot().rows).toHaveLength(1)
  })

  it('fails closed before the first projection when read permission is absent', async () => {
    const { client } = fixture()
    const controller = createDemoController(client, vi.fn(async () => true), { can: () => false })
    await expect(controller.list.refresh()).rejects.toMatchObject({ category: 'forbidden' })
    expect(client.list).not.toHaveBeenCalled()
    expect(controller.projectionVisible).toBe(false)
    expect(controller.can(demoPermissions.read)).toBe(false)
    expect(controller.list.snapshot().rows).toEqual([])
  })

  for (const test of [
    { name: 'save forbidden', category: 'forbidden' as const, run: async (client: DemoClient, controller: ReturnType<typeof createDemoController>) => {
      vi.mocked(client.create).mockRejectedValueOnce(new DemoRequestError('forbidden'))
      return controller.save(input)
    } },
    { name: 'remove relogin', category: 'relogin' as const, run: async (client: DemoClient, controller: ReturnType<typeof createDemoController>) => {
      vi.mocked(client.delete).mockRejectedValueOnce(new DemoRequestError('relogin'))
      return controller.remove([product()])
    } },
  ]) {
    it(`fails the existing projection closed after a DB-final ${test.name}`, async () => {
      const { client, controller } = fixture()
      await controller.list.refresh()
      controller.list.select(controller.list.snapshot().rows)
      expect(controller.list.snapshot().selectedKeys).toHaveLength(1)
      expect(await test.run(client, controller)).toBe('failed')
      expect(controller.failure()).toBe(test.category)
      expect(controller.projectionVisible).toBe(false)
      expect(controller.can(demoPermissions.read)).toBe(false)
      expect(controller.can(demoPermissions.write)).toBe(false)
      expect(controller.list.snapshot()).toMatchObject({ rows: [], total: 0, selectedKeys: [] })
      await controller.list.refresh()
      expect(controller.projectionVisible).toBe(true)
      expect(controller.list.snapshot().selectedKeys).toEqual([])
    })
  }

  it('normalizes search and keeps the stable projection when a local invalid request stales remote work', async () => {
    const { client, controller } = fixture()
    await controller.list.refresh()
    controller.list.select(controller.list.snapshot().rows)
    vi.mocked(client.list).mockClear()

    await expect(controller.list.search({ search: '😀'.repeat(101) })).rejects.toMatchObject({ category: 'validation' })
    expect(client.list).not.toHaveBeenCalled()
    expect(controller.projectionVisible).toBe(true)
    expect(controller.list.snapshot()).toMatchObject({ filters: { search: '' }, total: 1, selectedKeys: [product().id] })

    await controller.list.search({ search: '   ' })
    expect(client.list).toHaveBeenLastCalledWith(expect.objectContaining({ search: '' }))
    await controller.list.search({ search: ' demo ' })
    expect(client.list).toHaveBeenLastCalledWith(expect.objectContaining({ search: 'demo' }))
    expect(controller.list.snapshot().filters.search).toBe('demo')

    controller.list.select(controller.list.snapshot().rows)
    const stale = deferred<{ rows: Product[]; total: number }>()
    vi.mocked(client.list).mockImplementationOnce(() => stale.promise)
    const remote = controller.list.search({ search: 'future' })
    const calls = vi.mocked(client.list).mock.calls.length
    await expect(controller.list.search({ search: '😀'.repeat(101) })).rejects.toMatchObject({ category: 'validation' })
    expect(client.list).toHaveBeenCalledTimes(calls)
    expect(controller.list.snapshot()).toMatchObject({ filters: { search: 'demo' }, total: 1, selectedKeys: [product().id], loading: false })
    stale.resolve({ rows: [product('00000000-0000-4000-8000-000000000099')], total: 1 })
    await remote
    expect(controller.failure()).toBe('validation')
    expect(controller.list.snapshot()).toMatchObject({ filters: { search: 'demo' }, selectedKeys: [product().id] })
  })
})
