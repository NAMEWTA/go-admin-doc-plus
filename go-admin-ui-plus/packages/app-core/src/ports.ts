export interface AppHostPorts {
  readonly refreshSystemConfig: () => Promise<void> | void
  readonly closeCurrentView: (nextPath?: string) => Promise<void> | void
  readonly getSessionToken: () => string
  readonly getApiBaseUrl: () => string
  readonly listIcons: () => ReadonlyArray<string>
}

let configuredPorts: AppHostPorts | undefined

export const configureAppHostPorts = (ports: AppHostPorts) => {
  if (!ports || Object.values(ports).some(port => typeof port !== 'function')) {
    throw new Error('App Host ports are incomplete')
  }
  configuredPorts = ports
}

const ports = () => {
  if (!configuredPorts) throw new Error('App Host ports have not been configured')
  return configuredPorts
}

export const refreshSystemConfig = () => ports().refreshSystemConfig()
export const closeCurrentView = (nextPath?: string) => ports().closeCurrentView(nextPath)
export const getSessionToken = () => ports().getSessionToken()
export const getApiBaseUrl = () => ports().getApiBaseUrl()
export const listIcons = () => ports().listIcons()

export type DictionaryLoader = (type: string) => Promise<ReadonlyArray<import('./types').DictOption>>

let dictionaryLoader: DictionaryLoader | undefined

export const configureDictionaryLoader = (loader: DictionaryLoader) => {
  if (typeof loader !== 'function') throw new Error('Dictionary loader is required')
  dictionaryLoader = loader
}

export const loadDictionary = (type: string) => {
  if (!dictionaryLoader) throw new Error('Dictionary loader has not been configured by the App Shell')
  return dictionaryLoader(type)
}

let successNotifier: ((message: string) => void) | undefined

export const configureSuccessNotifier = (notifier: (message: string) => void) => {
  if (typeof notifier !== 'function') throw new Error('Success notifier is required')
  successNotifier = notifier
}

export const notifySuccess = (message: string) => {
  if (!successNotifier) throw new Error('Success notifier has not been configured by the App Shell')
  successNotifier(message)
}
