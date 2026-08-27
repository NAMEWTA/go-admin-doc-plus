import { describe, expect, it } from 'vitest'
import { defaultColumns, validateDraft, type TableMetadata } from './generator'

const table: TableMetadata = { table: { schema: 'main', name: 'products' }, columns: [
  { name: 'id', databaseType: 'TEXT', kind: 'uuid', nullable: false, primaryKey: true, ordinal: 1 },
  { name: 'display_name', databaseType: 'TEXT', kind: 'string', nullable: false, primaryKey: false, ordinal: 2 },
] }

describe('generator domain', () => {
  it('derives deterministic editable column defaults', () => {
    expect(defaultColumns(table)).toEqual([
      { name: 'id', field: 'ID', include: true, searchable: false, sortable: true },
      { name: 'display_name', field: 'DisplayName', include: true, searchable: true, sortable: false },
    ])
  })
  it('rejects path-like identifiers and excluded query columns', () => {
    const columns = defaultColumns(table)
    expect(validateDraft({ module: '../bad', entity: 'Product', plural: 'products', table: table.table, columns })).toHaveProperty('module')
    expect(validateDraft({ module: 'catalog', entity: 'Product', plural: 'products', table: table.table, columns: [{ ...columns[0]!, include: false }] })).toHaveProperty('columns')
  })
})
