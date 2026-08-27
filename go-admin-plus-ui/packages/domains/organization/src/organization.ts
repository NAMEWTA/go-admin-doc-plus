import type { components } from './generated/schema'

export type Department = components['schemas']['Department']
export type DepartmentInput = components['schemas']['DepartmentInput']
export type Position = components['schemas']['Position']
export type PositionInput = components['schemas']['PositionInput']
export type PositionPage = components['schemas']['PositionPage']

export interface DepartmentTreeNode extends Department {
  readonly children: ReadonlyArray<DepartmentTreeNode>
}

export interface OrganizationClient {
  listDepartments(): Promise<ReadonlyArray<Department>>
  createDepartment(input: DepartmentInput): Promise<Department>
  updateDepartment(id: string, input: DepartmentInput): Promise<Department>
  deleteDepartment(id: string): Promise<void>
  listPositions(search: string, page: number, pageSize: number): Promise<PositionPage>
  createPosition(input: PositionInput): Promise<Position>
  updatePosition(id: string, input: PositionInput): Promise<Position>
  deletePosition(id: string): Promise<void>
}

export class OrganizationRequestError extends Error {
  readonly category: string

  constructor(category: string) {
    super('Organization request failed')
    this.name = 'OrganizationRequestError'
    this.category = category
  }
}

export const codePointLength = (value: string): number => Array.from(value).length
export const validOrganizationName = (value: string): boolean => codePointLength(value.trim()) >= 1 && codePointLength(value.trim()) <= 100
export const validOrganizationSearch = (value: string): boolean => codePointLength(value.trim()) <= 100

/** Builds a defensive tree projection without mutating generated transport rows. */
export const buildDepartmentTree = (departments: ReadonlyArray<Department>): ReadonlyArray<DepartmentTreeNode> => {
  const rows = new Map<string, Department>()
  for (const item of departments) {
    if (rows.has(item.id)) throw new TypeError('Department projection contains a duplicate id')
    rows.set(item.id, item)
  }
  const visit = (item: Department, path: ReadonlySet<string>): DepartmentTreeNode => {
    if (path.has(item.id)) throw new TypeError('Department projection contains a cycle')
    const nextPath = new Set(path).add(item.id)
    const children = departments
      .filter((candidate) => candidate.parentId === item.id)
      .map((candidate) => visit(candidate, nextPath))
    return { ...item, children }
  }
  for (const item of departments) {
    if (item.parentId !== undefined && item.parentId !== null && !rows.has(item.parentId)) {
      throw new TypeError('Department projection contains an unknown parent')
    }
    const ancestry = new Set<string>()
    let current: Department | undefined = item
    while (current !== undefined) {
      if (ancestry.has(current.id)) throw new TypeError('Department projection contains a cycle')
      ancestry.add(current.id)
      current = current.parentId === undefined || current.parentId === null ? undefined : rows.get(current.parentId)
    }
  }
  return departments
    .filter((item) => item.parentId === undefined || item.parentId === null)
    .map((item) => visit(item, new Set()))
}
