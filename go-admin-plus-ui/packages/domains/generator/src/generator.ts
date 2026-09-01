export const generatorPermissions = {
  metadata: 'generator.metadata.read',
  preview: 'generator.preview',
  write: 'generator.write',
} as const

export type ColumnKind = 'string' | 'int64' | 'boolean' | 'decimal' | 'time' | 'uuid' | 'bytes'
export type GeneratorFailure = 'relogin' | 'forbidden' | 'validation' | 'not-found' | 'conflict' | 'gate' | 'unavailable'

export interface TableReference { readonly schema: string; readonly name: string }
export interface ColumnMetadata { readonly name: string; readonly databaseType: string; readonly kind: ColumnKind; readonly nullable: boolean; readonly primaryKey: boolean; readonly ordinal: number }
export interface TableMetadata { readonly table: TableReference; readonly columns: ReadonlyArray<ColumnMetadata> }
export interface ColumnDraft { readonly name: string; readonly field: string; readonly include: boolean; readonly searchable: boolean; readonly sortable: boolean }
export interface GenerationDraft { readonly module: string; readonly entity: string; readonly plural: string; readonly table: TableReference; readonly columns: ReadonlyArray<ColumnDraft> }
export interface PreviewFile { readonly path: string; readonly content: string; readonly sha256: string }
export interface GenerationPreview { readonly token: string; readonly digest: string; readonly module: string; readonly createdAt: string; readonly expiresAt: string; readonly files: ReadonlyArray<PreviewFile> }
export interface GenerationResult { readonly token: string; readonly directory: string; readonly files: ReadonlyArray<string> }

export interface GeneratorClient {
  getConfig(module: string): Promise<{ draft: GenerationDraft; previewDigest: string }>
  listTables(): Promise<ReadonlyArray<TableReference>>
  describe(table: TableReference): Promise<TableMetadata>
  preview(draft: GenerationDraft): Promise<GenerationPreview>
  write(previewToken: string): Promise<GenerationResult>
}

export class GeneratorRequestError extends Error {
  readonly category: GeneratorFailure
  readonly traceId?: string
  constructor(category: GeneratorFailure, traceId?: string) {
    super(`generator request failed: ${category}`)
    this.category = category
    if (traceId !== undefined) this.traceId = traceId
  }
}

const databaseIdentifier = /^[a-z][a-z0-9_]{0,62}$/
const moduleIdentifier = /^[a-z][a-z0-9]{1,31}$/
const codeIdentifier = /^[A-Z][A-Za-z0-9]{0,63}$/
const pathWord = /^[a-z][a-z0-9-]{1,63}$/

export const defaultColumns = (table: TableMetadata): ColumnDraft[] => table.columns.map(column => ({
  name: column.name,
  field: pascalCase(column.name),
  include: true,
  searchable: column.kind === 'string' && !column.primaryKey,
  sortable: column.primaryKey,
}))

export const validateDraft = (draft: GenerationDraft): Record<string, string> => {
  const errors: Record<string, string> = {}
  if (!moduleIdentifier.test(draft.module)) errors.module = 'invalid'
  if (!codeIdentifier.test(draft.entity)) errors.entity = 'invalid'
  if (!pathWord.test(draft.plural)) errors.plural = 'invalid'
  if (!databaseIdentifier.test(draft.table.schema) || !databaseIdentifier.test(draft.table.name)) errors.table = 'invalid'
  const names = new Set<string>(), fields = new Set<string>()
  let primaryKeys = 0
  for (const column of draft.columns) {
    if (!databaseIdentifier.test(column.name) || !codeIdentifier.test(column.field) || names.has(column.name) || fields.has(column.field)) errors.columns = 'invalid'
    names.add(column.name); fields.add(column.field)
    if (!column.include && (column.searchable || column.sortable)) errors.columns = 'invalid'
    if (column.include && column.field === 'ID') primaryKeys += 1
  }
  if (draft.columns.length === 0 || primaryKeys > 1) errors.columns = 'invalid'
  return errors
}

const pascalCase = (value: string): string => value.split('_').filter(Boolean).map(part => part.toLowerCase() === 'id' ? 'ID' : part[0]!.toUpperCase()+part.slice(1)).join('')
