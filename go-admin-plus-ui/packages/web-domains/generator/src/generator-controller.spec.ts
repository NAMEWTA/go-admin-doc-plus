import { describe, expect, it, vi } from 'vitest'
import { GeneratorRequestError, generatorPermissions, type GenerationPreview, type GeneratorClient, type TableMetadata } from '@go-admin-plus/domain-generator'
import { createGeneratorController } from './generator-controller'
import pageSource from './GeneratorWizardPage.vue?raw'

const table: TableMetadata = { table: { schema: 'main', name: 'products' }, columns: [
  { name: 'id', databaseType: 'TEXT', kind: 'uuid', nullable: false, primaryKey: true, ordinal: 1 },
  { name: 'name', databaseType: 'TEXT', kind: 'string', nullable: false, primaryKey: false, ordinal: 2 },
] }
const preview: GenerationPreview = { token: 'a'.repeat(64), digest: 'b'.repeat(64), module: 'product', createdAt: '2026-08-27T00:00:00Z', expiresAt: '2026-08-27T00:05:00Z', files: [{ path: 'module.go', content: 'package module\n', sha256: 'c'.repeat(64) }] }
const fixture = () => {
  const client: GeneratorClient = { getConfig: vi.fn(async () => { throw new GeneratorRequestError('not-found') }), listTables: vi.fn(async () => [table.table]), describe: vi.fn(async () => table), preview: vi.fn(async () => preview), write: vi.fn(async token => ({ token, directory: 'product-bbbbbbbbbbbb', files: ['module.go'] })) }
  return { client, controller: createGeneratorController(client, { can: () => true }) }
}

describe('generator controller', () => {
  it('uses shared list state and completes source/configure/preview/confirmed write', async () => {
    const { client, controller } = fixture()
    await controller.tables.refresh(); await controller.select(table.table)
    expect(controller.step).toBe('configure')
    expect(await controller.createPreview()).toBe('completed')
    expect(controller.step).toBe('preview')
    expect(await controller.confirmWrite(false)).toBe('invalid')
    expect(client.write).not.toHaveBeenCalled()
    expect(await controller.confirmWrite(true)).toBe('completed')
    expect(controller.step).toBe('complete')
    expect(client.write).toHaveBeenCalledTimes(1)
    expect(pageSource).toContain('我已检查隔离目录中的生成结果')
  })
  it('fails closed when capability is withdrawn', async () => {
    const granted = new Set<string>(Object.values(generatorPermissions))
    const { client } = fixture(); const controller = createGeneratorController(client, { can: value => granted.has(value) })
    await controller.tables.refresh(); granted.delete(generatorPermissions.metadata)
    await controller.select(table.table)
    expect(client.describe).not.toHaveBeenCalled(); expect(controller.projectionVisible).toBe(false); expect(controller.failure()).toBe('forbidden')
  })
  it('keeps a gate-failed preview visible and never retries a write implicitly', async () => {
    const { client, controller } = fixture(); await controller.tables.refresh(); await controller.select(table.table); await controller.createPreview()
    vi.mocked(client.write).mockRejectedValue(new GeneratorRequestError('gate', 'trace_gate_123'))
    expect(await controller.confirmWrite(true)).toBe('failed')
    expect(controller.previewValue).toEqual(preview); expect(controller.step).toBe('preview'); expect(client.write).toHaveBeenCalledTimes(1)
    expect(controller.failureTraceId()).toBe('trace_gate_123')
    expect(pageSource).toContain('必需门禁未通过')
    controller.returnToConfiguration()
    expect(controller.step).toBe('configure'); expect(controller.previewValue).toBeNull(); expect(controller.draft).not.toBeNull()
    expect(controller.failure()).toBeNull(); expect(controller.failureTraceId()).toBeNull()
  })
})
