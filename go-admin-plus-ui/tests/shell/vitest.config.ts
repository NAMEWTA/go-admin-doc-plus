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
      'packages/web-domains/organization/src/**/*.spec.ts',
      'packages/domains/generator/src/**/*.spec.ts',
      'packages/web-domains/generator/src/**/*.spec.ts',
      'packages/domains/settings/src/**/*.spec.ts',
      'packages/web-domains/settings/src/**/*.spec.ts',
      'packages/domains/scheduler/src/**/*.spec.ts',
      'packages/web-domains/scheduler/src/**/*.spec.ts',
      'packages/domains/files/src/**/*.spec.ts',
      'packages/web-domains/files/src/**/*.spec.ts',
      'packages/adapters/browser/src/files.spec.ts',
      'packages/adapters/desktop/src/**/*.spec.ts'
    ]
  }
})
