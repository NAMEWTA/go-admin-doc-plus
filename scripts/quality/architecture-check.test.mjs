import assert from 'node:assert/strict'
import { test } from 'node:test'
import { checkArchitecture } from './architecture-check.mjs'

test('current repository satisfies canonical architecture', () => {
  assert.deepEqual(checkArchitecture(new URL('../..', import.meta.url).pathname), [])
})
