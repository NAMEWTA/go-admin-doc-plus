import { createContractClient, SchedulerRequestError, type SchedulerClient } from '@go-admin-plus/domain-scheduler'

interface Problem { category?: string; code?: string }

export const createWebSchedulerClient = (fetcher: typeof fetch = fetch, baseUrl = '/api'): SchedulerClient => {
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
  const takeFailure = (fallback: string) => { const value = responseFailure ?? fallback; responseFailure = null; return value }
  const unwrap = <T>(data: T | undefined, error: unknown): T => {
    if (error !== undefined) throw new SchedulerRequestError(takeFailure(problemCategory(error)))
    if (data === undefined) throw new SchedulerRequestError('unavailable')
    return data
  }
  const complete = (error: unknown) => { if (error !== undefined) throw new SchedulerRequestError(takeFailure(problemCategory(error))) }
  return {
    taskTypes: () => serialized(async () => { const result = await contract.GET('/scheduler/task-types'); return unwrap(result.data, result.error) }),
    listDefinitions: (search, page, pageSize) => serialized(async () => { const result = await contract.GET('/scheduler/definitions', { params: { query: { search, page, pageSize } } }); return unwrap(result.data, result.error) }),
    createDefinition: (body) => serialized(async () => { const result = await contract.POST('/scheduler/definitions', { body }); return unwrap(result.data, result.error) }),
    updateDefinition: (definitionId, revision, body) => serialized(async () => { const result = await contract.PATCH('/scheduler/definitions/{definitionId}', { params: { path: { definitionId } }, body: { ...body, revision } }); return unwrap(result.data, result.error) }),
    enableDefinition: (definitionId, revision) => serialized(async () => { const result = await contract.POST('/scheduler/definitions/{definitionId}/enable', { params: { path: { definitionId } }, body: { revision } }); return unwrap(result.data, result.error) }),
    stopDefinition: (definitionId, revision) => serialized(async () => { const result = await contract.POST('/scheduler/definitions/{definitionId}/stop', { params: { path: { definitionId } }, body: { revision } }); return unwrap(result.data, result.error) }),
    deleteDefinition: (definitionId, revision) => serialized(async () => { const result = await contract.DELETE('/scheduler/definitions/{definitionId}', { params: { path: { definitionId }, query: { revision } } }); complete(result.error) }),
    listExecutions: (definitionId, status, page, pageSize) => serialized(async () => { const result = await contract.GET('/scheduler/executions', { params: { query: { ...(definitionId ? { definitionId } : {}), ...(status ? { status } : {}), page, pageSize } } }); return unwrap(result.data, result.error) }),
  }
}

const problemCategory = (value: unknown) => typeof value === 'object' && value !== null && 'category' in value ? String((value as Problem).category ?? 'unavailable') : 'unavailable'
const classifyResponse = (status: number, problem: Problem | null): string | null => {
  if (status === 401 || problem?.code === 'CSRF_REJECTED') return 'relogin'
  if (status === 403) return 'forbidden'
  if (status === 400 || status === 422) return 'validation'
  if (status === 404) return 'not-found'
  if (status === 409) return 'conflict'
  if (status >= 500) return 'unavailable'
  return null
}
