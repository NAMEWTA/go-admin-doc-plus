import { createHostRuntime } from './index'

describe('Host Runtime selection', () => {
  it('keeps normal browsers on the same-origin Web Runtime', async() => {
    await expect(createHostRuntime({}).bootstrap()).resolves.toEqual({
      kind: 'web',
      apiBaseUrl: '',
      requestTimeoutMs: 10_000
    })
  })

  it('uses the only Wails bootstrap binding for Desktop', async() => {
    const bootstrap = vi.fn(async() => ({
      apiBaseUrl: 'http://127.0.0.1:43125',
      requestTimeoutMs: 5_000,
      launchToken: {
        header: 'X-Go-Admin-Launch-Token',
        value: 'process-secret'
      }
    }))
    const runtime = createHostRuntime({
      go: { desktop: { Bridge: { Bootstrap: bootstrap } } }
    })

    await expect(runtime.bootstrap()).resolves.toMatchObject({
      kind: 'desktop',
      apiBaseUrl: 'http://127.0.0.1:43125',
      launchToken: { value: 'process-secret' }
    })
    expect(bootstrap).toHaveBeenCalledOnce()
  })

  it('fails closed when a partial Wails binding is present', () => {
    expect(() => createHostRuntime({ go: { desktop: { Bridge: {} } } })).toThrow(
      'desktop bootstrap binding is unavailable'
    )
  })
})
