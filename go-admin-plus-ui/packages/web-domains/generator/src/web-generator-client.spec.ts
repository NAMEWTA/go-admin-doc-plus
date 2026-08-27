import { describe, expect, it, vi } from 'vitest'
import { createWebGeneratorClient } from './web-generator-client'

const json = (status: number, body: unknown, headers: Record<string, string> = {}) => new Response(JSON.stringify(body), { status, headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json', ...headers } })

describe('web generator client', () => {
  it('carries CSRF only on mutations and classifies gate failures', async () => {
    const fetcher = vi.fn<typeof fetch>().mockResolvedValueOnce(json(200, [], { 'X-CSRF-Token': 'c'.repeat(43) })).mockResolvedValueOnce(json(422, { category: 'validation', code: 'OUTPUT_GATE_FAILED' }))
    const client = createWebGeneratorClient(fetcher, 'https://generator.test/api')
    await client.listTables()
    await expect(client.write('a'.repeat(64))).rejects.toMatchObject({ category: 'gate' })
    expect(new Request(fetcher.mock.calls[0]![0]).headers.has('X-CSRF-Token')).toBe(false)
    expect(new Request(fetcher.mock.calls[1]![0]).headers.get('X-CSRF-Token')).toBe('c'.repeat(43))
  })
  it('fails closed on malformed CSRF replacement', async () => {
    const client = createWebGeneratorClient(vi.fn<typeof fetch>().mockResolvedValue(json(200, [], { 'X-CSRF-Token': 'invalid' })), 'https://generator.test/api')
    await expect(client.listTables()).rejects.toMatchObject({ category: 'relogin' })
  })
})
