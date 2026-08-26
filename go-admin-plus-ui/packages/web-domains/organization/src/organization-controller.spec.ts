import { describe, expect, it, vi } from 'vitest'
import { OrganizationRequestError, type Department, type OrganizationClient } from '@go-admin/domain-organization'
import { createOrganizationController } from './organization-controller'

const root: Department = { id: 'department-root-001', key: 'root', name: 'Organization', protected: true, sortOrder: 0 }

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
    const controller = createOrganizationController(api, () => true, async () => true)
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
    const controller = createOrganizationController(api, () => true, async () => true)
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
    const controller = createOrganizationController(api, () => true, async () => true)
    await controller.positions.refresh()
    const first = controller.deletePosition('p1')
    expect(await controller.deletePosition('p2')).toBe('busy')
    release()
    expect(await first).toBe('completed')
    expect(api.deletePosition).toHaveBeenCalledTimes(1)
  })
})
