export type PermissionCode = `${string}:${string}` | `${string}:${string}:${string}`

export type RuntimeIdentity =
  | { readonly kind: 'unauthenticated' }
  | {
      readonly kind: 'authenticated'
      readonly subjectId: string
      readonly permissions: ReadonlyArray<PermissionCode>
    }

export interface NavigationEntry {
  readonly path: `/${string}`
  readonly permission?: PermissionCode
}

export interface RuntimeRequest {
  readonly signal?: AbortSignal
}

export interface ShellRuntimePort {
  /** Returns identity facts only; session material remains inside the adapter. */
  loadIdentity(request?: RuntimeRequest): Promise<RuntimeIdentity>
  loadNavigation(request?: RuntimeRequest): Promise<ReadonlyArray<NavigationEntry>>
}

export type HostCapability = 'clipboard-write' | 'file-open' | 'file-save' | 'notification'

export interface PlatformPort {
  readonly runtime: 'web' | 'desktop'
  /** Callers must check capability presence before invoking an optional host operation. */
  listCapabilities(): ReadonlySet<HostCapability>
  notify(message: string): Promise<void>
  writeClipboard(text: string): Promise<void>
}
