export type { components, operations, paths } from './generated'

export interface ApiEnvelope<T = unknown> {
  readonly code: number
  readonly data: T
  readonly msg: string
}

export type RuntimeCapabilities =
  import('./generated').components['schemas']['RuntimeCapabilities']

export type OperationalStatus =
  import('./generated').components['schemas']['OperationalStatus']
