import { createContractClient, GeneratorRequestError, type GenerationDraft, type GenerationPreview, type GenerationResult, type GeneratorClient, type GeneratorFailure, type TableMetadata, type TableReference } from '@go-admin/domain-generator'

interface Problem { category?: string; code?: string }
const csrfPattern = /^[A-Za-z0-9_-]{43}$/

export const createWebGeneratorClient = (fetcher: typeof fetch = fetch, baseUrl = '/api'): GeneratorClient => {
  let csrf = '', classified: GeneratorFailure | null = null
  let tail: Promise<void> = Promise.resolve()
  const serialized = <T>(operation: () => Promise<T>): Promise<T> => { const result = tail.then(operation, operation); tail = result.then(() => undefined, () => undefined); return result }
  const contract = createContractClient({ baseUrl, fetch: async input => {
    const headers = new Headers(input.headers)
    if (csrf && input.method !== 'GET') headers.set('X-CSRF-Token', csrf)
    const response = await fetcher(new Request(input, { credentials: 'include', headers }))
    const next = response.headers.get('X-CSRF-Token')
    if (next !== null && !csrfPattern.test(next)) { csrf = ''; classified = 'relogin'; throw new GeneratorRequestError('relogin') }
    const body = response.status >= 400 ? await response.clone().json().catch(() => null) as Problem | null : null
    classified = classify(response.status, body)
    if (next) csrf = next
    else if (classified === 'relogin') csrf = ''
    return response
  } })
  const fail = (error: unknown): never => { const category = classified ?? problemCategory(error); classified = null; throw new GeneratorRequestError(category) }
  const required = <T>(data: T | undefined, error: unknown): T => error === undefined && data !== undefined ? data : fail(error)
  return {
    getConfig: module => serialized(async () => { const result = await contract.GET('/generator/configs/{moduleName}', { params: { path: { moduleName: module } } }); return required(result.data, result.error) as { draft: GenerationDraft; previewDigest: string } }),
    listTables: () => serialized(async () => { const result = await contract.GET('/generator/tables'); return required(result.data, result.error).map(value => ({ ...value })) as TableReference[] }),
    describe: table => serialized(async () => { const result = await contract.GET('/generator/tables/{schemaName}/{tableName}', { params: { path: { schemaName: table.schema, tableName: table.name } } }); return required(result.data, result.error) as TableMetadata }),
    preview: draft => serialized(async () => { const result = await contract.POST('/generator/previews', { body: transportDraft(draft) }); return required(result.data, result.error) as GenerationPreview }),
    write: previewToken => serialized(async () => { const result = await contract.POST('/generator/writes', { body: { previewToken, confirmed: true } }); return required(result.data, result.error) as GenerationResult }),
  }
}

const transportDraft = (draft: GenerationDraft) => ({ ...draft, columns: draft.columns.map(column => ({ ...column })) })
const classify = (status: number, value: Problem | null): GeneratorFailure | null => {
  if (status === 401 || value?.code === 'CSRF_REJECTED') return 'relogin'
  if (status === 403) return 'forbidden'
  if (value?.code === 'OUTPUT_GATE_FAILED' || status === 422) return 'gate'
  if (status === 400) return 'validation'
  if (status === 404) return 'not-found'
  if (status === 409) return 'conflict'
  if (status >= 500) return 'unavailable'
  return null
}
const problemCategory = (value: unknown): GeneratorFailure => typeof value === 'object' && value !== null && 'category' in value
  ? ({ authentication: 'relogin', authorization: 'forbidden', validation: 'validation', not_found: 'not-found', conflict: 'conflict' } as const)[String((value as Problem).category) as 'authentication'|'authorization'|'validation'|'not_found'|'conflict'] ?? 'unavailable'
  : 'unavailable'
