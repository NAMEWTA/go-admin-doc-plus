import type { ApiEnvelope } from '../../contracts/src'
import type { RuntimeConfig } from '../../runtime/src'

export type ApiErrorKind = 'unauthorized' | 'business' | 'network' | 'protocol' | 'http'
export type ApiMethod = 'get' | 'post' | 'put' | 'patch' | 'delete'

export class ApiError extends Error {
  readonly kind: ApiErrorKind
  readonly code?: number
  readonly status?: number
  override readonly cause?: unknown

  constructor(
    kind: ApiErrorKind,
    message: string,
    options: { readonly code?: number, readonly status?: number, readonly cause?: unknown } = {}
  ) {
    super(message)
    this.name = 'ApiError'
    this.kind = kind
    this.code = options.code
    this.status = options.status
    this.cause = options.cause
  }
}

export interface TokenProvider {
  getToken(): string | undefined
}

export interface TransportRequest<TBody = unknown> {
  readonly url: string
  readonly method: ApiMethod
  readonly query?: Readonly<Record<string, unknown>>
  readonly body?: TBody
  readonly headers: Readonly<Record<string, string>>
  readonly timeoutMs: number
  readonly signal?: AbortSignal
}

export interface TransportResponse<T = unknown> {
  readonly status: number
  readonly data: T
}

export interface Transport {
  send(request: TransportRequest): Promise<TransportResponse>
}

export interface ApiRequest<TBody = unknown> {
  readonly path: string
  readonly method?: ApiMethod
  readonly query?: Readonly<Record<string, unknown>>
  readonly body?: TBody
  readonly headers?: Readonly<Record<string, string>>
  readonly signal?: AbortSignal
}

export interface ApiClient {
  request<T = unknown, TBody = unknown>(request: ApiRequest<TBody>): Promise<ApiEnvelope<T>>
  requestJson<T = unknown, TBody = unknown>(request: ApiRequest<TBody>): Promise<T>
}

const objectBody = (value: unknown): Record<string, unknown> | null =>
  value !== null && typeof value === 'object' && !Array.isArray(value)
    ? value as Record<string, unknown>
    : null

export const decodeApiEnvelope = <T>(body: unknown, status = 200): ApiEnvelope<T> => {
  const envelope = objectBody(body)
  if (!envelope || typeof envelope.code !== 'number' || typeof envelope.msg !== 'string') {
    throw new ApiError('protocol', 'Server returned an invalid API envelope', { status })
  }
  if (envelope.code === 200) {
    if (!('data' in envelope)) {
      throw new ApiError('protocol', 'Server returned a success envelope without data', { status })
    }
    return envelope as unknown as ApiEnvelope<T>
  }
  const message = envelope.msg || (envelope.code === 401
    ? 'Unauthorized'
    : envelope.code === 6401
      ? '登录状态已过期'
      : 'error')
  const kind: ApiErrorKind = envelope.code === 401 || envelope.code === 6401 ? 'unauthorized' : 'business'
  throw new ApiError(kind, message, { code: envelope.code, status })
}

export const toApiError = (error: unknown): ApiError => {
  if (error instanceof ApiError) return error
  const message = error instanceof Error
    ? error.message
    : objectBody(error) && typeof objectBody(error)?.message === 'string'
      ? String(objectBody(error)?.message)
      : 'Network request failed'
  return new ApiError('network', message || 'Network request failed', { cause: error })
}

const requestPath = (runtime: RuntimeConfig, path: string): string => {
  if (!path.startsWith('/') || path.startsWith('//') || path.includes('?') || path.includes('#')) {
    throw new ApiError('protocol', 'API request path must be an absolute path without query or fragment')
  }
  return `${runtime.apiBaseUrl}${path}`
}

const requestHeaders = (
  runtime: RuntimeConfig,
  tokenProvider: TokenProvider,
  supplied: Readonly<Record<string, string>> | undefined
): Readonly<Record<string, string>> => {
  const headers: Record<string, string> = {}
  const reserved = new Set(['authorization', runtime.launchToken?.header.toLowerCase()].filter(Boolean))
  for (const [name, value] of Object.entries(supplied ?? {})) {
    if (!reserved.has(name.toLowerCase())) headers[name] = value
  }
  const token = tokenProvider.getToken()
  if (typeof token === 'string' && token.trim() !== '') headers.Authorization = `Bearer ${token}`
  if (runtime.launchToken) headers[runtime.launchToken.header] = runtime.launchToken.value
  return headers
}

export const createApiClient = (
  runtime: RuntimeConfig,
  tokenProvider: TokenProvider,
  transport: Transport
): ApiClient => {
  if (!runtime || !tokenProvider || typeof tokenProvider.getToken !== 'function' || !transport || typeof transport.send !== 'function') {
    throw new Error('runtime, token provider and transport are required')
  }

  const send = async <T, TBody>(request: ApiRequest<TBody>, responseMode: 'envelope' | 'json'): Promise<T | ApiEnvelope<T>> => {
    const transportRequest: TransportRequest<TBody> = {
      url: requestPath(runtime, request.path),
      method: request.method ?? 'get',
      query: request.query,
      body: request.body,
      headers: requestHeaders(runtime, tokenProvider, request.headers),
      timeoutMs: runtime.requestTimeoutMs,
      signal: request.signal
    }
    let response: TransportResponse
    try {
      response = await transport.send(transportRequest)
    } catch (error) {
      throw toApiError(error)
    }
    if (responseMode === 'envelope') return decodeApiEnvelope<T>(response.data, response.status)
    if (response.status < 200 || response.status >= 300) {
      throw new ApiError(response.status === 401 ? 'unauthorized' : 'http', `HTTP ${response.status}`, { status: response.status })
    }
    return response.data as T
  }

  return {
    request: <T, TBody>(request: ApiRequest<TBody>) => send<T, TBody>(request, 'envelope') as Promise<ApiEnvelope<T>>,
    requestJson: <T, TBody>(request: ApiRequest<TBody>) => send<T, TBody>(request, 'json') as Promise<T>
  }
}
