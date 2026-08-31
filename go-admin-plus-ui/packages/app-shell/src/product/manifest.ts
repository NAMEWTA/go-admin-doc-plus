import type { Component } from 'vue'

export type ProductHost = 'web' | 'desktop'

export type ProductModuleId =
  | 'iam'
  | 'organization'
  | 'audit'
  | 'settings'
  | 'generator'
  | 'scheduler'
  | 'demo'
  | 'files'

export type ProductIcon =
  | 'users'
  | 'shield'
  | 'menu'
  | 'building'
  | 'briefcase'
  | 'file-clock'
  | 'settings'
  | 'book-open'
  | 'wand'
  | 'calendar-clock'
  | 'history'
  | 'package'
  | 'files'

export interface ProductRoute {
  readonly name: string
  readonly module: ProductModuleId
  readonly title: string
  readonly path: `/${string}`
  readonly permission: `${string}.${string}`
  readonly icon: ProductIcon
  readonly order: number
  readonly component: () => Promise<Component>
}

export interface ProductModule {
  readonly id: ProductModuleId
  readonly title: string
  readonly hosts: readonly ProductHost[]
  readonly routes: readonly ProductRoute[]
}

const bothHosts = ['web', 'desktop'] as const satisfies readonly ProductHost[]

export const productModules = [
  {
    id: 'iam',
    title: '权限管理',
    hosts: bothHosts,
    routes: [
      { name: 'iam-users', module: 'iam', title: '用户管理', path: '/iam/users', permission: 'iam.users.read', icon: 'users', order: 10, component: async () => (await import('@go-admin-plus/web-domain-iam/administration')).AdministrationPage },
      { name: 'iam-roles', module: 'iam', title: '角色管理', path: '/iam/roles', permission: 'iam.roles.read', icon: 'shield', order: 20, component: async () => (await import('@go-admin-plus/web-domain-iam/administration')).AdministrationPage },
      { name: 'iam-menus', module: 'iam', title: '菜单管理', path: '/iam/menus', permission: 'iam.menus.read', icon: 'menu', order: 30, component: async () => (await import('@go-admin-plus/web-domain-iam/administration')).AdministrationPage }
    ]
  },
  {
    id: 'audit',
    title: '审计管理',
    hosts: bothHosts,
    routes: [
      { name: 'audit-records', module: 'audit', title: '审计日志', path: '/audit/records', permission: 'audit.records.read', icon: 'file-clock', order: 100, component: async () => (await import('@go-admin-plus/web-domain-audit')).AuditPage }
    ]
  },
  {
    id: 'organization',
    title: '组织管理',
    hosts: bothHosts,
    routes: [
      { name: 'organization-departments', module: 'organization', title: '部门管理', path: '/organization/departments', permission: 'organization.departments.read', icon: 'building', order: 200, component: async () => (await import('@go-admin-plus/web-domain-organization')).OrganizationPage },
      { name: 'organization-positions', module: 'organization', title: '岗位管理', path: '/organization/positions', permission: 'organization.positions.read', icon: 'briefcase', order: 210, component: async () => (await import('@go-admin-plus/web-domain-organization')).OrganizationPage }
    ]
  },
  {
    id: 'settings',
    title: '系统设置',
    hosts: bothHosts,
    routes: [
      { name: 'settings-values', module: 'settings', title: '参数设置', path: '/settings/values', permission: 'settings.values.read', icon: 'settings', order: 300, component: async () => (await import('@go-admin-plus/web-domain-settings')).SettingsPage },
      { name: 'settings-dictionaries', module: 'settings', title: '字典管理', path: '/settings/dictionaries', permission: 'settings.dictionaries.read', icon: 'book-open', order: 310, component: async () => (await import('@go-admin-plus/web-domain-settings')).SettingsPage }
    ]
  },
  {
    id: 'generator',
    title: '开发工具',
    hosts: bothHosts,
    routes: [
      { name: 'code-generator', module: 'generator', title: '代码生成', path: '/generator', permission: 'generator.metadata.read', icon: 'wand', order: 400, component: async () => (await import('@go-admin-plus/web-domain-generator')).GeneratorWizardPage }
    ]
  },
  {
    id: 'scheduler',
    title: '任务调度',
    hosts: bothHosts,
    routes: [
      { name: 'scheduler-definitions', module: 'scheduler', title: '任务定义', path: '/scheduler/definitions', permission: 'scheduler.definitions.read', icon: 'calendar-clock', order: 500, component: async () => (await import('@go-admin-plus/web-domain-scheduler')).SchedulerPage },
      { name: 'scheduler-executions', module: 'scheduler', title: '执行记录', path: '/scheduler/executions', permission: 'scheduler.executions.read', icon: 'history', order: 510, component: async () => (await import('@go-admin-plus/web-domain-scheduler')).SchedulerPage }
    ]
  },
  {
    id: 'demo',
    title: '示例业务',
    hosts: bothHosts,
    routes: [
      { name: 'demo-products', module: 'demo', title: '产品示例', path: '/demo/products', permission: 'demo.products.read', icon: 'package', order: 600, component: async () => (await import('@go-admin-plus/web-domain-demo')).DemoProductsPage }
    ]
  },
  {
    id: 'files',
    title: '文件管理',
    hosts: bothHosts,
    routes: [
      { name: 'files-objects', module: 'files', title: '文件管理', path: '/files', permission: 'files.objects.read', icon: 'files', order: 700, component: async () => (await import('@go-admin-plus/web-domain-files')).FilesPage }
    ]
  }
] as const satisfies readonly ProductModule[]

const permissionPattern = /^[a-z][a-z0-9-]*(\.[a-z][a-z0-9-]*)+$/
const routeNamePattern = /^[a-z][a-z0-9-]*$/

export const assertProductManifest = (modules: readonly ProductModule[] = productModules): void => {
  const moduleIds = new Set<string>()
  const routeNames = new Set<string>()
  const routePaths = new Set<string>()
  let previousGlobalOrder = -1

  for (const productModule of modules) {
    if (moduleIds.has(productModule.id) || !productModule.title.trim() || productModule.hosts.length === 0 || productModule.routes.length === 0) {
      throw new Error('product module definition is invalid')
    }
    moduleIds.add(productModule.id)

    for (const host of bothHosts) {
      if (!productModule.hosts.includes(host)) throw new Error('product host coverage is incomplete')
    }

    for (const route of productModule.routes) {
      if (
        route.module !== productModule.id
        || routeNames.has(route.name)
        || routePaths.has(route.path)
        || !routeNamePattern.test(route.name)
        || !route.title.trim()
        || !permissionPattern.test(route.permission)
        || typeof route.component !== 'function'
      ) {
        throw new Error('product route definition is invalid')
      }
      if (route.order <= previousGlobalOrder) throw new Error('product route order is unstable')
      routeNames.add(route.name)
      routePaths.add(route.path)
      previousGlobalOrder = route.order
    }
  }
}

export const productRoutesFor = (host: ProductHost): readonly ProductRoute[] =>
  (productModules as readonly ProductModule[]).flatMap(productModule =>
    productModule.hosts.includes(host) ? productModule.routes : []
  )

export const productModuleFor = (id: ProductModuleId): ProductModule =>
  (productModules as readonly ProductModule[]).find(productModule => productModule.id === id) as ProductModule

assertProductManifest()
