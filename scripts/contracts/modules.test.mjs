import assert from 'node:assert/strict'
import { resolve } from 'node:path'
import test from 'node:test'
import { isManagedGeneratedOutput, parseManagedModuleOutput, resolveModuleMetadata } from './modules.mjs'

const repositoryRoot = resolve('/workspace/product')

test('resolves module transport outputs inside owner roots', () => {
  const metadata = resolveModuleMetadata(repositoryRoot, {
    'x-go-admin-module': 'demo',
    'x-go-admin-codegen': {
      owner: 'demo',
      goPackage: 'demotransport',
      goOutput: 'go-admin-plus/internal/modules/demo/transport/openapi.gen.go',
      typescriptOutput: 'go-admin-plus-ui/packages/domains/demo/src/generated'
    }
  }, 'demo.yaml')

  assert.equal(metadata.id, 'demo')
  assert.equal(metadata.goPackage, 'demotransport')
  assert.equal(metadata.goOutput, resolve(repositoryRoot, 'go-admin-plus/internal/modules/demo/transport/openapi.gen.go'))
  assert.equal(metadata.owner, 'demo')
  assert.equal(metadata.typescriptOutput, resolve(repositoryRoot, 'go-admin-plus-ui/packages/domains/demo/src/generated'))
})

for (const [name, document] of [
  ['invalid module id', {
    'x-go-admin-module': '../demo',
    'x-go-admin-codegen': { owner: 'demo', goPackage: 'demo', goOutput: 'go-admin-plus/internal/modules/demo/transport/openapi.gen.go', typescriptOutput: 'go-admin-plus-ui/packages/domains/demo/src/generated' }
  }],
  ['Go output traversal', {
    'x-go-admin-module': 'demo',
    'x-go-admin-codegen': { owner: 'demo', goPackage: 'demo', goOutput: '../outside.go', typescriptOutput: 'go-admin-plus-ui/packages/domains/demo/src/generated' }
  }],
  ['TypeScript output traversal', {
    'x-go-admin-module': 'demo',
    'x-go-admin-codegen': { owner: 'demo', goPackage: 'demo', goOutput: 'go-admin-plus/internal/modules/demo/transport/openapi.gen.go', typescriptOutput: '/tmp/generated' }
  }],
  ['cross-module output', {
    'x-go-admin-module': 'demo',
    'x-go-admin-codegen': { owner: 'demo', goPackage: 'demo', goOutput: 'go-admin-plus/internal/modules/audit/transport/openapi.gen.go', typescriptOutput: 'go-admin-plus-ui/packages/domains/demo/src/generated' }
  }],
  ['Go output missing the owner transport directory', {
    'x-go-admin-module': 'transport-fragment',
    'x-go-admin-codegen': { owner: 'transport', goPackage: 'transport', goOutput: 'go-admin-plus/internal/modules/transport/openapi.gen.go', typescriptOutput: 'go-admin-plus-ui/packages/domains/transport/src/generated' }
  }],
  ['unknown codegen metadata', {
    'x-go-admin-module': 'demo',
    'x-go-admin-codegen': { owner: 'demo', goPackage: 'demo', goOutput: 'go-admin-plus/internal/modules/demo/transport/openapi.gen.go', typescriptOutput: 'go-admin-plus-ui/packages/domains/demo/src/generated', unsupported: true }
  }]
]) {
  test(`rejects ${name}`, () => {
    assert.throws(() => resolveModuleMetadata(repositoryRoot, document, 'fixture.yaml'), /module|output|package/i)
  })
}

test('allows multiple fragments to write inside one explicit module owner', () => {
  const metadata = resolveModuleMetadata(repositoryRoot, {
    'x-go-admin-module': 'iam-session',
    'x-go-admin-codegen': {
      owner: 'iam',
      goPackage: 'sessiontransport',
      goOutput: 'go-admin-plus/internal/modules/iam/session/transport/openapi.gen.go',
      typescriptOutput: 'go-admin-plus-ui/packages/domains/iam/src/session/generated'
    }
  }, 'iam-session.yaml')

  assert.equal(metadata.id, 'iam-session')
  assert.equal(metadata.owner, 'iam')
})

test('uses one path grammar for nested generation targets and manifest entries', () => {
  const metadata = resolveModuleMetadata(repositoryRoot, {
    'x-go-admin-module': 'iam-session',
    'x-go-admin-codegen': {
      owner: 'iam',
      goPackage: 'sessionv2transport',
      goOutput: 'go-admin-plus/internal/modules/iam/session_v2/transport/openapi.gen.go',
      typescriptOutput: 'go-admin-plus-ui/packages/domains/iam/src/session_v2/generated'
    }
  }, 'iam-session.yaml')

  assert.equal(parseManagedModuleOutput('go-admin-plus/internal/modules/iam/session_v2/transport/openapi.gen.go')?.owner, 'iam')
  assert.equal(parseManagedModuleOutput('go-admin-plus-ui/packages/domains/iam/src/session_v2/generated')?.kind, 'typescript-directory')
  assert.equal(isManagedGeneratedOutput('go-admin-plus-ui/packages/domains/iam/src/session_v2/generated/client.ts'), true)
  assert.equal(isManagedGeneratedOutput('go-admin-plus-ui/packages/domains/iam/manual/generated/client.ts'), false)
  assert.equal(isManagedGeneratedOutput('go-admin-plus/internal/modules/transport/openapi.gen.go'), false)
  assert.match(metadata.goOutput, /session_v2/)
})

test('rejects mismatched nested Go and TypeScript slice paths', () => {
  assert.throws(() => resolveModuleMetadata(repositoryRoot, {
    'x-go-admin-module': 'iam-session',
    'x-go-admin-codegen': {
      owner: 'iam',
      goPackage: 'iamsessiontransport',
      goOutput: 'go-admin-plus/internal/modules/iam/session/transport/openapi.gen.go',
      typescriptOutput: 'go-admin-plus-ui/packages/domains/iam/src/administration/generated'
    }
  }), /same nested module path/)
})
