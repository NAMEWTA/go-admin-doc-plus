import { createApp, h, type Component } from 'vue'
import { GeneratorRequestError, generatorPermissions } from '@go-admin/domain-generator'
import { createGeneratorController, createWebGeneratorClient, GeneratorWizardPage } from '@go-admin/web-domain-generator'
import { createCapabilityController } from '@go-admin/domain-iam/administration'
import { createSessionController } from '@go-admin/domain-iam/session'
import { createWebAdministrationClient } from '@go-admin/web-domain-iam/administration'
import { createWebSessionClient } from '@go-admin/web-domain-iam/session'

const result = document.querySelector<HTMLElement>('#result')!
const session = createSessionController(createWebSessionClient(fetch, '/api'))
const capabilities = createCapabilityController(createWebAdministrationClient(fetch, '/api'))
const client = createWebGeneratorClient(fetch, '/api')
const controller = createGeneratorController(client, { can: permission => capabilities.can(permission) })
const wait = async (condition: () => boolean, message: string) => { const deadline = Date.now()+30_000; while(Date.now()<deadline){ if(condition()) return; await new Promise(resolve => setTimeout(resolve, 25)) }; throw new Error(message) }

try {
  await session.login({ username: 'admin', password: 'administrator password' }); await capabilities.refresh()
  for (const permission of Object.values(generatorPermissions)) if (!capabilities.can(permission)) throw new Error('generator capability is missing')
  createApp({ render: () => h(GeneratorWizardPage as Component, { controller }) }).mount('#app')
  await controller.tables.refresh(); await wait(() => controller.projectionVisible, 'metadata projection did not load')
  const tables = controller.tables.snapshot().rows
  if (tables.some(table => table.schema === 'information_schema' || table.schema === 'sqlite_master' || table.schema.startsWith('pg_'))) throw new Error('system schema leaked')
  const products = tables.find(table => table.name === 'products'); if (!products) throw new Error('allowlisted products table is missing')
  await controller.select(products); controller.setNames('../catalog', 'Product', 'products')
  if (await controller.createPreview() !== 'invalid' || controller.failure() !== 'validation') throw new Error('invalid identifier was accepted')
  controller.setNames('catalog', 'Product', 'products')
  const draft = controller.draft!
  try { await client.preview({ ...draft, table: { schema: '../main', name: 'products' } }); throw new Error('invalid table was accepted') } catch (error) { if (!(error instanceof GeneratorRequestError) || error.category !== 'validation') throw error }
  try { await client.preview({ ...draft, columns: draft.columns.map((column,index) => index === 0 ? { ...column, field: '../ID' } : column) }); throw new Error('invalid field was accepted') } catch (error) { if (!(error instanceof GeneratorRequestError) || error.category !== 'validation') throw error }
  const gateFailure = await fetch('/__test/gate-failure', { method: 'POST' }); if (gateFailure.status !== 204) throw new Error('gate failure probe failed')
  const before = await (await fetch('/__test/output')).json() as { entries: number; failedEntries: number }
  if (before.entries !== 0 || before.failedEntries !== 0) throw new Error('rejected generation left output')
  if (await controller.createPreview() !== 'completed' || !controller.previewValue?.files.some(file => file.path.endsWith('/service.go'))) throw new Error('preview did not produce CRUD output')
  const expected = await fetch('/__test/expected', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ directory: `catalog-${controller.previewValue.digest.slice(0,12)}`, files: controller.previewValue.files.map(file => ({ path: file.path, sha256: file.sha256 })) }) }); if (expected.status !== 204) throw new Error('preview hash registration failed')
  if (await controller.confirmWrite(false) !== 'invalid') throw new Error('unconfirmed write was accepted')
  if (await controller.confirmWrite(true) !== 'completed' || controller.result?.directory.startsWith('catalog-') !== true) throw new Error('confirmed write did not complete')
  const after = await (await fetch('/__test/output')).json() as { entries: number; failedEntries: number }
  if (after.entries !== 1 || after.failedEntries !== 0) throw new Error('published output count is invalid')
  result.textContent = 'GENERATOR_E2E_PASS'
} catch (error) {
  const message = error instanceof Error && /^[a-zA-Z0-9 .,'-]{1,160}$/.test(error.message) ? error.message : 'browser assertion failed'
  result.textContent = `GENERATOR_E2E_FAIL|${message}`
}
