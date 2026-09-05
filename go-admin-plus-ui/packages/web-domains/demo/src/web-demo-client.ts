import { createContractClient, DemoRequestError, type DemoClient, type DemoFailure } from '@go-admin-plus/domain-demo'
import { createSessionAwareFetch } from '@go-admin-plus/ui'

interface Problem { category?: string; code?: string; traceId?: string }
const tracePattern = /^[A-Za-z0-9_-]{8,128}$/

export const createWebDemoClient = (fetcher: typeof fetch = fetch, baseUrl = '/api'): DemoClient => {
  let tail: Promise<void> = Promise.resolve()
  const serialized = <T>(operation: () => Promise<T>): Promise<T> => {
    const result = tail.then(operation, operation)
    tail = result.then(() => undefined, () => undefined)
    return result
  }
  const contract = createContractClient({ baseUrl, fetch: createSessionAwareFetch(fetcher) })
  const failure = (error: unknown): never => {
    const category = problemCategory(error)
    const traceId = error instanceof DemoRequestError ? error.traceId : safeTraceId((error as Problem)?.traceId)
    throw new DemoRequestError(category, traceId)
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

const problemCategory = (value: unknown): DemoFailure => {
  if (typeof value !== 'object' || value === null) return 'unavailable'
  const problem = value as Problem
  if (problem.code === 'CSRF_REJECTED' || problem.category === 'authentication') return 'relogin'
  if (problem.category === 'authorization') return 'forbidden'
  if (problem.category === 'validation') return 'validation'
  if (problem.category === 'not-found') return 'not-found'
  if (problem.category === 'conflict') return 'conflict'
  return 'unavailable'
}
const safeTraceId = (value: unknown): string | undefined => typeof value === 'string' && tracePattern.test(value) ? value : undefined
