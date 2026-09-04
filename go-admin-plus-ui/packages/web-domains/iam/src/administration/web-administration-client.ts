import { AdministrationRequestError, createContractClient, type AdministrationClient } from '@go-admin-plus/domain-iam/administration'

interface Problem { category?: string; code?: string }

export const createWebAdministrationClient = (fetcher: typeof fetch = fetch, baseUrl = '/api'): AdministrationClient => {
  let csrf = ''
  let responseFailure: string | null = null
  let requestTail: Promise<void> = Promise.resolve()
  const serialized = <T>(operation: () => Promise<T>): Promise<T> => {
    const result = requestTail.then(operation, operation)
    requestTail = result.then(() => undefined, () => undefined)
    return result
  }
  const contract = createContractClient({
    baseUrl,
    fetch: async (input) => {
      const headers = new Headers(input.headers)
      if (csrf && input.method !== 'GET') headers.set('X-CSRF-Token', csrf)
      const response = await fetcher(new Request(input, { credentials: 'include', headers }))
      const next = response.headers.get('X-CSRF-Token')
      const body = response.status >= 400 ? await response.clone().json().catch(() => null) as Problem | null : null
      responseFailure = classifyResponse(response.status, body)
      if (next) csrf = next
      else if (responseFailure === 'relogin') csrf = ''
      return response
    },
  })
  const unwrap = <T>(data: T | undefined, error: unknown): T => {
    if (error !== undefined) throw new AdministrationRequestError(takeFailure(problemCategory(error)))
    if (data === undefined) throw new AdministrationRequestError('unavailable')
    return data
  }
  const complete = (error: unknown) => { if (error !== undefined) throw new AdministrationRequestError(takeFailure(problemCategory(error))) }
  const takeFailure = (fallback: string) => {
    const result = responseFailure ?? fallback
    responseFailure = null
    return result
  }
  return {
    manifest: () => serialized(async () => { const result = await contract.GET('/iam/administration/manifest'); return unwrap(result.data, result.error) }),
    listUsers: (search, page, pageSize) => serialized(async () => { const result = await contract.GET('/iam/administration/users', { params: { query: { search, page, pageSize } } }); return unwrap(result.data, result.error) }),
    createUser: (body) => serialized(async () => { const result = await contract.POST('/iam/administration/users', { body }); return unwrap(result.data, result.error) }),
    updateUser: (userId, body) => serialized(async () => { const result = await contract.PATCH('/iam/administration/users/{userId}', { params: { path: { userId } }, body }); return unwrap(result.data, result.error) }),
    startUserDeletion: (userId, body) => serialized(async () => { const result = await contract.POST('/iam/administration/users/{userId}/deletion', { params: { path: { userId } }, body }); return unwrap(result.data, result.error) }),
    getUserDeletion: (userId) => serialized(async () => { const result = await contract.GET('/iam/administration/users/{userId}/deletion', { params: { path: { userId } } }); return unwrap(result.data, result.error) }),
    cancelUserDeletion: (userId) => serialized(async () => { const result = await contract.POST('/iam/administration/users/{userId}/deletion/cancel', { params: { path: { userId } } }); complete(result.error) }),
    setUserRoles: (userId, roleIds) => serialized(async () => { const result = await contract.PUT('/iam/administration/users/{userId}/roles', { params: { path: { userId } }, body: { roleIds: [...roleIds] } }); complete(result.error) }),
    resetPassword: (userId, password) => serialized(async () => { const result = await contract.PUT('/iam/administration/users/{userId}/password', { params: { path: { userId } }, body: { password } }); complete(result.error) }),
    listRoles: () => serialized(async () => { const result = await contract.GET('/iam/administration/roles'); return unwrap(result.data, result.error) }),
    createRole: (body) => serialized(async () => { const result = await contract.POST('/iam/administration/roles', { body }); return unwrap(result.data, result.error) }),
    updateRole: (roleId, body) => serialized(async () => { const result = await contract.PATCH('/iam/administration/roles/{roleId}', { params: { path: { roleId } }, body }); complete(result.error) }),
    deleteRole: (roleId) => serialized(async () => { const result = await contract.DELETE('/iam/administration/roles/{roleId}', { params: { path: { roleId } } }); complete(result.error) }),
    setRoleGrants: (roleId, permissionCodes, menuIds) => serialized(async () => { const result = await contract.PUT('/iam/administration/roles/{roleId}/grants', { params: { path: { roleId } }, body: { permissionCodes: [...permissionCodes], menuIds: [...menuIds] } }); complete(result.error) }),
    listMenus: () => serialized(async () => { const result = await contract.GET('/iam/administration/menus'); return unwrap(result.data, result.error) }),
    createMenu: (body) => serialized(async () => { const result = await contract.POST('/iam/administration/menus', { body }); return unwrap(result.data, result.error) }),
    updateMenu: (menuId, body) => serialized(async () => { const result = await contract.PATCH('/iam/administration/menus/{menuId}', { params: { path: { menuId } }, body }); complete(result.error) }),
    deleteMenu: (menuId) => serialized(async () => { const result = await contract.DELETE('/iam/administration/menus/{menuId}', { params: { path: { menuId } } }); complete(result.error) }),
    listPermissions: () => serialized(async () => { const result = await contract.GET('/iam/administration/permissions'); return unwrap(result.data, result.error) }),
  }
}

const problemCategory = (value: unknown) => typeof value === 'object' && value !== null && 'category' in value
  ? String((value as Problem).category ?? 'unavailable')
  : 'unavailable'

const classifyResponse = (status: number, problem: Problem | null): string | null => {
  if (status === 401 || problem?.code === 'CSRF_REJECTED') return 'relogin'
  if (status === 403) return 'forbidden'
  if (status === 404) return 'not-found'
  if (status === 400 || status === 422) return 'validation'
  if (status === 409) return 'conflict'
  if (status >= 500) return 'unavailable'
  return null
}
