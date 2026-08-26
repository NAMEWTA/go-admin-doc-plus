import type { components } from './generated/client'

export type Product = components['schemas']['Product']
export type ProductInput = components['schemas']['ProductInput']
export type ProductPage = components['schemas']['ProductPage']
export type DeleteTarget = components['schemas']['DeleteProductTarget']
export type DemoFailure = 'relogin' | 'forbidden' | 'validation' | 'conflict' | 'not-found' | 'unavailable'
export type DemoPermissionCode = 'demo.products.read' | 'demo.products.write' | 'demo.products.delete'

export const demoPermissions = {
  read: 'demo.products.read',
  write: 'demo.products.write',
  delete: 'demo.products.delete',
} as const satisfies Readonly<Record<string, DemoPermissionCode>>

export interface ProductQuery {
  readonly search: string
  readonly page: number
  readonly pageSize: number
  readonly sort: 'sku' | 'name' | 'priceCents' | 'updatedAt'
  readonly direction: 'ascending' | 'descending'
}

export interface DemoClient {
  list(query: ProductQuery): Promise<ProductPage>
  get(id: string): Promise<Product>
  create(input: ProductInput): Promise<Product>
  update(id: string, input: ProductInput & { revision: number }): Promise<Product>
  delete(targets: ReadonlyArray<DeleteTarget>): Promise<void>
}

export class DemoRequestError extends Error {
  readonly category: DemoFailure
  constructor(category: DemoFailure) { super(category); this.category = category }
}

export const emptyProduct = (): ProductInput => ({ sku: '', name: '', description: '', priceCents: 0, status: 'active' })

export const validateProduct = (input: Readonly<ProductInput>): Readonly<Record<keyof ProductInput, string>> => {
  const errors: Partial<Record<keyof ProductInput, string>> = {}
  if (!/^[A-Z0-9][A-Z0-9_-]{2,31}$/.test(input.sku.trim().toUpperCase())) errors.sku = 'SKU_INVALID'
  if (input.name.trim().length < 3 || input.name.trim().length > 120) errors.name = 'NAME_INVALID'
  if (input.description.length > 500) errors.description = 'DESCRIPTION_INVALID'
  if (!Number.isSafeInteger(input.priceCents) || input.priceCents < 0 || input.priceCents > 100_000_000) errors.priceCents = 'PRICE_INVALID'
  if (input.status !== 'active' && input.status !== 'inactive') errors.status = 'STATUS_INVALID'
  return errors as Readonly<Record<keyof ProductInput, string>>
}
