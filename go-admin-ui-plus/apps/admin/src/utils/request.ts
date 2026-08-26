import axios from 'axios'
import type { AxiosError, AxiosRequestConfig, AxiosResponse } from 'axios'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { getToken } from '@/utils/auth'
import type { ApiResponse } from '@/types/api'
import {
  ApiError,
  createApiClient,
  decodeApiEnvelope,
  toApiError,
  type ApiMethod,
  type Transport
} from '../../../../packages/api-client/src'
import { createHostRuntime } from '../../../../packages/runtime/src/desktop'

/** An error the compatibility App Shell has already put in front of the user. */
export interface ReportedError extends Error {
  reported?: true
}

const reported = <E extends Error>(error: E): E & { reported: true } =>
  Object.assign(error, { reported: true as const })

export const asReportedError = (error: unknown): ReportedError =>
  error instanceof Error ? error : new Error(String(error))

// Runtime owns the endpoint. Web is always same-origin; Desktop supplies a
// loopback endpoint through its bootstrap provider in the desktop shell.
const service = axios.create({ timeout: 10_000 })

service.interceptors.request.use(
  config => {
    if (useUserStore().token) {
      config.headers.Authorization = 'Bearer ' + getToken()
      if (!(config.data instanceof FormData)) {
        config.headers['Content-Type'] = 'application/json'
      }
    }
    return config
  },
  error => Promise.reject(error)
)

const promptRelogin = () => {
  useUserStore().resetToken()
  ElMessageBox.confirm(
    '登录状态已过期，您可以继续留在该页面，或者重新登录',
    '系统提示',
    {
      confirmButtonText: '重新登录',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    location.reload()
  })
}

const presentError = (error: ApiError) => {
  if (error.kind === 'unauthorized') {
    if (error.code === 401 && location.href.includes('login')) {
      useUserStore().resetToken()
      location.reload()
    } else {
      promptRelogin()
    }
    return
  }
  const message = error.kind === 'network' && error.message === 'Network Error'
    ? '服务器连接异常，请检查服务器！'
    : error.message
  ElMessage({
    message,
    type: 'error',
    duration: error.code === 400 || error.code === 403 ? 5_000 : error.kind === 'network' ? 5_000 : 3_000
  })
}

// These interceptors retain the legacy facade's observable contract while all
// envelope/error decisions come from the UI-free api-client package.
service.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    try {
      return decodeApiEnvelope(response.data, response.status) as unknown as AxiosResponse
    } catch(value) {
      const error = toApiError(value)
      presentError(error)
      return Promise.reject(reported(error))
    }
  },
  (value: AxiosError<ApiResponse>) => {
    let error: ApiError
    if (value.response) {
      try {
        decodeApiEnvelope(value.response.data, value.response.status)
        error = new ApiError('http', `HTTP ${value.response.status}`, { status: value.response.status })
      } catch(decoded) {
        error = toApiError(decoded)
      }
    } else {
      error = toApiError(value)
    }
    presentError(error)
    return Promise.reject(reported(error))
  }
)

const transport: Transport = {
  async send(request) {
    const data = await service.request({
      url: request.url,
      method: request.method,
      params: request.query,
      data: request.body,
      headers: request.headers,
      timeout: request.timeoutMs,
      signal: request.signal
    })
    return { status: 200, data }
  }
}

const appendQuery = (url: string, query: Readonly<Record<string, unknown>> | undefined): string => {
  if (!query) return url
  const parsed = new URL(url)
  const append = (name: string, value: unknown) => {
    if (value === undefined || value === null) return
    parsed.searchParams.append(name, typeof value === 'object' ? JSON.stringify(value) : String(value))
  }
  for (const [name, value] of Object.entries(query)) {
    if (Array.isArray(value)) value.forEach(item => append(name, item))
    else append(name, value)
  }
  return parsed.toString()
}

export const createDesktopTransport = (fetcher: typeof fetch = fetch): Transport => ({
  async send(request) {
    const controller = new AbortController()
    const abort = () => controller.abort(request.signal?.reason)
    request.signal?.addEventListener('abort', abort, { once: true })
    const timer = setTimeout(() => controller.abort(new Error('Request timed out')), request.timeoutMs)
    try {
      const headers = new Headers(request.headers)
      let body: FormData | Blob | string | undefined
      if (request.body !== undefined) {
        if (request.body instanceof FormData || request.body instanceof Blob) body = request.body
        else {
          if (!headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
          body = JSON.stringify(request.body)
        }
      }
      const response = await fetcher(appendQuery(request.url, request.query), {
        method: request.method.toUpperCase(),
        headers,
        body,
        signal: controller.signal
      })
      const text = await response.text()
      return { status: response.status, data: text === '' ? null : JSON.parse(text) }
    } finally {
      clearTimeout(timer)
      request.signal?.removeEventListener('abort', abort)
    }
  }
})

export const domainApiClient = createHostRuntime().bootstrap().then(runtime => createApiClient(
  runtime,
  {
    getToken: () => useUserStore().token ? getToken() : undefined
  },
  runtime.kind === 'desktop' ? createDesktopTransport() : transport
))

const apiMethods = new Set<ApiMethod>(['get', 'post', 'put', 'patch', 'delete'])

const normalizeMethod = (method: AxiosRequestConfig['method']): ApiMethod => {
  const normalized = (method ?? 'get').toLowerCase()
  if (!apiMethods.has(normalized as ApiMethod)) {
    throw new ApiError('protocol', `Unsupported API method ${normalized}`)
  }
  return normalized as ApiMethod
}

const normalizeBody = (value: unknown): unknown => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return value
  const prototype = Object.getPrototypeOf(value)
  return prototype === Object.prototype || prototype === null
    ? { ...value as Record<string, unknown> }
    : value
}

export const normalizeRequest = (config: AxiosRequestConfig) => {
  if (typeof config.url !== 'string') throw new ApiError('protocol', 'API request URL is required')
  const [path, search = ''] = config.url.split('?', 2)
  const query: Record<string, unknown> = {
    ...(config.params && typeof config.params === 'object' ? config.params : {})
  }
  for (const [key, value] of new URLSearchParams(search)) query[key] = value
  const headers: Record<string, string> = {}
  for (const [name, value] of Object.entries(config.headers ?? {})) {
    if (typeof value === 'string') headers[name] = value
  }
  return {
    path,
    method: normalizeMethod(config.method),
    query: Object.keys(query).length ? query : undefined,
    body: normalizeBody(config.data),
    headers: Object.keys(headers).length ? headers : undefined,
    signal: config.signal as AbortSignal | undefined
  }
}

/**
 * Compatibility facade for the existing API modules. The generic still
 * describes the full `{ code, data, msg }` envelope they currently consume.
 * T-07 moves Domain consumers onto the package-level client directly.
 */
const request = async <T = unknown>(config: AxiosRequestConfig): Promise<T> => {
  const normalized = normalizeRequest(config)
  const client = await domainApiClient
  const envelope = await client.request(normalized)
  return envelope as T
}

export default request
