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

export interface ProductRoute {
  readonly key: string
  readonly label: string
  readonly path: `/${string}`
  readonly permission: `${string}.${string}`
  readonly order: number
}

export interface ProductModule {
  readonly id: ProductModuleId
  readonly hosts: readonly ProductHost[]
  readonly routes: readonly ProductRoute[]
}

const bothHosts = ['web', 'desktop'] as const satisfies readonly ProductHost[]

export const productModules = [
  {
    id: 'iam',
    hosts: bothHosts,
    routes: [
      { key: 'iam-users', label: 'Users', path: '/iam/users', permission: 'iam.users.read', order: 10 },
      { key: 'iam-roles', label: 'Roles', path: '/iam/roles', permission: 'iam.roles.read', order: 20 },
      { key: 'iam-menus', label: 'Menus', path: '/iam/menus', permission: 'iam.menus.read', order: 30 }
    ]
  },
  {
    id: 'audit',
    hosts: bothHosts,
    routes: [
      { key: 'audit-records', label: 'Audit records', path: '/audit/records', permission: 'audit.records.read', order: 500 }
    ]
  },
  {
    id: 'organization',
    hosts: bothHosts,
    routes: [
      { key: 'organization-departments', label: 'Departments', path: '/organization/departments', permission: 'organization.departments.read', order: 600 },
      { key: 'organization-positions', label: 'Positions', path: '/organization/positions', permission: 'organization.positions.read', order: 610 }
    ]
  },
  {
    id: 'settings',
    hosts: bothHosts,
    routes: [
      { key: 'settings-values', label: 'Settings', path: '/settings/values', permission: 'settings.values.read', order: 600 },
      { key: 'settings-dictionaries', label: 'Dictionaries', path: '/settings/dictionaries', permission: 'settings.dictionaries.read', order: 610 }
    ]
  },
  {
    id: 'generator',
    hosts: bothHosts,
    routes: [
      { key: 'code-generator', label: 'Code generator', path: '/generator', permission: 'generator.metadata.read', order: 700 }
    ]
  },
  {
    id: 'scheduler',
    hosts: bothHosts,
    routes: [
      { key: 'scheduler-definitions', label: 'Task schedules', path: '/scheduler/definitions', permission: 'scheduler.definitions.read', order: 700 },
      { key: 'scheduler-executions', label: 'Task executions', path: '/scheduler/executions', permission: 'scheduler.executions.read', order: 710 }
    ]
  },
  {
    id: 'demo',
    hosts: bothHosts,
    routes: [
      { key: 'demo-products', label: 'Demo products', path: '/demo/products', permission: 'demo.products.read', order: 800 }
    ]
  },
  {
    id: 'files',
    hosts: bothHosts,
    routes: [
      { key: 'files-objects', label: 'Files', path: '/files', permission: 'files.objects.read', order: 900 }
    ]
  }
] as const satisfies readonly ProductModule[]

const permissionPattern = /^[a-z][a-z0-9-]*(\.[a-z][a-z0-9-]*)+$/

export const assertProductManifest = (modules: readonly ProductModule[] = productModules): void => {
  const moduleIds = new Set<string>()
  const routeKeys = new Set<string>()
  const routePaths = new Set<string>()

  for (const module of modules) {
    if (moduleIds.has(module.id) || module.hosts.length === 0 || module.routes.length === 0) {
      throw new Error('product module definition is invalid')
    }
    moduleIds.add(module.id)
    let previousOrder = -1

    for (const host of bothHosts) {
      if (!module.hosts.includes(host)) throw new Error('product host coverage is incomplete')
    }

    for (const route of module.routes) {
      if (routeKeys.has(route.key) || routePaths.has(route.path) || !permissionPattern.test(route.permission)) {
        throw new Error('product route definition is invalid')
      }
      if (route.order <= previousOrder) throw new Error('product route order is unstable')
      routeKeys.add(route.key)
      routePaths.add(route.path)
      previousOrder = route.order
    }
  }
}

export const productRoutesFor = (host: ProductHost): readonly ProductRoute[] =>
  (productModules as readonly ProductModule[]).flatMap(module => module.hosts.includes(host) ? module.routes : [])

assertProductManifest()
