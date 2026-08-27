import { createContractClient, DemoRequestError, type DemoClient, type DemoFailure } from '@go-admin/domain-demo'

interface Problem { category?: string; code?: string }
const csrfPattern = /^[A-Za-z0-9_-]{43}$/

export const createWebDemoClient = (fetcher: typeof fetch = fetch, baseUrl = '/api'): DemoClient => {
  let csrf = ''
  let classified: DemoFailure | null = null
  let tail: Promise<void> = Promise.resolve()
  const serialized = <T>(operation: () => Promise<T>): Promise<T> => {
    const result = tail.then(operation, operation)
    tail = result.then(() => undefined, () => undefined)
    return result
  }
  const contract = createContractClient({ baseUrl, fetch: async (input) => {
    const headers = new Headers(input.headers)
    if (csrf && input.method !== 'GET') headers.set('X-CSRF-Token', csrf)
    const response = await fetcher(new Request(input, { credentials: 'include', headers }))
    const next = response.headers.get('X-CSRF-Token')
    if (next !== null && !csrfPattern.test(next)) { csrf = ''; classified = 'relogin'; throw new DemoRequestError('relogin') }
    const body = response.status >= 400 ? await response.clone().json().catch(() => null) as Problem | null : null
    classified = classify(response.status, body)
    if (next) csrf = next
    else if (classified === 'relogin') csrf = ''
    return response
  } })
  const failure = (error: unknown): never => {
    const category = classified ?? problemCategory(error)
    classified = null
    throw new DemoRequestError(category)
  }
  const required = <T>(data: T | undefined, error: unknown): T => error === undefined && data !== undefined ? data : failure(error)
  const completed = (error: unknown): void => { if (error !== undefined) failure(error) }
  return {
    list: query => serialized(async () => { const result = await contract.GET('/demo/products', { params: { query } }); return required(result.data, result.error) }),
    get: id => serialized(async () => { const result = await contract.GET('/demo/products/{productId}', { params: { path: { productId: id } } }); return required(result.data, result.error) }),
    create: body => serialized(async () => { const result = await contract.POST('/demo/products', { body }); return required(result.data, result.error) }),
    update: (id, body) => serialized(async () => { const result = await contract.PATCH('/demo/products/{productId}', { params: { path: { productId: id } }, body }); return required(result.data, result.error) }),
    delete: products => serialized(async () => { const result = await contract.POST('/demo/products/batch-delete', { body: { products: [...products] } }); completed(result.error) }),
  }
}

const classify = (status: number, value: Problem | null): DemoFailure | null => {
  if (status === 401 || value?.code === 'CSRF_REJECTED') return 'relogin'
  if (status === 403) return 'forbidden'
  if (status === 400 || status === 422) return 'validation'
  if (status === 404) return 'not-found'
  if (status === 409) return 'conflict'
  if (status >= 500) return 'unavailable'
  return null
}
const problemCategory = (value: unknown): DemoFailure => typeof value === 'object' && value !== null && 'category' in value
  && ['authentication', 'authorization', 'validation', 'not-found', 'conflict'].includes(String((value as Problem).category))
  ? ({ authentication: 'relogin', authorization: 'forbidden', validation: 'validation', 'not-found': 'not-found', conflict: 'conflict' } as const)[String((value as Problem).category) as 'authentication'|'authorization'|'validation'|'not-found'|'conflict']
  : 'unavailable'
