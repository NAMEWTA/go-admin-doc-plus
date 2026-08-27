import { describe, expect, it, vi } from 'vitest'

import { createBrowserSessionFetch } from './session-fetch'

const origin = 'https://app.example.test'
const csrf = (value: string) => value.repeat(43)
const response = (status: number, value: unknown, token?: string) => new Response(JSON.stringify(value), {
  status,
  headers: { 'content-type': 'application/json', ...(token ? { 'X-CSRF-Token': token } : {}) }
})

describe('browser session fetch', () => {
  it('shares the latest CSRF across API clients and overrides stale caller state', async () => {
    const requests: Request[] = []
    const fetcher = vi.fn<typeof fetch>(async input => {
      requests.push(new Request(input))
      return requests.length === 1
        ? response(200, { profile: {} }, csrf('a'))
        : response(201, {}, csrf('b'))
    })
    const shared = createBrowserSessionFetch(fetcher, origin)

    await shared(`${origin}/api/iam/session/login`, { method: 'POST' })
    await shared(new Request(`${origin}/api/settings/values`, {
      method: 'POST',
      headers: { 'X-CSRF-Token': csrf('x') }
    }))

    expect(requests[0]?.headers.has('X-CSRF-Token')).toBe(false)
    expect(requests[1]?.headers.get('X-CSRF-Token')).toBe(csrf('a'))
  })

  it('serializes API domains so a replacement is committed before the next request', async () => {
    let release!: () => void
    const gate = new Promise<void>(resolve => { release = resolve })
    const requests: Request[] = []
    const fetcher = vi.fn<typeof fetch>(async input => {
      requests.push(new Request(input))
      if (requests.length === 1) await gate
      return response(200, {}, requests.length === 1 ? csrf('a') : csrf('b'))
    })
    const shared = createBrowserSessionFetch(fetcher, origin)

    const first = shared(`${origin}/api/runtime/identity`)
    const second = shared(`${origin}/api/settings/values`, { method: 'POST' })
    await Promise.resolve()
    await Promise.resolve()
    expect(fetcher).toHaveBeenCalledTimes(1)

    release()
    await Promise.all([first, second])
    expect(requests[1]?.headers.get('X-CSRF-Token')).toBe(csrf('a'))
  })

  it('never carries session state to another origin and clears it after CSRF rejection', async () => {
    const requests: Request[] = []
    const fetcher = vi.fn<typeof fetch>(async input => {
      const request = new Request(input)
      requests.push(request)
      if (requests.length === 1) return response(200, {}, csrf('a'))
      if (requests.length === 2) return response(200, {})
      if (requests.length === 3) return response(403, { code: 'CSRF_REJECTED' })
      return response(200, {})
    })
    const shared = createBrowserSessionFetch(fetcher, origin)

    await shared(`${origin}/api/runtime/identity`)
    await shared('https://outside.example/api/probe', { method: 'POST' })
    await shared(`${origin}/api/settings/values`, {
      method: 'POST',
      headers: { 'X-CSRF-Token': csrf('x') }
    })
    await shared(`${origin}/api/settings/values`, {
      method: 'POST',
      headers: { 'X-CSRF-Token': csrf('x') }
    })

    expect(requests[1]?.headers.has('X-CSRF-Token')).toBe(false)
    expect(requests[2]?.headers.get('X-CSRF-Token')).toBe(csrf('a'))
    expect(requests[3]?.headers.has('X-CSRF-Token')).toBe(false)
  })
})
