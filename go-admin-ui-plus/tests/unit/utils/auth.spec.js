const cookies = vi.hoisted(() => ({
  get: vi.fn(),
  set: vi.fn(),
  remove: vi.fn()
}))

vi.mock('js-cookie', () => ({ default: cookies }))

import { getToken, removeToken, setToken } from '@/utils/auth'

describe('Host-aware auth token storage', () => {
  beforeEach(() => {
    delete globalThis.go
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('preserves cookies for the Web Host', () => {
    cookies.get.mockReturnValue('web-token')

    expect(getToken()).toBe('web-token')
    setToken('next-web-token')
    removeToken()

    expect(cookies.set).toHaveBeenCalledWith('Admin-Token', 'next-web-token')
    expect(cookies.remove).toHaveBeenCalledWith('Admin-Token')
  })

  it('uses local storage for the Wails Desktop Host', () => {
    globalThis.go = { desktop: { Bridge: { Bootstrap: vi.fn() }}}

    expect(setToken('desktop-token')).toBe('desktop-token')
    expect(getToken()).toBe('desktop-token')
    removeToken()

    expect(getToken()).toBeUndefined()
    expect(cookies.set).not.toHaveBeenCalled()
    expect(cookies.remove).not.toHaveBeenCalled()
  })
})
