import type { ShellRuntimePort } from '@go-admin/platform'

export type ShellState =
  | { readonly kind: 'authenticated', readonly path: string, readonly subjectId: string }
  | { readonly kind: 'unauthenticated', readonly redirectTo: '/login' }
  | { readonly kind: 'unauthorized', readonly path: string }
  | { readonly kind: 'not-found', readonly path: string }
  | { readonly kind: 'adapter-failed', readonly retryable: true }

/** Resolves navigation into a display-safe state and never exposes adapter errors. */
export const resolveShellState = async (
  runtime: ShellRuntimePort,
  path: string
): Promise<ShellState> => {
  try {
    const identity = await runtime.loadIdentity()
    if (identity.kind === 'unauthenticated') {
      return { kind: 'unauthenticated', redirectTo: '/login' }
    }

    const navigation = await runtime.loadNavigation()
    const route = navigation.find(entry => entry.path === path)
    if (!route) return { kind: 'not-found', path }
    if (route.permission && !identity.permissions.includes(route.permission)) {
      return { kind: 'unauthorized', path }
    }

    return { kind: 'authenticated', path, subjectId: identity.subjectId }
  } catch {
    return { kind: 'adapter-failed', retryable: true }
  }
}
