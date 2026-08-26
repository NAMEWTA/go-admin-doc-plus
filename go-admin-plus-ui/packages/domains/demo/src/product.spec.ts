import { describe, expect, it } from 'vitest'
import { emptyProduct, validateProduct } from './product'

describe('demo product domain', () => {
  it('validates the public product contract without Vue', () => {
    expect(validateProduct({ sku: 'DEMO-01', name: 'Demo product', description: '', priceCents: 120, status: 'active' })).toEqual({})
    expect(validateProduct(emptyProduct())).toMatchObject({ sku: 'SKU_INVALID', name: 'NAME_INVALID' })
    expect(validateProduct({ sku: 'DROP TABLE', name: 'Demo product', description: '', priceCents: -1, status: 'active' })).toMatchObject({ sku: 'SKU_INVALID', priceCents: 'PRICE_INVALID' })
  })
})
