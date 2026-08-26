import { AdministrationRequestError, type AdministrationClient, type Menu, type MenuInput, type Permission, type Role, type User } from '@go-admin/domain-iam/administration'
import { createFormController, createListController, createRemovalController, type FormController, type ListController, type RemovalController } from '@go-admin/ui'

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
  let failure: AdministrationFailure | null = null
  const recordFailure = (error: unknown) => {
    failure = error instanceof AdministrationRequestError && isAdministrationFailure(error.category)
      ? error.category
      : 'unavailable'
  }
  const clearFailure = () => { failure = null }
  const refreshAuthorizationData = async () => {
    const manifest = await client.manifest()
    capabilityCodes = new Set(manifest.permissionCodes)
    const [nextRoles, nextMenus, nextPermissions] = await Promise.all([
      capabilityCodes.has('iam.roles.read') ? client.listRoles() : Promise.resolve([]),
      capabilityCodes.has('iam.menus.read') ? client.listMenus() : Promise.resolve([]),
      capabilityCodes.has('iam.permissions.read') ? client.listPermissions() : Promise.resolve([]),
    ])
    roles = [...nextRoles]; menus = [...nextMenus]; permissions = [...nextPermissions]
  }
  const createUser = createFormController<CreateUserModel>({
    validate: async (model) => validName(model.username, 64) && validName(model.displayName, 80) && model.email.includes('@') && model.password.length >= 12,
    submit: async (model) => {
      clearFailure()
      try { await client.createUser(model) } catch (error) { recordFailure(error); throw error }
    },
    submitted: () => users.refresh(),
  })
  const deleteUsers = createRemovalController<string>({
    confirm,
    execute: async (ids) => {
      clearFailure()
      try { await client.deleteUsers(ids) } catch (error) { recordFailure(error); throw error }
    },
    refreshed: () => users.refresh(),
    clearSelection: () => users.clearSelection(),
  })
  const createRole = createFormController<CreateRoleModel>({
    validate: async (model) => validName(model.key, 64) && validName(model.name, 100),
    submit: async (model) => {
      clearFailure()
      try { await client.createRole(model) } catch (error) { recordFailure(error); throw error }
    },
    submitted: refreshAuthorizationData,
  })
  const createMenu = createFormController<MenuInput>({
    validate: async (model) => validName(model.key, 64) && validName(model.label, 80) && model.path.startsWith('/') && model.permissionCode.length >= 3,
    submit: async (model) => {
      clearFailure()
      try { await client.createMenu(model) } catch (error) { recordFailure(error); throw error }
    },
    submitted: refreshAuthorizationData,
  })
  let commandBusy = false
  const command = async (operation: () => Promise<void>, refreshed: () => Promise<void>, destructive = false): Promise<CommandResult> => {
    if (commandBusy) return 'busy'
    commandBusy = true
    clearFailure()
    try {
      if (destructive && !await confirm(1)) return 'cancelled'
      try { await operation() } catch (error) { recordFailure(error); return 'failed' }
      try { await refreshed(); return 'completed' } catch (error) { recordFailure(error); return 'refresh-failed' }
    } finally { commandBusy = false }
  }
  return {
    users, createUser, deleteUsers, createRole, createMenu,
    roles: () => [...roles], menus: () => [...menus], permissions: () => [...permissions], refreshAuthorizationData,
    can: (permissionCode) => capabilityCodes.has(permissionCode),
    failure: () => failure,
    clearFailure,
    get busy() { return commandBusy },
    updateUser(user, enabled) { return command(async () => { await client.updateUser(user.id, { displayName: user.displayName, email: user.email, enabled }) }, () => users.refresh(), !enabled) },
    deleteRole(id) { return command(() => client.deleteRole(id), refreshAuthorizationData, true) },
    updateRole(role) { return command(() => client.updateRole(role.id, { key: role.key, name: role.name, dataScope: role.dataScope, enabled: role.enabled }), refreshAuthorizationData, !role.enabled) },
    deleteMenu(id) { return command(() => client.deleteMenu(id), refreshAuthorizationData, true) },
    updateMenu(menu) { return command(() => client.updateMenu(menu.id, { key: menu.key, label: menu.label, path: menu.path, permissionCode: menu.permissionCode, sortOrder: menu.sortOrder }), refreshAuthorizationData) },
    setUserRoles(id, roleIds) { return command(() => client.setUserRoles(id, roleIds), () => users.refresh()) },
    setRoleGrants(id, permissionCodes, menuIds) { return command(() => client.setRoleGrants(id, permissionCodes, menuIds), refreshAuthorizationData) },
    resetPassword(id, password) { if (password.length < 12) return Promise.resolve('invalid'); return command(() => client.resetPassword(id, password), () => users.refresh(), true) },
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

const isAdministrationFailure = (value: string): value is AdministrationFailure =>
  ['relogin', 'forbidden', 'validation', 'conflict', 'unavailable'].includes(value)

const validName = (value: string, maximum: number) => value.trim().length >= 3 && value.length <= maximum
