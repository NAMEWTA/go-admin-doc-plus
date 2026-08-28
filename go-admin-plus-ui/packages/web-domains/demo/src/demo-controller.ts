import { DemoRequestError, demoPermissions, emptyProduct, validateProduct, validateProductSearch, type DeleteTarget, type DemoClient, type DemoFailure, type DemoPermissionCode, type Product, type ProductInput } from '@go-admin-plus/domain-demo'
import { createListController, type ListController } from '@go-admin-plus/ui'

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
  const record = (error: unknown) => { failure = error instanceof DemoRequestError ? error.category : 'unavailable' }
  let rawList!: ListController<DemoFilters, Product, string>
  const hiddenProjection = () => { projectionVisible = false; rawList.clearSelection() }
  const failMutation = (error: unknown) => {
    record(error)
    if (failure === 'relogin' || failure === 'forbidden' || failure === 'unavailable') {
      requestSequence += 1
      hiddenProjection()
    }
  }
  rawList = createListController<DemoFilters, Product, string>({
    initialFilters: () => ({ search: '' }),
    normalizeFilters: filters => ({ search: filters.search.trim() }),
    validate: request => {
      if (!validateProductSearch(request.filters.search)) {
        requestSequence += 1
        failure = 'validation'
        throw new DemoRequestError('validation')
      }
    },
    rowKey: row => row.id,
    load: async request => {
      const sequence = ++requestSequence
      failure = null
      try {
        if (!capabilities.can(demoPermissions.read)) throw new DemoRequestError('forbidden')
        const result = await client.list({ search: request.filters.search, page: request.page, pageSize: request.pageSize,
          sort: (request.sort?.key ?? 'updatedAt') as 'sku'|'name'|'priceCents'|'updatedAt', direction: request.sort?.direction ?? 'descending' })
        if (sequence === requestSequence) { failure = null; projectionVisible = true; projectionGeneration += 1 }
        return result
      } catch (error) {
        if (sequence === requestSequence) { record(error); hiddenProjection() }
        throw error
      }
    },
  })
  const list: ListController<DemoFilters, Product, string> = {
    snapshot() {
      const snapshot = rawList.snapshot()
      if (projectionVisible && capabilities.can(demoPermissions.read)) return snapshot
      return { ...snapshot, rows: [], total: 0, selectedKeys: [] }
    },
    refresh: () => rawList.refresh(),
    search: filters => rawList.search(filters),
    reset: () => rawList.reset(),
    setPage: page => rawList.setPage(page),
    setPageSize: pageSize => rawList.setPageSize(pageSize),
    setSort: sort => rawList.setSort(sort),
    select(rows) { if (projectionVisible && capabilities.can(demoPermissions.read)) rawList.select(rows) },
    clearSelection: () => rawList.clearSelection(),
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
