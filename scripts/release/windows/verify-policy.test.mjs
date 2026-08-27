import test from 'node:test'
import assert from 'node:assert/strict'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { verifyIdentity, verifyRepository, verifyWorkflow } from './verify-policy.mjs'

const repository = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../..')

test('repository Windows release policy is complete', async () => {
  await verifyRepository(repository)
})

test('identity rejects unsigned, non-x64, or destructive uninstall releases', async () => {
  const identity = JSON.parse(await import('node:fs/promises').then(fs => fs.readFile(path.join(repository, 'release/windows/identity.json'), 'utf8')))
  for (const change of [
    { signingRequired: false },
    { architecture: 'arm64' },
    { releaseClass: 'unsigned-self-use' },
    { uninstallPolicy: { ...identity.uninstallPolicy, preserveAppData: false } }
  ]) assert.throws(() => verifyIdentity({ ...identity, ...change }))
})

test('workflow rejects Wails, self-use, and remote publishing regressions', async () => {
  const workflow = await import('node:fs/promises').then(fs => fs.readFile(path.join(repository, '.github/workflows/release-windows.yml'), 'utf8'))
  for (const regression of ['wails build', 'unsigned-self-use', 'gh release create']) {
    assert.throws(() => verifyWorkflow(`${workflow}\n${regression}\n`))
  }
})
