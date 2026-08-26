export const desktopLaunchTokenHeader = 'X-Go-Admin-Launch-Token'

export interface LaunchToken {
  readonly header: typeof desktopLaunchTokenHeader
  readonly value: string
}

export interface RuntimeConfig {
  readonly kind: 'web' | 'desktop'
  readonly apiBaseUrl: string
  readonly requestTimeoutMs: number
  readonly launchToken?: LaunchToken
}

export interface Runtime {
  bootstrap(): Promise<RuntimeConfig>
}

export interface DesktopBootstrapProvider {
  bootstrap(): Promise<unknown>
}

const defaultRequestTimeoutMs = 10_000

const requireObject = (value: unknown): Record<string, unknown> => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error('runtime bootstrap must return an object')
  }
  return value as Record<string, unknown>
}

const requireTimeout = (value: unknown): number => {
  if (value === undefined) return defaultRequestTimeoutMs
  if (!Number.isInteger(value) || Number(value) <= 0) {
    throw new Error('runtime requestTimeoutMs must be a positive integer')
  }
  return Number(value)
}

const normalizeDesktopBaseUrl = (value: unknown): string => {
  if (typeof value !== 'string' || value.trim() === '') {
    throw new Error('desktop runtime apiBaseUrl is required')
  }
  let parsed: URL
  try {
    parsed = new URL(value)
  } catch {
    throw new Error('desktop runtime apiBaseUrl is invalid')
  }
  if (parsed.protocol !== 'http:' || !['127.0.0.1', '[::1]'].includes(parsed.hostname) || parsed.port === '') {
    throw new Error('desktop runtime apiBaseUrl must use loopback HTTP with an explicit port')
  }
  if (parsed.username || parsed.password || parsed.search || parsed.hash || (parsed.pathname !== '' && parsed.pathname !== '/')) {
    throw new Error('desktop runtime apiBaseUrl must contain only a loopback origin')
  }
  return parsed.origin
}

export const validateDesktopBootstrap = (value: unknown): RuntimeConfig => {
  const payload = requireObject(value)
  const launchToken = requireObject(payload.launchToken)
  if (launchToken.header !== desktopLaunchTokenHeader) {
    throw new Error(`desktop runtime launch token header must be ${desktopLaunchTokenHeader}`)
  }
  if (typeof launchToken.value !== 'string' || launchToken.value.trim() === '') {
    throw new Error('desktop runtime launch token is required')
  }
  return Object.freeze({
    kind: 'desktop' as const,
    apiBaseUrl: normalizeDesktopBaseUrl(payload.apiBaseUrl),
    requestTimeoutMs: requireTimeout(payload.requestTimeoutMs),
    launchToken: Object.freeze({
      header: desktopLaunchTokenHeader,
      value: launchToken.value
    })
  })
}

export const createWebRuntime = (): Runtime => ({
  async bootstrap() {
    return Object.freeze({
      kind: 'web' as const,
      apiBaseUrl: '',
      requestTimeoutMs: defaultRequestTimeoutMs
    })
  }
})

export const createDesktopRuntime = (provider: DesktopBootstrapProvider): Runtime => {
  if (!provider || typeof provider.bootstrap !== 'function') {
    throw new Error('desktop bootstrap provider is required')
  }
  return {
    async bootstrap() {
      return validateDesktopBootstrap(await provider.bootstrap())
    }
  }
}
