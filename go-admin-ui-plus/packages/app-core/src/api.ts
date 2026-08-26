import type { ApiClient, ApiRequest } from '../../api-client/src'
import type { ApiEnvelope } from '../../contracts/src'

let configuredClient: Promise<ApiClient> | undefined

export const configureDomainApiClient = (client: ApiClient | Promise<ApiClient>) => {
  if (!client) throw new Error('Domain ApiClient is required')
  configuredClient = Promise.resolve(client)
}

export const requestDomain = async <T = unknown, TBody = unknown>(
  request: ApiRequest<TBody>
): Promise<ApiEnvelope<T>> => {
  if (!configuredClient) throw new Error('Domain ApiClient has not been configured by the App Shell')
  return (await configuredClient).request<T, TBody>(request)
}

export interface DomainRequestConfig<TBody = unknown> {
  readonly url: string
  readonly method?: ApiRequest<TBody>['method']
  readonly params?: object
  readonly data?: TBody
  readonly headers?: Readonly<Record<string, string>>
  readonly signal?: AbortSignal
}

/** Compatibility shape for APIs moved out of the legacy Axios facade. */
export const requestDomainCompat = <T = unknown, TBody = unknown>(config: DomainRequestConfig<TBody>): Promise<T> => {
  const [path, search = ''] = config.url.split('?', 2)
  const query: Record<string, unknown> = { ...(config.params ?? {}) }
  for (const [key, value] of new URLSearchParams(search)) query[key] = value
  return requestDomain({
    path,
    method: config.method,
    query: Object.keys(query).length ? query : undefined,
    body: config.data,
    headers: config.headers,
    signal: config.signal
  }) as Promise<T>
}

export const clearDomainApiClientForTests = () => {
  configuredClient = undefined
}
