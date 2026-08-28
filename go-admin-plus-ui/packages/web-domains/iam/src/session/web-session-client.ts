import { createContractClient, SessionRequestError, type AccountProfile, type LoginCredentials, type SessionClient, type UpdateProfile } from '@go-admin-plus/domain-iam/session'

interface Problem { category?: string }

export const createWebSessionClient = (fetcher: typeof fetch = fetch, baseUrl = '/api'): SessionClient => {
  let csrf = ''
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
      const nextCSRF = response.headers.get('X-CSRF-Token')
      if (nextCSRF) csrf = nextCSRF
      if (response.status === 401 || response.status === 403) csrf = ''
      return response
    },
  })
  const unwrap = <T>(data: T | undefined, error: unknown): T => {
    if (error !== undefined) throw new SessionRequestError(problemCategory(error))
    if (data === undefined) throw new SessionRequestError('unavailable')
    return data
  }
  const complete = (error: unknown) => {
    if (error !== undefined) throw new SessionRequestError(problemCategory(error))
  }
  return {
    login: (credentials: LoginCredentials) => serialized(async () => {
      const result = await contract.POST('/iam/session/login', { body: credentials })
      const session = unwrap(result.data, result.error)
      csrf = session.csrfToken
      return session.profile
    }),
    current: () => serialized(async () => {
      const result = await contract.GET('/iam/session/current')
      const session = unwrap(result.data, result.error)
      csrf = session.csrfToken
      return session.profile
    }),
    logout: () => serialized(async () => {
      const result = await contract.POST('/iam/session/logout')
      complete(result.error)
      csrf = ''
    }),
    profile: () => serialized(async () => {
      const result = await contract.GET('/iam/account/profile')
      return unwrap(result.data, result.error) as AccountProfile
    }),
    updateProfile: (update: UpdateProfile) => serialized(async () => {
      const result = await contract.PATCH('/iam/account/profile', { body: update })
      return unwrap(result.data, result.error) as AccountProfile
    }),
    changePassword: (currentPassword: string, newPassword: string) => serialized(async () => {
      const result = await contract.PUT('/iam/account/password', { body: { currentPassword, newPassword } })
      complete(result.error)
      csrf = ''
    }),
  }
}

const problemCategory = (value: unknown) => typeof value === 'object' && value !== null && 'category' in value
  ? String((value as Problem).category ?? 'unavailable')
  : 'unavailable'
