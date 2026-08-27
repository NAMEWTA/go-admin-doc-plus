import { buildDepartmentTree, OrganizationRequestError, validOrganizationName, validOrganizationSearch, type Department, type DepartmentInput, type DepartmentTreeNode, type OrganizationClient, type Position, type PositionInput } from '@go-admin/domain-organization'
import { createListController, type ListController } from '@go-admin/ui'

export interface PositionFilters { search: string }
export type OrganizationFailure = 'relogin' | 'forbidden' | 'validation' | 'not-found' | 'conflict' | 'unavailable'
export type OrganizationCommandResult = 'completed' | 'cancelled' | 'busy' | 'invalid' | 'failed' | 'refresh-failed'
export interface OrganizationCapabilityPort { can(permission: string): boolean; scope(): 'self' | 'all' | null }

export interface OrganizationController {
  positions: ListController<PositionFilters, Position, string>
  departments(): ReadonlyArray<Department>
  departmentTree(): ReadonlyArray<DepartmentTreeNode>
  refreshDepartments(): Promise<void>
  can(permission: string): boolean
  failure(): OrganizationFailure | null
  clearFailure(): void
  hasPendingRepair(): boolean
  repairProjection(): Promise<OrganizationCommandResult>
  readonly busy: boolean
  createDepartment(input: DepartmentInput): Promise<OrganizationCommandResult>
  updateDepartment(id: string, input: DepartmentInput): Promise<OrganizationCommandResult>
  deleteDepartment(id: string): Promise<OrganizationCommandResult>
  createPosition(input: PositionInput): Promise<OrganizationCommandResult>
  updatePosition(id: string, input: PositionInput): Promise<OrganizationCommandResult>
  deletePosition(id: string): Promise<OrganizationCommandResult>
}

export const createOrganizationController = (
  client: OrganizationClient,
  capabilities: OrganizationCapabilityPort,
  confirm: (count: number) => Promise<boolean>,
): OrganizationController => {
  let departments: ReadonlyArray<Department> = []
  let departmentProjectionVisible = false
  let positionProjectionVisible = false
  let failure: OrganizationFailure | null = null
  let mutationBusy = false
  let repairBusy = false
  let departmentRequestSequence = 0
  let positionRequestSequence = 0
  const pendingRepairs = new Map<string, () => Promise<void>>()
  const clearFailure = () => { failure = null }
  const recordFailure = (error: unknown) => {
    const category = error instanceof OrganizationRequestError ? error.category : 'unavailable'
    failure = isFailure(category) ? category : 'unavailable'
  }
  const hideDepartments = () => { departmentProjectionVisible = false; departments = [] }
  let rawPositions!: ListController<PositionFilters, Position, string>
  const hidePositions = () => { positionProjectionVisible = false; rawPositions.clearSelection() }
  const hideAll = () => { hideDepartments(); hidePositions() }
  const canManage = (permission: string) => {
    if (capabilities.scope() !== 'all') {
      hideAll()
      return false
    }
    const allowed = capabilities.can(permission)
    if (!allowed && permission === 'organization.departments.read') hideDepartments()
    if (!allowed && permission === 'organization.positions.read') hidePositions()
    return allowed
  }
  rawPositions = createListController<PositionFilters, Position, string>({
    initialFilters: () => ({ search: '' }),
    normalizeFilters: filters => ({ search: filters.search.trim() }),
    validate: request => {
      if (!validOrganizationSearch(request.filters.search)) {
        positionRequestSequence += 1
        failure = 'validation'
        throw new OrganizationRequestError('validation')
      }
    },
    rowKey: (row) => row.id,
    load: async ({ filters, page, pageSize }) => {
      const sequence = ++positionRequestSequence
      failure = null
      try {
        if (!canManage('organization.positions.read')) throw new OrganizationRequestError('forbidden')
        const result = await client.listPositions(filters.search, page, pageSize)
        if (sequence === positionRequestSequence) positionProjectionVisible = true
        return result
      } catch (error) {
        if (sequence === positionRequestSequence) { hidePositions(); recordFailure(error) }
        throw error
      }
    },
  })
  const positions: ListController<PositionFilters, Position, string> = {
    snapshot: () => {
      const snapshot = rawPositions.snapshot()
      return positionProjectionVisible && canManage('organization.positions.read') ? snapshot : { ...snapshot, rows: [], total: 0, selectedKeys: [] }
    },
    refresh: () => rawPositions.refresh(),
    search: (filters) => rawPositions.search(filters),
    reset: () => rawPositions.reset(),
    setPage: (page) => rawPositions.setPage(page),
    setPageSize: (pageSize) => rawPositions.setPageSize(pageSize),
    setSort: (sort) => rawPositions.setSort(sort),
    select: (rows) => { if (positionProjectionVisible && canManage('organization.positions.read')) rawPositions.select(rows) },
    clearSelection: () => rawPositions.clearSelection(),
  }
  const refreshDepartments = async () => {
    const sequence = ++departmentRequestSequence
    clearFailure()
    try {
      if (!canManage('organization.departments.read')) throw new OrganizationRequestError('forbidden')
      const next = await client.listDepartments()
      buildDepartmentTree(next)
      if (sequence !== departmentRequestSequence) return
      departments = [...next]
      departmentProjectionVisible = true
    } catch (error) {
      if (sequence === departmentRequestSequence) { hideDepartments(); recordFailure(error) }
      throw error
    }
  }
  const refreshAll = async () => {
    await refreshDepartments()
    await positions.refresh()
  }
  const command = async (
    key: string,
    operation: () => Promise<void>,
    refresh: () => Promise<void>,
    options: { destructive?: boolean; valid?: boolean } = {},
  ): Promise<OrganizationCommandResult> => {
    if (mutationBusy || repairBusy) return 'busy'
    if (pendingRepairs.size > 0) return 'refresh-failed'
    if (options.valid === false) return 'invalid'
    mutationBusy = true
    clearFailure()
    try {
      if (options.destructive && !await confirm(1)) return 'cancelled'
      try { await operation() } catch (error) {
        recordFailure(error)
        if (failure === 'relogin' || failure === 'forbidden' || failure === 'unavailable') {
          departmentRequestSequence += 1; positionRequestSequence += 1; hideAll()
        }
        return 'failed'
      }
      pendingRepairs.set(key, refresh)
      try { await refresh(); pendingRepairs.delete(key); return 'completed' }
      catch (error) { recordFailure(error); return 'refresh-failed' }
    } finally { mutationBusy = false }
  }
  const validDepartment = (input: DepartmentInput) => validStableKey(input.key) && validOrganizationName(input.name) && input.parentId.length > 0 && Number.isSafeInteger(input.sortOrder) && input.sortOrder >= -1_000_000 && input.sortOrder <= 1_000_000
  const validPosition = (input: PositionInput) => validStableKey(input.key) && validOrganizationName(input.name) && input.departmentId.length > 0
  return {
    positions,
    departments: () => departmentProjectionVisible && canManage('organization.departments.read') ? [...departments] : [],
    departmentTree: () => departmentProjectionVisible && canManage('organization.departments.read') ? buildDepartmentTree(departments) : [],
    refreshDepartments,
    can: canManage,
    failure: () => failure,
    clearFailure,
    hasPendingRepair: () => pendingRepairs.size > 0,
    async repairProjection() {
      if (repairBusy || mutationBusy) return 'busy'
      const pending = pendingRepairs.entries().next().value as [string, () => Promise<void>] | undefined
      if (!pending) return 'completed'
      repairBusy = true
      clearFailure()
      try {
        try { await pending[1](); pendingRepairs.delete(pending[0]); return 'completed' }
        catch (error) { recordFailure(error); return 'refresh-failed' }
      } finally { repairBusy = false }
    },
    get busy() { return mutationBusy || repairBusy },
    createDepartment(input) { return command(`department:create:${input.key}`, async () => { await client.createDepartment(input) }, refreshDepartments, { valid: validDepartment(input) }) },
    updateDepartment(id, input) { return command(`department:update:${id}`, async () => { await client.updateDepartment(id, input) }, refreshDepartments, { valid: id.length > 0 && validDepartment(input) }) },
    deleteDepartment(id) { return command(`department:delete:${id}`, () => client.deleteDepartment(id), refreshAll, { destructive: true, valid: id.length > 0 }) },
    createPosition(input) { return command(`position:create:${input.key}`, async () => { await client.createPosition(input) }, () => positions.refresh(), { valid: validPosition(input) }) },
    updatePosition(id, input) { return command(`position:update:${id}`, async () => { await client.updatePosition(id, input) }, () => positions.refresh(), { valid: id.length > 0 && validPosition(input) }) },
    deletePosition(id) { return command(`position:delete:${id}`, () => client.deletePosition(id), () => positions.refresh(), { destructive: true, valid: id.length > 0 }) },
  }
}

export const settleOrganizationPageOperation = async (operation: () => Promise<unknown>, settled: () => void) => {
  try { await operation() }
  catch { /* The controller owns stable failure classification. */ }
  finally { settled() }
}

const validStableKey = (value: string) => value.length >= 3 && value.length <= 64 && /^[a-z0-9][a-z0-9_-]*$/.test(value)
const isFailure = (value: string): value is OrganizationFailure => ['relogin', 'forbidden', 'validation', 'not-found', 'conflict', 'unavailable'].includes(value)
