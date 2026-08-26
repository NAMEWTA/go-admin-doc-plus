import {
  createDesktopRuntime,
  createWebRuntime,
  desktopLaunchTokenHeader,
  validateDesktopBootstrap
} from './index'

describe('Runtime bootstrap', () => {
  it('uses the same origin for Web without reading build-time URLs', async() => {
    await expect(createWebRuntime().bootstrap()).resolves.toEqual({
      kind: 'web',
      apiBaseUrl: '',
      requestTimeoutMs: 10_000
    })
  })

  it('accepts only an explicit loopback Desktop origin and controlled header', async() => {
    const runtime = createDesktopRuntime({
      bootstrap: async() => ({
        apiBaseUrl: 'http://127.0.0.1:43125/',
        requestTimeoutMs: 5_000,
        launchToken: { header: desktopLaunchTokenHeader, value: 'launch-secret' }
      })
    })
    await expect(runtime.bootstrap()).resolves.toEqual({
      kind: 'desktop',
      apiBaseUrl: 'http://127.0.0.1:43125',
      requestTimeoutMs: 5_000,
      launchToken: { header: desktopLaunchTokenHeader, value: 'launch-secret' }
    })
  })

  it.each([
    null,
    {},
    { apiBaseUrl: 'https://127.0.0.1:43125', launchToken: { header: desktopLaunchTokenHeader, value: 'x' } },
    { apiBaseUrl: 'http://example.test:43125', launchToken: { header: desktopLaunchTokenHeader, value: 'x' } },
    { apiBaseUrl: 'http://127.0.0.1:43125?token=x', launchToken: { header: desktopLaunchTokenHeader, value: 'x' } },
    { apiBaseUrl: 'http://127.0.0.1:43125', launchToken: { header: 'Authorization', value: 'x' } },
    { apiBaseUrl: 'http://127.0.0.1:43125', launchToken: { header: desktopLaunchTokenHeader, value: '' } }
  ])('fails closed for an invalid Desktop payload %#j', payload => {
    expect(() => validateDesktopBootstrap(payload)).toThrow()
  })
})
