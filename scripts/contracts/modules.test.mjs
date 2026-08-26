import assert from 'node:assert/strict'
import { join } from 'node:path'
import test from 'node:test'
import { resolveModuleMetadata } from './modules.mjs'

const repositoryRoot = '/workspace/product'

test('resolves module transport outputs inside owner roots', () => {
  const metadata = resolveModuleMetadata(repositoryRoot, {
    'x-go-admin-module': 'demo',
    'x-go-admin-codegen': {
      goPackage: 'demotransport',
      goOutput: 'go-admin-plus/internal/modules/demo/transport/openapi.gen.go',
      typescriptOutput: 'go-admin-ui-plus/packages/domains/demo/src/generated'
    }
  }, 'demo.yaml')

  assert.equal(metadata.id, 'demo')
  assert.equal(metadata.goPackage, 'demotransport')
  assert.equal(metadata.goOutput, join(repositoryRoot, 'go-admin-plus/internal/modules/demo/transport/openapi.gen.go'))
  assert.equal(metadata.typescriptOutput, join(repositoryRoot, 'go-admin-ui-plus/packages/domains/demo/src/generated'))
})

for (const [name, document] of [
  ['invalid module id', {
    'x-go-admin-module': '../demo',
    'x-go-admin-codegen': { goPackage: 'demo', goOutput: 'go-admin-plus/internal/modules/demo/transport/openapi.gen.go', typescriptOutput: 'go-admin-ui-plus/packages/domains/demo/src/generated' }
  }],
  ['Go output traversal', {
    'x-go-admin-module': 'demo',
    'x-go-admin-codegen': { goPackage: 'demo', goOutput: '../outside.go', typescriptOutput: 'go-admin-ui-plus/packages/domains/demo/src/generated' }
  }],
  ['TypeScript output traversal', {
    'x-go-admin-module': 'demo',
    'x-go-admin-codegen': { goPackage: 'demo', goOutput: 'go-admin-plus/internal/modules/demo/transport/openapi.gen.go', typescriptOutput: '/tmp/generated' }
  }],
  ['cross-module output', {
    'x-go-admin-module': 'demo',
    'x-go-admin-codegen': { goPackage: 'demo', goOutput: 'go-admin-plus/internal/modules/settings/transport/openapi.gen.go', typescriptOutput: 'go-admin-ui-plus/packages/domains/demo/src/generated' }
  }],
  ['unknown codegen metadata', {
    'x-go-admin-module': 'demo',
    'x-go-admin-codegen': { goPackage: 'demo', goOutput: 'go-admin-plus/internal/modules/demo/transport/openapi.gen.go', typescriptOutput: 'go-admin-ui-plus/packages/domains/demo/src/generated', unsupported: true }
  }]
]) {
  test(`rejects ${name}`, () => {
    assert.throws(() => resolveModuleMetadata(repositoryRoot, document, 'fixture.yaml'), /module|output|package/i)
  })
}
