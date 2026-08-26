import { describe, expect, it } from 'vitest'
import { createWebSessionClient } from './web-session-client'

const json = (value: unknown, init: ResponseInit = {}) => new Response(JSON.stringify(value), { status: 200, headers: { 'Content-Type': 'application/json' }, ...init })

describe('web session client', () => {
  it('uses browser-managed cookies and never accepts a session token', async () => {
    const calls: Array<[Parameters<typeof fetch>[0], RequestInit | undefined]> = []
    const fetcher: typeof fetch = async (input, init) => {
      calls.push([input, init])
      return json({ profile: { id: '1', username: 'admin', displayName: 'Admin', email: 'a@example.test' }, csrfToken: 'csrf-value' })
    }
    const client = createWebSessionClient(fetcher, 'https://app.example.test/api')
    await client.login({ username: 'admin', password: 'sensitive-password' })
    await client.updateProfile({ displayName: 'Admin', email: 'a@example.test' })
    const loginRequest = new Request(calls[0]![0], calls[0]![1])
    const updateRequest = new Request(calls[1]![0], calls[1]![1])
    expect(loginRequest.credentials).toBe('include')
    expect(loginRequest.headers.has('Authorization')).toBe(false)
    expect(updateRequest.headers.get('X-CSRF-Token')).toBe('csrf-value')
    expect(JSON.stringify(client)).not.toMatch(/csrf-value|sensitive-password/)
  })

  it('clears the CSRF capability after authentication failure', async () => {
    let call = 0
    const headers: Headers[] = []
    const fetcher: typeof fetch = async (input, init) => {
      headers.push(new Request(input, init).headers)
      call += 1
      if (call === 1) return json({ profile: { id: '1', username: 'a', displayName: 'A', email: 'a@b.test' }, csrfToken: 'csrf-value' })
      return json({ category: 'authentication' }, { status: 401 })
    }
    const client = createWebSessionClient(fetcher, 'https://app.example.test/api')
    await client.login({ username: 'a', password: 'long-enough-password' })
    await expect(client.updateProfile({ displayName: 'A', email: 'a@b.test' })).rejects.toThrow('Session request failed')
    await expect(client.logout()).rejects.toThrow()
    expect(headers[2]?.has('X-CSRF-Token')).toBe(false)
  })

  it('serializes requests so a rotated CSRF value cannot be overwritten out of order', async () => {
    let release: (() => void) | undefined
    const gate = new Promise<void>((resolve) => { release = resolve })
    let calls = 0
    const fetcher: typeof fetch = async () => {
      calls += 1
      if (calls === 1) await gate
      return json({ profile: { id: '1', username: 'a', displayName: 'A', email: 'a@b.test' }, csrfToken: `csrf-${calls}` })
    }
    const client = createWebSessionClient(fetcher, 'https://app.example.test/api')
    const first = client.current()
    const second = client.current()
    await Promise.resolve()
    await Promise.resolve()
    expect(calls).toBe(1)
    release?.()
    await Promise.all([first, second])
    expect(calls).toBe(2)
  })
})
