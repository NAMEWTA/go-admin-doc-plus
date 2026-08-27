import { createListController, type ListController } from '@go-admin/ui'
import { GeneratorRequestError, defaultColumns, generatorPermissions, validateDraft, type ColumnDraft, type GenerationDraft, type GenerationPreview, type GenerationResult, type GeneratorClient, type GeneratorFailure, type TableMetadata, type TableReference } from '@go-admin/domain-generator'

export interface GeneratorCapabilityPort { can(permissionCode: string): boolean }
export type WizardStep = 'source' | 'configure' | 'preview' | 'complete'
export interface TableFilters { search: string }
export interface GeneratorController {
  readonly tables: ListController<TableFilters, TableReference, string>
  readonly step: WizardStep
  readonly busy: boolean
  readonly selected: TableMetadata | null
  readonly draft: GenerationDraft | null
  readonly previewValue: GenerationPreview | null
  readonly result: GenerationResult | null
  readonly projectionVisible: boolean
  failure(): GeneratorFailure | null
  select(table: TableReference): Promise<void>
  setNames(module: string, entity: string, plural: string): void
  configureColumn(name: string, changes: Partial<Omit<ColumnDraft, 'name'>>): void
  createPreview(): Promise<'completed'|'invalid'|'failed'|'busy'>
  confirmWrite(confirmed: boolean): Promise<'completed'|'invalid'|'failed'|'busy'>
  reset(): void
}

export const createGeneratorController = (client: GeneratorClient, capabilities: GeneratorCapabilityPort): GeneratorController => {
  let step: WizardStep = 'source', busy = false, selected: TableMetadata | null = null, draft: GenerationDraft | null = null
  let previewValue: GenerationPreview | null = null, result: GenerationResult | null = null, failure: GeneratorFailure | null = null
  let requestSequence = 0, projectionVisible = false
  const hide = () => { projectionVisible = false; tables.clearSelection() }
  const record = (error: unknown) => { failure = error instanceof GeneratorRequestError ? error.category : 'unavailable'; if (failure === 'relogin' || failure === 'forbidden' || failure === 'unavailable') hide() }
  const tables = createListController<TableFilters, TableReference, string>({
    initialFilters: () => ({ search: '' }), normalizeFilters: filters => ({ search: filters.search.trim().toLowerCase() }),
    validate: request => { if ([...request.filters.search].length > 100) throw new GeneratorRequestError('validation') },
    rowKey: table => `${table.schema}.${table.name}`,
    load: async request => {
      const sequence = ++requestSequence; failure = null
      try {
        if (!capabilities.can(generatorPermissions.metadata)) throw new GeneratorRequestError('forbidden')
        const all = await client.listTables(); const search = request.filters.search
        const filtered = search ? all.filter(table => `${table.schema}.${table.name}`.toLowerCase().includes(search)) : all
        if (sequence === requestSequence) projectionVisible = true
        const start = (request.page-1)*request.pageSize
        return { rows: filtered.slice(start, start+request.pageSize), total: filtered.length }
      } catch (error) { if (sequence === requestSequence) record(error); throw error }
    },
  })
  const controller: GeneratorController = {
    tables,
    get step() { return step }, get busy() { return busy }, get selected() { return selected }, get draft() { return draft },
    get previewValue() { return previewValue }, get result() { return result },
    get projectionVisible() { return projectionVisible && capabilities.can(generatorPermissions.metadata) },
    failure: () => failure,
    async select(table) {
      if (busy || !capabilities.can(generatorPermissions.metadata)) { record(new GeneratorRequestError('forbidden')); return }
      busy = true; failure = null
      try {
        selected = await client.describe(table)
        const singular = table.name.endsWith('s') ? table.name.slice(0, -1) : table.name
        draft = { module: singular.replaceAll('_', ''), entity: singular.split('_').map(part => part[0]!.toUpperCase()+part.slice(1)).join(''), plural: table.name.replaceAll('_', '-'), table: selected.table, columns: defaultColumns(selected) }
        previewValue = null; result = null; step = 'configure'
      } catch (error) { record(error) } finally { busy = false }
    },
    setNames(module, entity, plural) { if (draft && !busy) draft = { ...draft, module, entity, plural } },
    configureColumn(name, changes) { if (draft && !busy) draft = { ...draft, columns: draft.columns.map(column => column.name === name ? { ...column, ...changes } : column) } },
    async createPreview() {
      if (busy) return 'busy'
      if (!draft || Object.keys(validateDraft(draft)).length > 0) { failure = 'validation'; return 'invalid' }
      if (!capabilities.can(generatorPermissions.preview)) { record(new GeneratorRequestError('forbidden')); return 'failed' }
      busy = true; failure = null
      try { previewValue = await client.preview(draft); result = null; step = 'preview'; return 'completed' }
      catch (error) { record(error); return 'failed' } finally { busy = false }
    },
    async confirmWrite(confirmed) {
      if (busy) return 'busy'
      if (!confirmed || !previewValue) { failure = 'validation'; return 'invalid' }
      if (!capabilities.can(generatorPermissions.write)) { record(new GeneratorRequestError('forbidden')); return 'failed' }
      busy = true; failure = null
      const token = previewValue.token
      try { result = await client.write(token); previewValue = null; step = 'complete'; return 'completed' }
      catch (error) { previewValue = null; record(error); step = 'configure'; return 'failed' } finally { busy = false }
    },
    reset() { if (!busy) { selected = null; draft = null; previewValue = null; result = null; failure = null; step = 'source'; tables.clearSelection() } },
  }
  return controller
}
