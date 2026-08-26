const sensitiveAssignment = /((?:password|secret|sessionToken|csrfToken|credential|authorization)["']?\s*[:=]\s*)(?:"[^"\r\n]*"|'[^'\r\n]*'|[^,\s;}]+)/gi
const sensitiveHeader = /((?:cookie|set-cookie|x-csrf-token|authorization):\s*)[^\r\n]+/gi
const sessionCookie = /(__Host-go-admin-session=)[^;\s]+/gi
const URIUserInfo = /((?:postgres(?:ql)?|https?):\/\/)[^@\s/]+@/gi

export const redactDiagnostic = (value, secrets = []) => {
  let result = String(value ?? '')
  for (const secret of secrets) {
    if (secret) result = result.replaceAll(String(secret), '[redacted]')
  }
  return result
    .replace(URIUserInfo, '$1[redacted]@')
    .replace(sessionCookie, '$1[redacted]')
    .replace(sensitiveHeader, '$1[redacted]')
    .replace(sensitiveAssignment, '$1[redacted]')
}

export const createDiagnosticBuffer = (limit = 8_192, secrets = []) => {
  let buffered = ''
  return {
    append(value) {
      buffered = (buffered + redactDiagnostic(value, secrets)).slice(-limit)
    },
    text() { return buffered.trim() },
  }
}

export const browserExceptionDiagnostic = (details, secrets = []) => {
  const description = details?.exception?.description ?? details?.exception?.value ?? details?.text ?? 'browser exception without public detail'
  return redactDiagnostic(description, secrets).slice(0, 4_096)
}
