import { createApiClient, ApiError, type Transport, type TransportRequest } from './index'
import { desktopLaunchTokenHeader, type RuntimeConfig } from '../../runtime/src'

const web: RuntimeConfig = {
  kind: 'web',
  apiBaseUrl: '',
  requestTimeoutMs: 10_000
}

const response = (data: unknown, status = 200) => ({ status, data })

const transportWith = (result: unknown) => {
  const requests: TransportRequest[] = []
  const transport: Transport = {
    async send(request) {
      requests.push(request)
      if (result instanceof Error) throw result
      return result as ReturnType<typeof response>
    }
  }
  return { requests, transport }
}

describe('pure ApiClient', () => {
  it('returns a typed success envelope and keeps query separate from the URL', async() => {
    const fake = transportWith(response({ code: 200, data: { id: 7 }, msg: 'ok' }))
    const client = createApiClient(web, { getToken: () => 'jwt-secret' }, fake.transport)
    await expect(client.request<{ id: number }>({ path: '/api/v1/sys-user', query: { pageIndex: 1 } }))
      .resolves.toEqual({ code: 200, data: { id: 7 }, msg: 'ok' })
    expect(fake.requests[0]).toMatchObject({
      url: '/api/v1/sys-user',
      query: { pageIndex: 1 },
      headers: { Authorization: 'Bearer jwt-secret' }
    })
    expect(fake.requests[0].url).not.toContain('jwt-secret')
  })

  it('adds the Desktop launch token only as its controlled header', async() => {
    const fake = transportWith(response({ code: 200, data: null, msg: 'ok' }))
    const desktop: RuntimeConfig = {
      kind: 'desktop',
      apiBaseUrl: 'http://127.0.0.1:43125',
      requestTimeoutMs: 3_000,
      launchToken: { header: desktopLaunchTokenHeader, value: 'launch-secret' }
    }
    const client = createApiClient(desktop, { getToken: () => 'jwt-secret' }, fake.transport)
    await client.request({
      path: '/api/v1/menurole',
      headers: { Authorization: 'attacker', [desktopLaunchTokenHeader]: 'attacker', Accept: 'application/json' }
    })
    expect(fake.requests[0].url).toBe('http://127.0.0.1:43125/api/v1/menurole')
    expect(fake.requests[0].headers).toEqual({
      Accept: 'application/json',
      Authorization: 'Bearer jwt-secret',
      [desktopLaunchTokenHeader]: 'launch-secret'
    })
    expect(fake.requests[0].url).not.toMatch(/secret|token/i)
  })

  it.each([
    [401, 'Unauthorized'],
    [6401, '登录状态已过期']
  ])('normalizes code %i as an unauthorized ApiError', async(code, msg) => {
    const fake = transportWith(response({ code, data: null, msg }))
    const error = await createApiClient(web, { getToken: () => undefined }, fake.transport)
      .request({ path: '/api/v1/getinfo' })
      .catch(value => value)
    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({ kind: 'unauthorized', code, message: msg })
  })

  it('normalizes a business error without presenting UI', async() => {
    // Error envelopes are allowed to omit data by the canonical OpenAPI schema.
    const fake = transportWith(response({ code: 403, msg: '没有权限' }))
    await expect(createApiClient(web, { getToken: () => undefined }, fake.transport)
      .request({ path: '/api/v1/menurole' }))
      .rejects.toMatchObject({ kind: 'business', code: 403, message: '没有权限' })
  })

  it('rejects a success envelope without data', async() => {
    const fake = transportWith(response({ code: 200, msg: 'ok' }))
    await expect(createApiClient(web, { getToken: () => undefined }, fake.transport)
      .request({ path: '/api/v1/menurole' }))
      .rejects.toMatchObject({ kind: 'protocol' })
  })

  it('rejects an unknown envelope explicitly', async() => {
    const fake = transportWith(response({ unexpected: true }))
    await expect(createApiClient(web, { getToken: () => undefined }, fake.transport)
      .request({ path: '/api/v1/menurole' }))
      .rejects.toMatchObject({ kind: 'protocol' })
  })

  it('normalizes a transport failure without UI or logging dependencies', async() => {
    const fake = transportWith(new Error('Network Error'))
    await expect(createApiClient(web, { getToken: () => undefined }, fake.transport)
      .request({ path: '/api/v1/menurole' }))
      .rejects.toMatchObject({ kind: 'network', message: 'Network Error' })
  })

  it('supports non-envelope operational JSON', async() => {
    const fake = transportWith(response({ hostProfile: 'server', desktop: false }))
    await expect(createApiClient(web, { getToken: () => undefined }, fake.transport)
      .requestJson({ path: '/api/v1/runtime/capabilities' }))
      .resolves.toEqual({ hostProfile: 'server', desktop: false })
  })

  it.each(['https://example.test/api/v1/menu', '//example.test/api', '/api/v1/menu?token=x'])
  ('rejects an unsafe path %s before transport', async path => {
    const fake = transportWith(response({ code: 200, data: null, msg: 'ok' }))
    await expect(createApiClient(web, { getToken: () => undefined }, fake.transport).request({ path }))
      .rejects.toMatchObject({ kind: 'protocol' })
    expect(fake.requests).toHaveLength(0)
  })
})
