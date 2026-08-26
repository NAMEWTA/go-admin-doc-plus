import { describe, expect, it, vi } from 'vitest'
import { createWebOrganizationClient } from './web-organization-client'

describe('createWebOrganizationClient', () => {
  it('uses Cookie credentials, rotates CSRF, and never adds Authorization or WebStorage', async () => {
    const requests: Request[] = []
    const fetcher = vi.fn(async (request: Request) => {
      requests.push(request)
      if (request.method === 'GET') {
        return new Response(JSON.stringify([]), { status: 200, headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': 'csrf-next' } })
      }
      return new Response(JSON.stringify({ id: 'department-1', key: 'ops', name: 'Operations', parentId: 'department-root-001', sortOrder: 0, protected: false }), { status: 201, headers: { 'Content-Type': 'application/json' } })
    })
    const client = createWebOrganizationClient(fetcher as typeof fetch, 'https://admin.test/api')
    await client.listDepartments()
    await client.createDepartment({ key: 'ops', name: 'Operations', parentId: 'department-root-001', sortOrder: 0 })
    expect(requests).toHaveLength(2)
    expect(requests.every((request) => request.credentials === 'include')).toBe(true)
    expect(requests.every((request) => !request.headers.has('Authorization'))).toBe(true)
    expect(requests[1]?.headers.get('X-CSRF-Token')).toBe('csrf-next')
  })

  it('maps stable problems without exposing response details', async () => {
    const fetcher = vi.fn(async () => new Response(JSON.stringify({ category: 'conflict', code: 'RESOURCE_CONFLICT', detail: 'private SQL' }), { status: 409, headers: { 'Content-Type': 'application/problem+json' } }))
    const client = createWebOrganizationClient(fetcher as typeof fetch, 'https://admin.test/api')
    await expect(client.deleteDepartment('department-1')).rejects.toMatchObject({ category: 'conflict', message: 'Organization request failed' })
  })
})
