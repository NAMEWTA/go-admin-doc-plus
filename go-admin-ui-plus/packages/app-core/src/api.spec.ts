import { clearDomainApiClientForTests, configureDomainApiClient, requestDomain, requestDomainCompat } from './api'
import type { ApiClient } from '../../api-client/src'

describe('Domain API port', () => {
  afterEach(clearDomainApiClientForTests)

  it('fails explicitly before the App Shell configures a client', async() => {
    await expect(requestDomain({ path: '/api/v1/demo-product' })).rejects.toThrow('has not been configured')
  })

  it('delegates the typed request without changing its path or query', async() => {
    const request = vi.fn().mockResolvedValue({ code: 200, data: { id: 7 }, msg: 'ok' })
    configureDomainApiClient({ request, requestJson: vi.fn() } as unknown as ApiClient)
    await expect(requestDomain<{ id: number }>({ path: '/api/v1/demo-product', query: { pageIndex: 1 } }))
      .resolves.toEqual({ code: 200, data: { id: 7 }, msg: 'ok' })
    expect(request).toHaveBeenCalledWith({ path: '/api/v1/demo-product', query: { pageIndex: 1 } })
  })

  it('splits legacy inline query parameters before reaching the pure client', async() => {
    const request = vi.fn().mockResolvedValue({ code: 200, data: [], msg: 'ok' })
    configureDomainApiClient({ request, requestJson: vi.fn() } as unknown as ApiClient)

    await requestDomainCompat({
      url: '/api/v1/dict-data/option-select?dictType=sys_common_status',
      method: 'get',
      params: { locale: 'zh-CN' }
    })

    expect(request).toHaveBeenCalledWith({
      path: '/api/v1/dict-data/option-select',
      method: 'get',
      query: { locale: 'zh-CN', dictType: 'sys_common_status' },
      body: undefined,
      headers: undefined,
      signal: undefined
    })
  })
})
