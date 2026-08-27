import { describe, expect, it } from 'vitest'
import { buildDepartmentTree, codePointLength, validOrganizationName, validOrganizationSearch } from './organization'

describe('buildDepartmentTree', () => {
  it('projects the ordered flat transport into a nested tree', () => {
    const tree = buildDepartmentTree([
      { id: 'root', key: 'root', name: 'Root', sortOrder: 0, protected: true },
      { id: 'child', key: 'child', name: 'Child', parentId: 'root', sortOrder: 0, protected: false },
    ])
    expect(tree).toHaveLength(1)
    expect(tree[0]?.children.map((item) => item.id)).toEqual(['child'])
  })

  it('rejects unknown parents and cycles', () => {
    expect(() => buildDepartmentTree([
      { id: 'x', key: 'x', name: 'X', parentId: 'missing', sortOrder: 0, protected: false },
    ])).toThrow(TypeError)
    expect(() => buildDepartmentTree([
      { id: 'x', key: 'x', name: 'X', parentId: 'y', sortOrder: 0, protected: false },
      { id: 'y', key: 'y', name: 'Y', parentId: 'x', sortOrder: 0, protected: false },
    ])).toThrow(TypeError)
  })
})

describe('organization Unicode validation', () => {
  it('counts code points at name and search boundaries', () => {
    expect(codePointLength('😀')).toBe(1)
    expect(validOrganizationName('😀'.repeat(100))).toBe(true)
    expect(validOrganizationName('😀'.repeat(101))).toBe(false)
    expect(validOrganizationSearch('界'.repeat(100))).toBe(true)
    expect(validOrganizationSearch('界'.repeat(101))).toBe(false)
  })
})
