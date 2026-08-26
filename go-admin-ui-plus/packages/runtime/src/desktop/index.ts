import {
  createDesktopRuntime,
  createWebRuntime,
  type Runtime
} from '../index'

interface WailsGlobal {
  readonly go?: {
    readonly desktop?: {
      readonly Bridge?: {
        readonly Bootstrap?: () => Promise<unknown>
      }
    }
  }
}

const browserGlobal = (): WailsGlobal => globalThis as WailsGlobal

/** Selects Desktop only when the Wails bootstrap binding is actually present. */
export const createHostRuntime = (host: WailsGlobal = browserGlobal()): Runtime => {
  if (!host.go) return createWebRuntime()
  const bootstrap = host.go.desktop?.Bridge?.Bootstrap
  if (typeof bootstrap !== 'function') {
    throw new Error('desktop bootstrap binding is unavailable')
  }
  return createDesktopRuntime({ bootstrap: () => bootstrap() })
}
