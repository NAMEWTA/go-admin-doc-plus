#!/usr/bin/env node

import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { checkCompatibility } from '../../scripts/quality/compatibility-zero.mjs'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../..')
const failures = checkCompatibility(root)
if (failures.length) {
  console.error(`COMPATIBILITY_ZERO_FAIL\n${failures.join('\n')}`)
  process.exit(1)
}
console.log('COMPATIBILITY_ZERO_PASS')
