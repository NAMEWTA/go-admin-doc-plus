import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    environment: 'node',
    include: [
      'tests/shell/**/*.spec.ts',
      'packages/domains/iam/src/session/**/*.spec.ts',
      'packages/web-domains/iam/src/session/**/*.spec.ts',
      'packages/domains/demo/src/**/*.spec.ts',
      'packages/web-domains/demo/src/**/*.spec.ts',
      'packages/domains/organization/src/**/*.spec.ts',
      'packages/web-domains/organization/src/**/*.spec.ts'
    ]
  }
})
