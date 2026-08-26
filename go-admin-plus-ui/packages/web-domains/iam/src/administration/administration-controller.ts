import { AdministrationRequestError, type AdministrationClient, type Menu, type MenuInput, type Permission, type Role, type User } from '@go-admin/domain-iam/administration'
import { createListController, type FormController, type FormRunResult, type ListController, type RemovalController, type RemovalRunResult } from '@go-admin/ui'

export interface UserFilters { search: string }
export interface CreateUserModel { username: string; displayName: string; email: string; password: string }
export interface CreateRoleModel { key: string; name: string; dataScope: 'all' | 'self' }

export interface AdministrationController {
  users: ListController<UserFilters, User, string>
  createUser: FormController<CreateUserModel>
  deleteUsers: RemovalController<string>
  roles(): ReadonlyArray<Role>
  menus(): ReadonlyArray<Menu>
  permissions(): ReadonlyArray<Permission>
  can(permissionCode: string): boolean
  failure(): AdministrationFailure | null
  clearFailure(): void
  hasPendingRepair(): boolean
  repairProjection(): Promise<CommandResult>
  refreshAuthorizationData(): Promise<void>
  createRole: FormController<CreateRoleModel>
  createMenu: FormController<MenuInput>
  readonly busy: boolean
  updateUser(user: User, enabled: boolean): Promise<CommandResult>
  deleteRole(id: string): Promise<CommandResult>
  updateRole(role: Role): Promise<CommandResult>
  deleteMenu(id: string): Promise<CommandResult>
  updateMenu(menu: Menu): Promise<CommandResult>
  setUserRoles(id: string, roleIds: ReadonlyArray<string>): Promise<CommandResult>
  setRoleGrants(id: string, permissionCodes: ReadonlyArray<string>, menuIds: ReadonlyArray<string>): Promise<CommandResult>
  resetPassword(id: string, password: string): Promise<CommandResult>
}

export type CommandResult = 'completed' | 'cancelled' | 'busy' | 'invalid' | 'failed' | 'refresh-failed'
export type AdministrationFailure = 'relogin' | 'forbidden' | 'validation' | 'conflict' | 'unavailable'

export const createAdministrationController = (client: AdministrationClient, confirm: (count: number) => Promise<boolean>): AdministrationController => {
  const users = createListController<UserFilters, User, string>({
    initialFilters: () => ({ search: '' }),
    rowKey: (row) => row.id,
    load: async ({ filters, page, pageSize }) => client.listUsers(filters.search, page, pageSize),
  })
  let roles: ReadonlyArray<Role> = []
  let menus: ReadonlyArray<Menu> = []
  let permissions: ReadonlyArray<Permission> = []
  let capabilityCodes = new Set<string>()
  let capabilityScope: 'all' | 'self' = 'self'
  let failure: AdministrationFailure | null = null
  const pendingRepairs = new Map<string, () => Promise<void>>()
  let repairBusy = false
  const recordFailure = (error: unknown) => {
    failure = error instanceof AdministrationRequestError && isAdministrationFailure(error.category)
      ? error.category
      : 'unavailable'
  }
  const clearFailure = () => { failure = null }
  const refreshAuthorizationData = async () => {
    clearFailure()
    try {
      const manifest = await client.manifest()
      capabilityCodes = new Set(manifest.permissionCodes)
      capabilityScope = manifest.dataScope
      const can = (permissionCode: string) => capabilityCodes.has(permissionCode) && (capabilityScope === 'all' || permissionCode === 'iam.users.read')
      const [nextRoles, nextMenus, nextPermissions] = await Promise.all([
        can('iam.roles.read') ? client.listRoles() : Promise.resolve([]),
        can('iam.menus.read') ? client.listMenus() : Promise.resolve([]),
        can('iam.permissions.read') ? client.listPermissions() : Promise.resolve([]),
      ])
      roles = [...nextRoles]; menus = [...nextMenus]; permissions = [...nextPermissions]
    } catch (error) {
      capabilityCodes = new Set(); capabilityScope = 'self'; roles = []; menus = []; permissions = []
      recordFailure(error)
      throw error
    }
  }
  const form = <TModel>(key: string, options: {
    validate(model: Readonly<TModel>): Promise<boolean>
    submit(model: Readonly<TModel>): Promise<void>
    refresh(): Promise<void>
  }): FormController<TModel> => {
    let busy = false
    const refresh = async (): Promise<FormRunResult> => {
      clearFailure()
      try { await options.refresh(); pendingRepairs.delete(key); return 'submitted' }
      catch (error) { recordFailure(error); return 'refresh-failed' }
    }
    return {
      get busy() { return busy },
      async run(model) {
        if (busy) return 'busy'
        busy = true
        try {
          if (pendingRepairs.has(key)) return await refresh()
          if (!await options.validate(model)) return 'invalid'
          clearFailure()
          try { await options.submit(model) } catch (error) { recordFailure(error); return 'failed' }
          pendingRepairs.set(key, options.refresh)
          return await refresh()
        } finally { busy = false }
      },
    }
  }
  const removal = <TKey>(key: string, options: {
    execute(keys: ReadonlyArray<TKey>): Promise<void>
    refresh(): Promise<void>
    clearSelection(): void
  }): RemovalController<TKey> => {
    let busy = false
    const repair = async () => { options.clearSelection(); await options.refresh() }
    const refresh = async (): Promise<RemovalRunResult> => {
      clearFailure()
      try { await repair(); pendingRepairs.delete(key); return 'completed' }
      catch (error) { recordFailure(error); return 'refresh-failed' }
    }
    return {
      get busy() { return busy },
      async run(keys) {
        if (busy) return 'busy'
        if (pendingRepairs.has(key)) { busy = true; try { return await refresh() } finally { busy = false } }
        if (keys.length === 0) return 'empty'
        busy = true
        try {
          if (!await confirm(keys.length)) return 'cancelled'
          clearFailure()
          try { await options.execute([...keys]) } catch (error) { recordFailure(error); return 'failed' }
          pendingRepairs.set(key, repair)
          return await refresh()
        } finally { busy = false }
      },
    }
  }
  const createUser = form<CreateUserModel>('create-user', {
    validate: async (model) => validName(model.username, 64) && validName(model.displayName, 80) && model.email.includes('@') && model.password.length >= 12,
    submit: async (model) => { await client.createUser(model) },
    refresh: () => users.refresh(),
  })
  const deleteUsers = removal<string>('delete-users', {
    execute: (ids) => client.deleteUsers(ids),
    refresh: () => users.refresh(),
    clearSelection: () => users.clearSelection(),
  })
  const createRole = form<CreateRoleModel>('create-role', {
    validate: async (model) => validStableKey(model.key) && validName(model.name, 100),
    submit: async (model) => { await client.createRole(model) },
    refresh: refreshAuthorizationData,
  })
  const createMenu = form<MenuInput>('create-menu', {
    validate: async (model) => validStableKey(model.key) && validName(model.label, 80) && model.path.startsWith('/') && model.permissionCode.length >= 3,
    submit: async (model) => { await client.createMenu(model) },
    refresh: refreshAuthorizationData,
  })
  let commandBusy = false
  const command = async (key: string, operation: () => Promise<void>, refreshed: () => Promise<void>, destructive = false): Promise<CommandResult> => {
    if (commandBusy) return 'busy'
    commandBusy = true
    clearFailure()
    try {
      if (pendingRepairs.has(key)) {
        try { await refreshed(); pendingRepairs.delete(key); return 'completed' }
        catch (error) { recordFailure(error); return 'refresh-failed' }
      }
      if (destructive && !await confirm(1)) return 'cancelled'
      try { await operation() } catch (error) { recordFailure(error); return 'failed' }
      pendingRepairs.set(key, refreshed)
      try { await refreshed(); pendingRepairs.delete(key); return 'completed' } catch (error) { recordFailure(error); return 'refresh-failed' }
    } finally { commandBusy = false }
  }
  const repairProjection = async (): Promise<CommandResult> => {
    if (repairBusy || commandBusy) return 'busy'
    const next = pendingRepairs.entries().next().value as [string, () => Promise<void>] | undefined
    if (!next) return 'completed'
    repairBusy = true
    clearFailure()
    try {
      try { await next[1](); pendingRepairs.delete(next[0]); return 'completed' }
      catch (error) { recordFailure(error); return 'refresh-failed' }
    } finally { repairBusy = false }
  }
  return {
    users, createUser, deleteUsers, createRole, createMenu,
    roles: () => [...roles], menus: () => [...menus], permissions: () => [...permissions], refreshAuthorizationData,
    can: (permissionCode) => capabilityCodes.has(permissionCode) && (capabilityScope === 'all' || permissionCode === 'iam.users.read'),
    failure: () => failure,
    clearFailure,
    hasPendingRepair: () => pendingRepairs.size > 0,
    repairProjection,
    get busy() { return commandBusy || repairBusy },
    updateUser(user, enabled) { return command(`update-user:${user.id}`, async () => { await client.updateUser(user.id, { displayName: user.displayName, email: user.email, enabled }) }, () => users.refresh(), !enabled) },
    deleteRole(id) { return command(`delete-role:${id}`, () => client.deleteRole(id), refreshAuthorizationData, true) },
    updateRole(role) { if (!validStableKey(role.key) || !validName(role.name, 100)) return Promise.resolve('invalid'); return command(`update-role:${role.id}`, () => client.updateRole(role.id, { key: role.key, name: role.name, dataScope: role.dataScope, enabled: role.enabled }), refreshAuthorizationData, !role.enabled) },
    deleteMenu(id) { return command(`delete-menu:${id}`, () => client.deleteMenu(id), refreshAuthorizationData, true) },
    updateMenu(menu) { if (!validStableKey(menu.key)) return Promise.resolve('invalid'); return command(`update-menu:${menu.id}`, () => client.updateMenu(menu.id, { key: menu.key, label: menu.label, path: menu.path, permissionCode: menu.permissionCode, sortOrder: menu.sortOrder }), refreshAuthorizationData) },
    setUserRoles(id, roleIds) { return command(`set-user-roles:${id}`, () => client.setUserRoles(id, roleIds), () => users.refresh()) },
    setRoleGrants(id, permissionCodes, menuIds) { return command(`set-role-grants:${id}`, () => client.setRoleGrants(id, permissionCodes, menuIds), refreshAuthorizationData) },
    resetPassword(id, password) { if (password.length < 12) return Promise.resolve('invalid'); return command(`reset-password:${id}`, () => client.resetPassword(id, password), () => users.refresh(), true) },
  }
}

export const resetPasswordAndClear = async (controller: AdministrationController, id: string, password: string, clear: () => void) => {
  try { return await controller.resetPassword(id, password) }
  finally { clear() }
}

export const createUserAndClearPassword = async (controller: AdministrationController, model: CreateUserModel, clear: () => void) => {
  try { return await controller.createUser.run(model) }
  finally { clear() }
}

export const settleAdministrationPageOperation = async (operation: () => Promise<unknown>, settled: () => void): Promise<void> => {
  try { await operation() }
  catch { /* The controller owns stable failure classification; page operations must settle. */ }
  finally { settled() }
}

const isAdministrationFailure = (value: string): value is AdministrationFailure =>
  ['relogin', 'forbidden', 'validation', 'conflict', 'unavailable'].includes(value)

const validName = (value: string, maximum: number) => value.trim().length >= 3 && value.length <= maximum
const validStableKey = (value: string) => value.length >= 3 && value.length <= 64 && /^[a-z0-9][a-z0-9_-]*$/.test(value)
