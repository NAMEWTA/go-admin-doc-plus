import { describe, expect, it } from 'vitest'
import { createSessionController, SessionRequestError, type AccountProfile, type SessionClient } from './session-controller'

const profile: AccountProfile = { id: '1', username: 'admin', displayName: 'Admin', email: 'admin@example.test' }
const client = (overrides: Partial<SessionClient> = {}): SessionClient => ({
  login: async () => profile, current: async () => profile, logout: async () => undefined,
  profile: async () => profile, updateProfile: async () => profile, changePassword: async () => undefined,
  ...overrides,
})

describe('session controller', () => {
  it('does not expose cookie, token, csrf, or password state', async () => {
    const controller = createSessionController(client())
    await controller.login({ username: 'admin', password: 'sensitive-value' })
    expect(controller.state()).toEqual({ status: 'authenticated', profile, error: null })
    expect(JSON.stringify(controller.state())).not.toMatch(/token|csrf|password|sensitive-value/i)
  })

  it('maps expired sessions to the stable unauthenticated state', async () => {
    const controller = createSessionController(client({ current: async () => { throw new SessionRequestError('authentication') } }))
    await controller.restore()
    expect(controller.state().status).toBe('unauthenticated')
  })

  it('invalidates the local identity after password change', async () => {
    const controller = createSessionController(client())
    await controller.restore()
    await controller.changePassword('old-sensitive', 'new-sensitive')
    expect(controller.state().status).toBe('unauthenticated')
  })

  it('keeps the authenticated profile when password validation fails', async () => {
    const controller = createSessionController(client({ changePassword: async () => { throw new SessionRequestError('validation') } }))
    await controller.restore()
    await controller.changePassword('wrong-old-value', 'invalid-new-value')
    expect(controller.state()).toEqual({ status: 'error', profile, error: 'validation' })
  })
})
