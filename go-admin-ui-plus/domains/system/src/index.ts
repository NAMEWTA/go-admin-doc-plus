import { defineDomain } from './app-core'

export const systemDomain = defineDomain({
  id: 'system',
  routes: [
    { routeKey: 'system.user', legacyComponent: '/admin/sys-user/index', load: () => import('./pages/sys-user/index.vue') },
    { routeKey: 'system.menu', legacyComponent: '/admin/sys-menu/index', load: () => import('./pages/sys-menu/index.vue') },
    { routeKey: 'system.role', legacyComponent: '/admin/sys-role/index', load: () => import('./pages/sys-role/index.vue') },
    { routeKey: 'system.department', legacyComponent: '/admin/sys-dept/index', load: () => import('./pages/sys-dept/index.vue') },
    { routeKey: 'system.post', legacyComponent: '/admin/sys-post/index', load: () => import('./pages/sys-post/index.vue') },
    { routeKey: 'system.dictionary', legacyComponent: '/admin/dict/index', load: () => import('./pages/dict/index.vue') },
    { routeKey: 'system.dictionary-data', legacyComponent: '/admin/dict/data', load: () => import('./pages/dict/data.vue') },
    { routeKey: 'system.config', legacyComponent: '/admin/sys-config/index', load: () => import('./pages/sys-config/index.vue') },
    { routeKey: 'system.config-settings', legacyComponent: '/admin/sys-config/set', load: () => import('./pages/sys-config/set.vue') },
    { routeKey: 'system.api', legacyComponent: '/admin/sys-api/index', load: () => import('./pages/sys-api/index.vue') },
    { routeKey: 'system.login-log', legacyComponent: '/admin/sys-login-log/index', load: () => import('./pages/sys-login-log/index.vue') },
    { routeKey: 'system.operation-log', legacyComponent: '/admin/sys-oper-log/index', load: () => import('./pages/sys-oper-log/index.vue') }
  ]
})

export * from './types'
export * from './api/sys-role'
export * from './api/dict/data'
export * from './api/dict/type'
