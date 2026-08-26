import { DemoRequestError, demoPermissions, emptyProduct, validateProduct, type DeleteTarget, type DemoClient, type DemoFailure, type DemoPermissionCode, type Product, type ProductInput } from '@go-admin/domain-demo'
import { createListController, type ListController } from '@go-admin/ui'

export type MutationResult = 'completed' | 'invalid' | 'cancelled' | 'empty' | 'busy' | 'failed' | 'refresh-failed'
export interface DemoFilters { readonly search: string }
export interface DemoCapabilityPort { can(permissionCode: DemoPermissionCode): boolean }
export type CompletedIntent = 'save' | 'remove'

export interface DemoController {
  readonly list: ListController<DemoFilters, Product, string>
  readonly busy: boolean
  readonly pendingRepair: boolean
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
  const record = (error: unknown) => { failure = error instanceof DemoRequestError ? error.category : 'unavailable' }
  const list = createListController<DemoFilters, Product, string>({
    initialFilters: () => ({ search: '' }), rowKey: row => row.id,
    load: async request => {
      failure = null
      if (!capabilities.can(demoPermissions.read)) { failure = 'forbidden'; throw new DemoRequestError('forbidden') }
      try { return await client.list({ search: request.filters.search, page: request.page, pageSize: request.pageSize,
        sort: (request.sort?.key ?? 'updatedAt') as 'sku'|'name'|'priceCents'|'updatedAt', direction: request.sort?.direction ?? 'descending' }) }
      catch (error) { record(error); throw error }
    },
  })
  const refresh = async (): Promise<MutationResult> => {
    failure = null
    try { await list.refresh(); completion = pending; pending = null; return 'completed' }
    catch (error) { record(error); return 'refresh-failed' }
  }
  return {
    list,
    get busy() { return mutationBusy || repairBusy },
    get pendingRepair() { return pending !== null },
    failure: () => failure,
    clearFailure: () => { failure = null },
    can: permissionCode => capabilities.can(permissionCode),
    takeCompletion() { const value = completion; completion = null; return value },
    empty: emptyProduct,
    async save(model) {
      if (mutationBusy || repairBusy) return 'busy'
      if (pending) return 'refresh-failed'
      if (!capabilities.can(demoPermissions.write)) { failure = 'forbidden'; return 'failed' }
      if (Object.keys(validateProduct(model)).length > 0 || (model.id !== undefined && (!Number.isSafeInteger(model.revision) || (model.revision ?? 0) < 1))) return 'invalid'
      mutationBusy = true; failure = null
      try {
        try {
          const input = { sku: model.sku.trim().toUpperCase(), name: model.name.trim(), description: model.description.trim(), priceCents: model.priceCents, status: model.status }
          if (model.id) await client.update(model.id, { ...input, revision: model.revision! }); else await client.create(input)
        } catch (error) { record(error); return 'failed' }
        pending = 'save'
        return await refresh()
      } finally { mutationBusy = false }
    },
    async remove(products) {
      if (mutationBusy || repairBusy) return 'busy'
      if (pending) return 'refresh-failed'
      if (!capabilities.can(demoPermissions.delete)) { failure = 'forbidden'; return 'failed' }
      if (products.length === 0) return 'empty'
      mutationBusy = true; failure = null
      try {
        if (!await confirm(products.length)) return 'cancelled'
        const targets: DeleteTarget[] = products.map(({ id, revision }) => ({ id, revision }))
        try { await client.delete(targets) } catch (error) { record(error); return 'failed' }
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
