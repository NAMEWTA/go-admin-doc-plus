import type { Component } from 'vue'

export type RouteComponentModule = Component | { readonly default?: Component }
export type RouteComponentLoader = () => Promise<RouteComponentModule>

export interface DomainRoute {
  readonly routeKey: string
  readonly legacyComponent: string
  readonly load: RouteComponentLoader
}

export interface DomainDefinition {
  readonly id: string
  readonly routes: ReadonlyArray<DomainRoute>
}

export type RouteResolution =
  | { readonly status: 'resolved', readonly domainId: string, readonly route: DomainRoute }
  | { readonly status: 'unavailable', readonly reason: 'missing-route-identity' | 'unknown-route-key' | 'unknown-legacy-component', readonly identity?: string }

export interface DomainRegistry {
  readonly domains: ReadonlyArray<DomainDefinition>
  resolve(input: { readonly routeKey?: string, readonly legacyComponent?: string }): RouteResolution
}

const domainIdPattern = /^[a-z][a-z0-9-]*$/
const routeSegmentPattern = /^[a-z][a-z0-9-]*$/

const requireDomain = (domain: DomainDefinition) => {
  if (!domain || !domainIdPattern.test(domain.id)) {
    throw new Error(`invalid domain id ${JSON.stringify(domain?.id)}`)
  }
  if (!Array.isArray(domain.routes) || domain.routes.length === 0) {
    throw new Error(`domain ${domain.id} must register at least one route`)
  }
}

const requireRoute = (domain: DomainDefinition, route: DomainRoute) => {
  const segments = route.routeKey.split('.')
  if (segments.length < 2 || segments[0] !== domain.id || segments.some(segment => !routeSegmentPattern.test(segment))) {
    throw new Error(`routeKey ${JSON.stringify(route.routeKey)} must be a stable ${domain.id}.* key`)
  }
  if (!route.legacyComponent.startsWith('/') || route.legacyComponent.includes('..')) {
    throw new Error(`legacy component ${JSON.stringify(route.legacyComponent)} must be a rooted logical path`)
  }
  if (typeof route.load !== 'function') {
    throw new Error(`routeKey ${route.routeKey} must provide a component loader`)
  }
}

export const defineDomain = (domain: DomainDefinition): DomainDefinition => {
  requireDomain(domain)
  for (const route of domain.routes) requireRoute(domain, route)
  return Object.freeze({
    id: domain.id,
    routes: Object.freeze(domain.routes.map(route => Object.freeze({ ...route })))
  })
}

export const createDomainRegistry = (definitions: ReadonlyArray<DomainDefinition>): DomainRegistry => {
  const domains = definitions.map(defineDomain)
  const domainIds = new Set<string>()
  const byRouteKey = new Map<string, { domainId: string, route: DomainRoute }>()
  const byLegacyComponent = new Map<string, { domainId: string, route: DomainRoute }>()

  for (const domain of domains) {
    if (domainIds.has(domain.id)) throw new Error(`duplicate domain id ${domain.id}`)
    domainIds.add(domain.id)
    for (const route of domain.routes) {
      if (byRouteKey.has(route.routeKey)) throw new Error(`duplicate routeKey ${route.routeKey}`)
      if (byLegacyComponent.has(route.legacyComponent)) {
        throw new Error(`duplicate legacy component ${route.legacyComponent}`)
      }
      const registration = { domainId: domain.id, route }
      byRouteKey.set(route.routeKey, registration)
      byLegacyComponent.set(route.legacyComponent, registration)
    }
  }

  return Object.freeze({
    domains: Object.freeze(domains),
    resolve({ routeKey, legacyComponent }: { readonly routeKey?: string, readonly legacyComponent?: string }) {
      const normalizedKey = routeKey?.trim()
      if (normalizedKey) {
        const registered = byRouteKey.get(normalizedKey)
        return registered
          ? { status: 'resolved' as const, ...registered }
          : { status: 'unavailable' as const, reason: 'unknown-route-key' as const, identity: normalizedKey }
      }
      const normalizedLegacy = legacyComponent?.trim()
      if (normalizedLegacy) {
        const registered = byLegacyComponent.get(normalizedLegacy)
        return registered
          ? { status: 'resolved' as const, ...registered }
          : { status: 'unavailable' as const, reason: 'unknown-legacy-component' as const, identity: normalizedLegacy }
      }
      return { status: 'unavailable' as const, reason: 'missing-route-identity' as const }
    }
  })
}

export const nameRouteComponent = (loader: RouteComponentLoader, name?: string): RouteComponentLoader => {
  if (!name) return loader
  return async() => {
    const loaded = await loader()
    const component = ((loaded as { default?: Component })?.default ?? loaded) as Component & { readonly name?: string }
    if (!component || component.name === name) return component
    return { ...component, name }
  }
}
