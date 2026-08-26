import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    environment: 'node',
    include: [
      'tests/shell/**/*.spec.ts',
      'packages/domains/iam/src/session/**/*.spec.ts',
      'packages/web-domains/iam/src/session/**/*.spec.ts'
    ]
  }
})
