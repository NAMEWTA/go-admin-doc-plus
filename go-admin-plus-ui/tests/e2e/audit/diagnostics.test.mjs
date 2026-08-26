import assert from 'node:assert/strict'
import test from 'node:test'
import { browserExceptionDiagnostic, createDiagnosticBuffer, redactDiagnostic } from './diagnostics.mjs'

test('diagnostics retain stable failures while redacting credential material', () => {
  const dsn = 'postgres://private-user:private-password@127.0.0.1:55439/postgres?sslmode=disable'
  const diagnostic = redactDiagnostic([
    'Error: filtered Audit list failed',
    dsn,
    'Set-Cookie: __Host-go-admin-session=raw-session-value; Secure',
    'X-CSRF-Token: raw-csrf-value',
    'password="raw password value"',
  ].join('\n'), [dsn, 'known-fixture-password'])

  assert.match(diagnostic, /filtered Audit list failed/)
  for (const secret of ['private-user', 'private-password', 'raw-session-value', 'raw-csrf-value', 'raw password value']) {
    assert.doesNotMatch(diagnostic, new RegExp(secret))
  }
})

test('diagnostic buffers are bounded and browser exception descriptions survive', () => {
  const buffer = createDiagnosticBuffer(32)
  buffer.append('x'.repeat(64))
  assert.equal(buffer.text().length, 32)
  assert.match(browserExceptionDiagnostic({ exception: { description: 'Error: cleanup count mismatch' } }), /cleanup count mismatch/)
})
