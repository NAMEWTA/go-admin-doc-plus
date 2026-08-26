import { defineDomain } from './app-core'

export const jobsDomain = defineDomain({
  id: 'jobs',
  routes: [
    { routeKey: 'jobs.schedule', legacyComponent: '/schedule/index', load: () => import('./pages/index.vue') },
    { routeKey: 'jobs.log', legacyComponent: '/schedule/log', load: () => import('./pages/log.vue') }
  ]
})

export * from './types'
export * from './api/sys-job'
