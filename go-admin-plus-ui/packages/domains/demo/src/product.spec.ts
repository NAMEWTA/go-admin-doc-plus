import { describe, expect, it } from 'vitest'
import { codePointLength, emptyProduct, validateProduct, validateProductSearch } from './product'

describe('demo product domain', () => {
  it('validates the public product contract without Vue', () => {
    expect(validateProduct({ sku: 'DEMO-01', name: 'Demo product', description: '', priceCents: 120, status: 'active' })).toEqual({})
    expect(validateProduct(emptyProduct())).toMatchObject({ sku: 'SKU_INVALID', name: 'NAME_INVALID' })
    expect(validateProduct({ sku: 'DROP TABLE', name: 'Demo product', description: '', priceCents: -1, status: 'active' })).toMatchObject({ sku: 'SKU_INVALID', priceCents: 'PRICE_INVALID' })
  })

  it('counts Unicode code points rather than UTF-16 units at every public boundary', () => {
    expect(codePointLength('😀')).toBe(1)
    expect(validateProduct({ sku: 'ASTRAL-01', name: '😀'.repeat(3), description: '界'.repeat(500), priceCents: 1, status: 'active' })).toEqual({})
    expect(validateProduct({ sku: 'ASTRAL-01', name: '😀'.repeat(120), description: '', priceCents: 1, status: 'active' })).toEqual({})
    expect(validateProduct({ sku: 'ASTRAL-01', name: '😀'.repeat(121), description: '界'.repeat(501), priceCents: 1, status: 'active' })).toMatchObject({ name: 'NAME_INVALID', description: 'DESCRIPTION_INVALID' })
    expect(validateProductSearch('😀'.repeat(100))).toBe(true)
    expect(validateProductSearch('😀'.repeat(101))).toBe(false)
  })
})
