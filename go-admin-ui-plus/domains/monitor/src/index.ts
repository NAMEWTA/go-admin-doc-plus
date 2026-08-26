import { defineDomain } from './app-core'

export const monitorDomain = defineDomain({
  id: 'monitor',
  routes: [
    { routeKey: 'monitor.server', legacyComponent: '/sys-tools/monitor', load: () => import('./pages/monitor.vue') }
  ]
})

export * from './api/server'
