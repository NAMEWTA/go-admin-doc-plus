import { DemoRequestError, demoPermissions, emptyProduct, validateProduct, validateProductSearch, type DeleteTarget, type DemoClient, type DemoFailure, type DemoPermissionCode, type Product, type ProductInput } from '@go-admin/domain-demo'
import type { ListController, ListRequest, SortState } from '@go-admin/ui'

export type MutationResult = 'completed' | 'invalid' | 'cancelled' | 'empty' | 'busy' | 'failed' | 'refresh-failed'
export interface DemoFilters { readonly search: string }
export interface DemoCapabilityPort { can(permissionCode: DemoPermissionCode): boolean }
export type CompletedIntent = 'save' | 'remove'

export interface DemoController {
  readonly list: ListController<DemoFilters, Product, string>
  readonly busy: boolean
  readonly pendingRepair: boolean
  readonly projectionVisible: boolean
  failure(): DemoFailure | null
  clearFailure(): void
  can(permissionCode: DemoPermissionCode): boolean
  takeCompletion(): CompletedIntent | null
  save(model: ProductInput & { id?: string; revision?: number }): Promise<MutationResult>
  remove(products: ReadonlyArray<Product>): Promise<MutationResult>
  repairProjection(): Promise<MutationResult>
  empty(): ProductInput
}

export const createDemoController = (client: DemoClient, confirm: (count: number) => Promise<boolean>, capabilities: DemoCapabilityPort): DemoController => {
  let mutationBusy = false
  let repairBusy = false
  let pending: CompletedIntent | null = null
  let completion: CompletedIntent | null = null
  let failure: DemoFailure | null = null
  let projectionVisible = false
  let projectionGeneration = 0
  let requestSequence = 0
  let filters: DemoFilters = { search: '' }
  let page = 1
  let pageSize = 20
  let sort: SortState | undefined
  let rows: ReadonlyArray<Product> = []
  let total = 0
  let selectedKeys: ReadonlyArray<string> = []
  let loading = false
  const record = (error: unknown) => { failure = error instanceof DemoRequestError ? error.category : 'unavailable' }
  const hiddenProjection = () => { projectionVisible = false; selectedKeys = [] }
  const failMutation = (error: unknown) => {
    requestSequence += 1
    loading = false
    record(error)
    hiddenProjection()
  }
  const requirePositiveInteger = (value: number, name: string) => {
    if (!Number.isSafeInteger(value) || value < 1) throw new RangeError(`${name} must be a positive integer`)
  }
  const load = async (request: ListRequest<DemoFilters>): Promise<void> => {
    const sequence = ++requestSequence
    const normalized = { ...request, filters: { search: request.filters.search.trim() } }
    failure = null
    if (!validateProductSearch(normalized.filters.search)) {
      loading = false
      failure = 'validation'
      throw new DemoRequestError('validation')
    }
    loading = true
    try {
      if (!capabilities.can(demoPermissions.read)) throw new DemoRequestError('forbidden')
      const result = await client.list({ search: normalized.filters.search, page: normalized.page, pageSize: normalized.pageSize,
        sort: (normalized.sort?.key ?? 'updatedAt') as 'sku'|'name'|'priceCents'|'updatedAt', direction: normalized.sort?.direction ?? 'descending' })
      if (sequence !== requestSequence) return
      filters = normalized.filters; page = normalized.page; pageSize = normalized.pageSize; sort = normalized.sort
      rows = [...result.rows]; total = result.total; selectedKeys = []; failure = null; projectionVisible = true; projectionGeneration += 1
    } catch (error) {
      if (sequence !== requestSequence) return
      record(error); hiddenProjection()
      throw error
    } finally {
      if (sequence === requestSequence) loading = false
    }
  }
  const list: ListController<DemoFilters, Product, string> = {
    snapshot() {
      const visible = projectionVisible && capabilities.can(demoPermissions.read)
      return { filters: { ...filters }, page, pageSize, ...(sort ? { sort: { ...sort } } : {}), rows: visible ? [...rows] : [],
        total: visible ? total : 0, selectedKeys: visible ? [...selectedKeys] : [], loading }
    },
    refresh: () => load({ filters, page, pageSize, ...(sort ? { sort } : {}) }),
    search: nextFilters => load({ filters: nextFilters, page: 1, pageSize, ...(sort ? { sort } : {}) }),
    reset: () => load({ filters: { search: '' }, page: 1, pageSize }),
    setPage(nextPage) { requirePositiveInteger(nextPage, 'page'); return load({ filters, page: nextPage, pageSize, ...(sort ? { sort } : {}) }) },
    setPageSize(nextPageSize) { requirePositiveInteger(nextPageSize, 'pageSize'); return load({ filters, page: 1, pageSize: nextPageSize, ...(sort ? { sort } : {}) }) },
    setSort(nextSort) { return load({ filters, page: 1, pageSize, ...(nextSort ? { sort: nextSort } : {}) }) },
    select(nextRows) { if (projectionVisible && capabilities.can(demoPermissions.read)) selectedKeys = nextRows.map(row => row.id) },
    clearSelection() { selectedKeys = [] },
  }
  const refresh = async (): Promise<MutationResult> => {
    failure = null
    const previousGeneration = projectionGeneration
    try {
      await list.refresh()
      if (!projectionVisible || projectionGeneration === previousGeneration) return 'refresh-failed'
      completion = pending; pending = null; return 'completed'
    }
    catch (error) { record(error); return 'refresh-failed' }
  }
  return {
    list,
    get busy() { return mutationBusy || repairBusy },
    get pendingRepair() { return pending !== null },
    get projectionVisible() { return projectionVisible && capabilities.can(demoPermissions.read) },
    failure: () => failure,
    clearFailure: () => { failure = null },
    can: permissionCode => projectionVisible && capabilities.can(demoPermissions.read) && capabilities.can(permissionCode),
    takeCompletion() { const value = completion; completion = null; return value },
    empty: emptyProduct,
    async save(model) {
      if (mutationBusy || repairBusy) return 'busy'
      if (pending) return 'refresh-failed'
      if (!projectionVisible || !capabilities.can(demoPermissions.write)) { failMutation(new DemoRequestError('forbidden')); return 'failed' }
      if (Object.keys(validateProduct(model)).length > 0 || (model.id !== undefined && (!Number.isSafeInteger(model.revision) || (model.revision ?? 0) < 1))) return 'invalid'
      mutationBusy = true; failure = null
      try {
        try {
          const input = { sku: model.sku.trim().toUpperCase(), name: model.name.trim(), description: model.description.trim(), priceCents: model.priceCents, status: model.status }
          if (model.id) await client.update(model.id, { ...input, revision: model.revision! }); else await client.create(input)
        } catch (error) { failMutation(error); return 'failed' }
        pending = 'save'
        return await refresh()
      } finally { mutationBusy = false }
    },
    async remove(products) {
      if (mutationBusy || repairBusy) return 'busy'
      if (pending) return 'refresh-failed'
      if (!projectionVisible || !capabilities.can(demoPermissions.delete)) { failMutation(new DemoRequestError('forbidden')); return 'failed' }
      if (products.length === 0) return 'empty'
      mutationBusy = true; failure = null
      try {
        if (!await confirm(products.length)) return 'cancelled'
        const targets: DeleteTarget[] = products.map(({ id, revision }) => ({ id, revision }))
        try { await client.delete(targets) } catch (error) { failMutation(error); return 'failed' }
        list.clearSelection(); pending = 'remove'
        return await refresh()
      } finally { mutationBusy = false }
    },
    async repairProjection() {
      if (mutationBusy || repairBusy) return 'busy'
      if (!pending) return 'completed'
      repairBusy = true
      try { return await refresh() } finally { repairBusy = false }
    },
  }
}
