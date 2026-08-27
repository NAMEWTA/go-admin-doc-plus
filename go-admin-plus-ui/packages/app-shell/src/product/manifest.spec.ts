import { describe, expect, it } from 'vitest'

import { assertProductManifest, productModules, productRoutesFor, type ProductModule } from './manifest'

describe('product manifest', () => {
  it('publishes the complete stable module order', () => {
    expect(productModules.map(module => module.id)).toEqual([
      'iam',
      'audit',
      'organization',
      'settings',
      'generator',
      'scheduler',
      'demo',
      'files'
    ])
  })

  it('gives Web and Desktop the same business navigation', () => {
    expect(productRoutesFor('desktop')).toEqual(productRoutesFor('web'))
    expect(productRoutesFor('web')).toHaveLength(13)
  })

  it('rejects duplicate routes and incomplete host coverage', () => {
    const invalid = [
      productModules[0],
      {
        ...productModules[1],
        hosts: ['web'],
        routes: [{ ...productModules[0].routes[0] }]
      }
    ] satisfies readonly ProductModule[]

    expect(() => assertProductManifest(invalid)).toThrow('product host coverage is incomplete')
  })
})
