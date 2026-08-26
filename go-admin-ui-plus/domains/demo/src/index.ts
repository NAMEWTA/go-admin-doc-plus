import { defineDomain } from './app-core'

export const demoDomain = defineDomain({
  id: 'demo',
  routes: [
    { routeKey: 'demo.product', legacyComponent: '/demo/product/index', load: () => import('./pages/product/index.vue') }
  ]
})

export * from './api/product'
