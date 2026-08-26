import { describe, expect, it } from 'vitest'
import { AdministrationRequestError, createCapabilityController, type Manifest } from './administration-controller'

const manifest: Manifest = { dataScope: 'self', permissionCodes: ['iam.users.read'], menus: [{ key: 'iam-users', label: 'Users', path: '/iam/users', permissionCode: 'iam.users.read', sortOrder: 10 }] }

describe('capability controller', () => {
  it('projects only server-issued permission codes and routes', async () => {
    const controller = createCapabilityController({ manifest: async () => manifest })
    await controller.refresh()
    expect(controller.can('iam.users.read')).toBe(true)
    expect(controller.can('iam.users.write')).toBe(false)
    expect(controller.route('/iam/users')).toBe('allowed')
    expect(controller.route('/iam/roles')).toBe('unauthorized')
    expect(controller.route('/unknown')).toBe('not-found')
  })

  it('discards stale manifest results and fails closed', async () => {
    let resolveFirst!: (value: Manifest) => void
    let call = 0
    const controller = createCapabilityController({ manifest: () => ++call === 1 ? new Promise((resolve) => { resolveFirst = resolve }) : Promise.reject(new AdministrationRequestError('authorization')) })
    const first = controller.refresh(); await controller.refresh(); resolveFirst(manifest); await first
    expect(controller.state()).toEqual({ status: 'unauthorized', manifest: null, error: 'authorization' })
    expect(controller.can('iam.users.read')).toBe(false)
  })
})
