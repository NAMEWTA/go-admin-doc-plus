import assert from 'node:assert/strict'
import test from 'node:test'
import { validatePolicy } from './policy.mjs'

const contract = operation => ({
  openapi: '3.1.0',
  paths: { '/users/{id}': { delete: operation } }
})

test('allows ordinary API descriptions containing CRUD verbs', () => {
  assert.doesNotThrow(() => validatePolicy(contract({
    operationId: 'deleteUser',
    summary: 'Delete user',
    responses: { 204: { description: 'User deleted' } }
  })))
})

test('rejects SQL detail in a public response example', () => {
  assert.throws(() => validatePolicy(contract({
    operationId: 'deleteUser',
    responses: {
      500: {
        content: {
          'application/problem+json': {
            example: { detail: 'SELECT password FROM sys_user' }
          }
        }
      }
    }
  })), /sensitive internal detail/)
})

test('rejects public failures without Problem JSON', () => {
  assert.throws(() => validatePolicy(contract({
    operationId: 'deleteUser',
    responses: { 409: { content: { 'application/json': {} } } }
  })), /application\/problem\+json/)
})
