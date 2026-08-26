import { createDomainRegistry, defineDomain, nameRouteComponent } from './domain'

const loader = async() => ({ default: { name: 'Original' } })
const demo = () => defineDomain({
  id: 'demo',
  routes: [{ routeKey: 'demo.product', legacyComponent: '/demo/product/index', load: loader }]
})

describe('Domain registry', () => {
  it('resolves the stable routeKey before the legacy component path', () => {
    const registry = createDomainRegistry([demo()])
    expect(registry.resolve({ routeKey: 'demo.product', legacyComponent: '/unknown' }))
      .toMatchObject({ status: 'resolved', domainId: 'demo', route: { routeKey: 'demo.product' } })
  })

  it('uses the exact legacy whitelist only when routeKey is absent', () => {
    const registry = createDomainRegistry([demo()])
    expect(registry.resolve({ legacyComponent: '/demo/product/index' })).toMatchObject({ status: 'resolved' })
    expect(registry.resolve({ legacyComponent: '/demo/product/../secret' })).toEqual({
      status: 'unavailable',
      reason: 'unknown-legacy-component',
      identity: '/demo/product/../secret'
    })
  })

  it('fails closed for an unknown routeKey instead of falling back to component', () => {
    const registry = createDomainRegistry([demo()])
    expect(registry.resolve({ routeKey: 'demo.missing', legacyComponent: '/demo/product/index' })).toEqual({
      status: 'unavailable',
      reason: 'unknown-route-key',
      identity: 'demo.missing'
    })
  })

  it.each([
    () => createDomainRegistry([demo(), demo()]),
    () => createDomainRegistry([demo(), defineDomain({
      id: 'other',
      routes: [{ routeKey: 'other.product', legacyComponent: '/demo/product/index', load: loader }]
    })]),
    () => defineDomain({ id: 'demo', routes: [{ routeKey: 'system.user', legacyComponent: '/x', load: loader }] }),
    () => defineDomain({ id: 'Demo', routes: [{ routeKey: 'Demo.user', legacyComponent: '/x', load: loader }] })
  ])('rejects duplicate or unstable registrations', register => {
    expect(register).toThrow()
  })

  it('preserves keep-alive by naming a copy of the loaded component', async() => {
    const first = await nameRouteComponent(loader, 'First')()
    const second = await nameRouteComponent(loader, 'Second')()
    expect(first).toMatchObject({ name: 'First' })
    expect(second).toMatchObject({ name: 'Second' })
  })
})
