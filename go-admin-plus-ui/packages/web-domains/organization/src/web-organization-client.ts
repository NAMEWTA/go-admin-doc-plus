import { createContractClient, OrganizationRequestError, type OrganizationClient } from '@go-admin-plus/domain-organization'

interface Problem { category?: string; code?: string }

export const createWebOrganizationClient = (fetcher: typeof fetch = fetch, baseUrl = '/api'): OrganizationClient => {
  let csrf = ''
  let responseFailure: string | null = null
  let requestTail: Promise<void> = Promise.resolve()
  const serialized = <T>(operation: () => Promise<T>): Promise<T> => {
    const result = requestTail.then(operation, operation)
    requestTail = result.then(() => undefined, () => undefined)
    return result
  }
  const contract = createContractClient({
    baseUrl,
    fetch: async (input) => {
      const headers = new Headers(input.headers)
      if (csrf && input.method !== 'GET') headers.set('X-CSRF-Token', csrf)
      const response = await fetcher(new Request(input, { credentials: 'include', headers }))
      const next = response.headers.get('X-CSRF-Token')
      const body = response.status >= 400 ? await response.clone().json().catch(() => null) as Problem | null : null
      responseFailure = classifyResponse(response.status, body)
      if (next) csrf = next
      else if (responseFailure === 'relogin') csrf = ''
      return response
    },
  })
  const takeFailure = (fallback: string) => {
    const value = responseFailure ?? fallback
    responseFailure = null
    return value
  }
  const unwrap = <T>(data: T | undefined, error: unknown): T => {
    if (error !== undefined) throw new OrganizationRequestError(takeFailure(problemCategory(error)))
    if (data === undefined) throw new OrganizationRequestError('unavailable')
    return data
  }
  const complete = (error: unknown) => {
    if (error !== undefined) throw new OrganizationRequestError(takeFailure(problemCategory(error)))
  }
  return {
    listDepartments: () => serialized(async () => { const result = await contract.GET('/organization/departments'); return unwrap(result.data, result.error) }),
    createDepartment: (body) => serialized(async () => { const result = await contract.POST('/organization/departments', { body }); return unwrap(result.data, result.error) }),
    updateDepartment: (departmentId, body) => serialized(async () => { const result = await contract.PATCH('/organization/departments/{departmentId}', { params: { path: { departmentId } }, body }); return unwrap(result.data, result.error) }),
    deleteDepartment: (departmentId) => serialized(async () => { const result = await contract.DELETE('/organization/departments/{departmentId}', { params: { path: { departmentId } } }); complete(result.error) }),
    listPositions: (search, page, pageSize) => serialized(async () => { const result = await contract.GET('/organization/positions', { params: { query: { search, page, pageSize } } }); return unwrap(result.data, result.error) }),
    createPosition: (body) => serialized(async () => { const result = await contract.POST('/organization/positions', { body }); return unwrap(result.data, result.error) }),
    updatePosition: (positionId, body) => serialized(async () => { const result = await contract.PATCH('/organization/positions/{positionId}', { params: { path: { positionId } }, body }); return unwrap(result.data, result.error) }),
    deletePosition: (positionId) => serialized(async () => { const result = await contract.DELETE('/organization/positions/{positionId}', { params: { path: { positionId } } }); complete(result.error) }),
  }
}

const problemCategory = (value: unknown) => typeof value === 'object' && value !== null && 'category' in value
  ? String((value as Problem).category ?? 'unavailable')
  : 'unavailable'

const classifyResponse = (status: number, problem: Problem | null): string | null => {
  if (status === 401 || problem?.code === 'CSRF_REJECTED') return 'relogin'
  if (status === 403) return 'forbidden'
  if (status === 400 || status === 422) return 'validation'
  if (status === 404) return 'not-found'
  if (status === 409) return 'conflict'
  if (status >= 500) return 'unavailable'
  return null
}
