import { describe, expect, it, vi } from 'vitest'
import { OrganizationRequestError, type Department, type OrganizationClient } from '@go-admin/domain-organization'
import { createOrganizationController } from './organization-controller'

const root: Department = { id: 'department-root-001', key: 'root', name: 'Organization', protected: true, sortOrder: 0 }
const allCapabilities = { can: () => true, scope: () => 'all' as const }
const deferred = <T>() => {
  let resolve!: (value: T) => void
  let reject!: (error: unknown) => void
  const promise = new Promise<T>((success, failure) => { resolve = success; reject = failure })
  return { promise, resolve, reject }
}

const client = (): OrganizationClient => ({
  listDepartments: vi.fn(async () => [root]),
  createDepartment: vi.fn(async (input) => ({ id: 'department-1', protected: false, ...input })),
  updateDepartment: vi.fn(async (id, input) => ({ id, protected: false, ...input })),
  deleteDepartment: vi.fn(async () => undefined),
  listPositions: vi.fn(async () => ({ rows: [], total: 0 })),
  createPosition: vi.fn(async (input) => ({ id: 'position-1', protected: false, ...input })),
  updatePosition: vi.fn(async (id, input) => ({ id, protected: false, ...input })),
  deletePosition: vi.fn(async () => undefined),
})

describe('createOrganizationController', () => {
  it('blocks repeated writes while a completed write awaits projection repair', async () => {
    const api = client()
    let loads = 0
    vi.mocked(api.listDepartments).mockImplementation(async () => {
      loads += 1
      if (loads === 1) return [root]
      if (loads === 2) throw new OrganizationRequestError('unavailable')
      return [root]
    })
    const controller = createOrganizationController(api, allCapabilities, async () => true)
    await controller.refreshDepartments()
    const input = { key: 'ops', name: 'Operations', parentId: root.id, sortOrder: 0 }
    expect(await controller.createDepartment(input)).toBe('refresh-failed')
    expect(controller.departments()).toEqual([])
    expect(await controller.createDepartment({ ...input, key: 'sales' })).toBe('refresh-failed')
    expect(api.createDepartment).toHaveBeenCalledTimes(1)
    expect(await controller.repairProjection()).toBe('completed')
    expect(controller.hasPendingRepair()).toBe(false)
  })

  it('hides stale successful list data after a current failure', async () => {
    const api = client()
    vi.mocked(api.listPositions)
      .mockResolvedValueOnce({ rows: [{ id: 'p1', key: 'lead', name: 'Lead', departmentId: root.id, enabled: true, protected: false }], total: 1 })
      .mockRejectedValueOnce(new OrganizationRequestError('forbidden'))
    const controller = createOrganizationController(api, allCapabilities, async () => true)
    await controller.positions.refresh()
    expect(controller.positions.snapshot().rows).toHaveLength(1)
    await expect(controller.positions.refresh()).rejects.toMatchObject({ category: 'forbidden' })
    expect(controller.positions.snapshot()).toMatchObject({ rows: [], total: 0, selectedKeys: [] })
    expect(controller.failure()).toBe('forbidden')
  })

  it('serializes destructive confirmation and mutation', async () => {
    let release!: () => void
    const api = client()
    vi.mocked(api.deletePosition).mockImplementation(() => new Promise<void>((resolve) => { release = resolve }))
    const controller = createOrganizationController(api, allCapabilities, async () => true)
    await controller.positions.refresh()
    const first = controller.deletePosition('p1')
    expect(await controller.deletePosition('p2')).toBe('busy')
    release()
    expect(await first).toBe('completed')
    expect(api.deletePosition).toHaveBeenCalledTimes(1)
  })

  it('uses the shared list sequence fence for stale position success and failure', async () => {
    const api = client()
    const stale = deferred<{ rows: []; total: number }>()
    const current = deferred<{ rows: []; total: number }>()
    vi.mocked(api.listPositions).mockImplementationOnce(() => stale.promise).mockImplementationOnce(() => current.promise)
    const controller = createOrganizationController(api, allCapabilities, async () => true)
    const first = controller.positions.search({ search: 'old' })
    const second = controller.positions.search({ search: 'new' })
    current.resolve({ rows: [], total: 0 })
    await second
    stale.reject(new OrganizationRequestError('forbidden'))
    await first
    expect(controller.failure()).toBeNull()
    expect(controller.positions.snapshot().filters.search).toBe('new')
  })

  it('normalizes local search validation without calling transport or losing the stable projection', async () => {
    const api = client()
    const controller = createOrganizationController(api, allCapabilities, async () => true)
    await controller.positions.search({ search: '  lead  ' })
    expect(api.listPositions).toHaveBeenLastCalledWith('lead', 1, 20)
    vi.mocked(api.listPositions).mockClear()
    await expect(controller.positions.search({ search: '😀'.repeat(101) })).rejects.toMatchObject({ category: 'validation' })
    expect(api.listPositions).not.toHaveBeenCalled()
    expect(controller.positions.snapshot()).toMatchObject({ filters: { search: 'lead' }, rows: [], total: 0 })
  })

  it('fails all management projections closed for self scope without network access', async () => {
    const api = client()
    const capabilities = { can: () => true, scope: () => 'self' as const }
    const controller = createOrganizationController(api, capabilities, async () => true)
    await expect(controller.refreshDepartments()).rejects.toMatchObject({ category: 'forbidden' })
    await expect(controller.positions.refresh()).rejects.toMatchObject({ category: 'forbidden' })
    expect(api.listDepartments).not.toHaveBeenCalled()
    expect(api.listPositions).not.toHaveBeenCalled()
    expect(controller.can('organization.departments.write')).toBe(false)
    expect(controller.departments()).toEqual([])
    expect(controller.positions.snapshot().rows).toEqual([])
  })

  it('fails an existing controller projection closed as soon as capabilities are revoked', async () => {
    const api = client()
    vi.mocked(api.listPositions).mockResolvedValue({
      rows: [{ id: 'p1', key: 'lead', name: 'Lead', departmentId: root.id, enabled: true, protected: false }],
      total: 1,
    })
    let allowed = true
    const capabilities = { can: () => allowed, scope: () => 'all' as const }
    const controller = createOrganizationController(api, capabilities, async () => true)
    await controller.refreshDepartments()
    await controller.positions.refresh()
    controller.positions.select(controller.positions.snapshot().rows)
    expect(controller.departments()).toHaveLength(1)
    expect(controller.positions.snapshot()).toMatchObject({ total: 1, selectedKeys: ['p1'] })

    allowed = false

    expect(controller.can('organization.departments.read')).toBe(false)
    expect(controller.departments()).toEqual([])
    expect(controller.positions.snapshot()).toMatchObject({ rows: [], total: 0, selectedKeys: [] })
  })
})
