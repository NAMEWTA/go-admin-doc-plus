const sanitize = (value, fallback, maximum) => {
  const message = typeof value === 'string' ? value : value instanceof Error ? value.message : fallback
  return message
    .replace(/\b(?:https?|postgres(?:ql)?):\/\/\S+/gi, '[redacted-url]')
    .replace(/\bBearer\s+\S+/gi, 'Bearer [redacted]')
    .replace(/\b(password|token|secret|cookie|authorization|dsn)\b\s*[:=]\s*\S+/gi, '$1=[redacted]')
    .replace(/[^a-z0-9 .,:;_()[\]#@=?|-]+/gi, '?')
    .replace(/\s+/g, ' ')
    .trim()
    .slice(0, maximum) || fallback
}

export const safeRunnerDiagnostic = (value) => sanitize(value, 'unknown runner failure', 200)
export const safeBrowserDiagnostic = (value) => sanitize(value, 'unknown browser assertion', 160)

const failures = new Set(['relogin', 'forbidden', 'validation', 'conflict', 'unavailable'])
const alertCodes = new Map([
  ['Your session must be renewed.', 'relogin'],
  ['You do not have permission for that action.', 'forbidden'],
  ['Review the submitted values.', 'validation'],
  ['The resource changed or is protected.', 'conflict'],
  ['The administration service is unavailable.', 'unavailable'],
])

export const administrationMountDiagnostic = ({ failure, canUsersRead, rows, total, loading, alertText }) => {
  const failureCode = failures.has(failure) ? failure : 'none'
  const alertCode = alertText ? alertCodes.get(alertText) ?? 'unrecognized' : 'none'
  const count = (value) => Number.isSafeInteger(value) && value >= 0 ? value : -1
  return `administration page did not load failure=${failureCode} can-users-read=${Boolean(canUsersRead)} rows=${count(rows)} total=${count(total)} loading=${Boolean(loading)} alert=${alertCode}`
}

export const browserDiagnostic = (output) => {
  const match = output.match(/IAM_ADMIN_E2E_FAIL\|ASSERTION\|([^<\r\n]{1,200})/)
  return match ? safeRunnerDiagnostic(match[1]) : 'safe browser diagnostic unavailable'
}

export const runnerFailureLine = (value) => `IAM_ADMIN_E2E_RUN_FAIL|${safeRunnerDiagnostic(value)}`
