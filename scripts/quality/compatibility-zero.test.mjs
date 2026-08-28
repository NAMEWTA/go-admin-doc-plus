import assert from 'node:assert/strict'
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'
import { checkCompatibility } from './compatibility-zero.mjs'

test('detects removed paths and active compatibility references', () => {
  const root = mkdtempSync(join(tmpdir(), 'go-admin-compatibility-'))
  mkdirSync(join(root, 'go-admin-ui-plus'), { recursive: true })
  mkdirSync(join(root, 'docs'), { recursive: true })
  mkdirSync(join(root, '.agents/skills/stale-module'), { recursive: true })
  mkdirSync(join(root, 'go-admin-plus/internal/application'), { recursive: true })
  mkdirSync(join(root, 'go-admin-plus/internal/modules/demo'), { recursive: true })
  mkdirSync(join(root, 'go-admin-plus-ui/apps/admin-web/src'), { recursive: true })
  mkdirSync(join(root, 'bin'), { recursive: true })
  writeFileSync(join(root, 'README.md'), 'Use Redis and unsigned-self-use.\n')
  writeFileSync(join(root, '.agents/skills/stale-module/SKILL.md'), 'Connect the module to Casbin.\n')
  writeFileSync(join(root, 'go-admin-plus/internal/application/architecture_test.go'), 'const removedHost = "github.com/wailsapp/wails"\n')
  writeFileSync(join(root, 'go-admin-plus/internal/modules/demo/service.go'), 'import "gorm.io/gorm"\nconst tokenKind = "jwt"\n')
  writeFileSync(join(root, 'go-admin-plus-ui/apps/admin-web/src/legacy.rs'), 'const TOKEN: &str = "refresh_token";\n')
  writeFileSync(join(root, 'bin/sidecar'), Buffer.from('jwt\0compiled-binary'))
  const failures = checkCompatibility(root)
  assert.ok(failures.some(message => message.includes('removed path')))
  assert.ok(failures.some(message => message.includes('Redis')))
  assert.ok(failures.some(message => message.includes('old release class')))
  assert.ok(failures.some(message => message.includes('.agents/skills/stale-module/SKILL.md')))
  assert.ok(failures.some(message => message === 'GORM remains in go-admin-plus/internal/modules/demo/service.go'))
  assert.ok(failures.some(message => message === 'JWT remains in go-admin-plus/internal/modules/demo/service.go'))
  assert.ok(failures.some(message => message === 'refresh token remains in go-admin-plus-ui/apps/admin-web/src/legacy.rs'))
  assert.ok(!failures.some(message => message.includes('go-admin-plus/internal/application/architecture_test.go')))
  assert.ok(!failures.some(message => message.includes('bin/sidecar')))
})
