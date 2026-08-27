type Fetch = typeof globalThis.fetch

interface Problem { code?: string }

const csrfPattern = /^[A-Za-z0-9_-]{43}$/
const safeMethods = new Set(['GET', 'HEAD', 'OPTIONS'])

/**
 * Owns the browser session's rolling CSRF value across every API domain.
 * The server replaces that value after each protected request, so API traffic
 * must share both one value and one request queue.
 */
export const createBrowserSessionFetch = (
  fetcher: Fetch = globalThis.fetch,
  apiOrigin = globalThis.location.origin
): Fetch => {
  let csrf = ''
  let requestTail: Promise<void> = Promise.resolve()

  const serialized = <T>(operation: () => Promise<T>): Promise<T> => {
    const result = requestTail.then(operation, operation)
    requestTail = result.then(() => undefined, () => undefined)
    return result
  }

  return async (input, init) => {
    const request = new Request(input, init)
    const url = new URL(request.url)
    if (url.origin !== apiOrigin || !url.pathname.startsWith('/api/')) return fetcher(request)

    return serialized(async () => {
      const headers = new Headers(request.headers)
      if (safeMethods.has(request.method.toUpperCase()) || !csrf) headers.delete('X-CSRF-Token')
      else headers.set('X-CSRF-Token', csrf)

      const response = await fetcher(new Request(request, { credentials: 'include', headers }))
      const replacement = response.headers.get('X-CSRF-Token')
      if (replacement !== null) csrf = csrfPattern.test(replacement) ? replacement : ''
      if (response.status === 401) csrf = ''
      if (response.status === 403) {
        const problem = await response.clone().json().catch(() => null) as Problem | null
        if (problem?.code === 'CSRF_REJECTED') csrf = ''
      }
      return response
    })
  }
}
