export function safeRunnerDiagnostic(value: unknown): string
export function safeBrowserDiagnostic(value: unknown): string
export function browserDiagnostic(output: string): string
export function runnerFailureLine(value: unknown): string
export function administrationMountDiagnostic(state: {
  failure: string | null
  canUsersRead: boolean
  rows: number
  total: number
  loading: boolean
  alertText: string | null
}): string
