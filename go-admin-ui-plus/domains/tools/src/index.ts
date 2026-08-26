import { defineDomain } from './app-core'

export const toolsDomain = defineDomain({
  id: 'tools',
  routes: [
    { routeKey: 'tools.swagger', legacyComponent: '/dev-tools/swagger/index', load: () => import('./pages/swagger/index.vue') },
    { routeKey: 'tools.generator', legacyComponent: '/dev-tools/gen/index', load: () => import('./pages/gen/index.vue') },
    { routeKey: 'tools.generator-edit', legacyComponent: '/dev-tools/gen/editTable', load: () => import('./pages/gen/editTable.vue') },
    { routeKey: 'tools.form-builder', legacyComponent: '/dev-tools/build/index', load: () => import('./pages/build/index.vue') }
  ]
})

export * from './api/gen'
