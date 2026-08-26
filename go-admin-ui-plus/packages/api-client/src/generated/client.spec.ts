import { createContractClient, type components } from './client'

describe('generated contract client', () => {
  it('roundtrips the declared probe request and response', async() => {
    let request: Request | undefined
    const client = createContractClient({
      baseUrl: 'https://api.example.test',
      fetch: async input => {
        request = input
        return Response.json({ value: 'roundtrip' })
      }
    })

    const result = await client.POST('/contract/probe', {
      params: { query: { page: 2, pageSize: 40 } },
      body: { value: 'roundtrip' }
    })

    expect(result.data).toEqual({ value: 'roundtrip' })
    expect(result.error).toBeUndefined()
    expect(request?.method).toBe('POST')
    expect(new URL(request?.url ?? '').search).toBe('?page=2&pageSize=40')
    await expect(request?.clone().json()).resolves.toEqual({ value: 'roundtrip' })
  })

  it('returns the declared stable conflict problem', async() => {
    const problem: components['schemas']['Problem'] = {
      type: 'urn:go-admin-plus:problem:conflict',
      title: 'Resource conflict',
      status: 409,
      category: 'conflict',
      code: 'RESOURCE_CONFLICT',
      traceId: '0123456789abcdef'
    }
    const client = createContractClient({
      baseUrl: 'https://api.example.test',
      fetch: async() => Response.json(problem, {
        status: 409,
        headers: { 'Content-Type': 'application/problem+json' }
      })
    })

    const result = await client.POST('/contract/probe', { body: { value: 'duplicate' } })

    expect(result.data).toBeUndefined()
    expect(result.error).toEqual(problem)
  })
})
