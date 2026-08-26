import { createAuditController, createWebAuditClient } from '@go-admin/web-domain-audit'

const baseURL = process.env.GO_ADMIN_AUDIT_E2E_BASE_URL
if (!baseURL) throw new Error('Audit E2E base URL is required')

const assert: (condition: unknown, message: string) => asserts condition = (condition, message) => {
  if (!condition) throw new Error(message)
}

const snapshot = async (): Promise<{ count: number }> => {
  const response = await fetch(new URL('/__test/snapshot', baseURL))
  assert(response.ok, 'Audit E2E snapshot failed')
  return response.json() as Promise<{ count: number }>
}

const shutdown = async () => {
  await fetch(new URL('/__test/shutdown', baseURL), { method: 'POST' })
}

const run = async () => {
  const client = createWebAuditClient(fetch, new URL('/api', baseURL).toString())
  const denied = createAuditController(client, async () => false)
  assert(await denied.cleanup('2026-06-01T00:00:00Z') === 'cancelled', 'Audit E2E confirmation rejection failed')
  assert((await snapshot()).count === 1, 'Audit E2E rejected cleanup changed state')

  const controller = createAuditController(client, async () => true)
  await controller.list.search({ source: 'web', action: 'update' })
  const listed = controller.list.snapshot()
  assert(listed.total === 1 && listed.rows[0]?.subject === 'demo:ui-record-revision-2', 'Audit E2E filtered list failed')
  const detail = await controller.detail(listed.rows[0].id)
  assert(detail.action === 'update' && detail.source === 'web', 'Audit E2E detail failed')
  assert(!JSON.stringify(detail).includes('payload') && !JSON.stringify(detail).includes('businessKey'), 'Audit E2E detail leaked envelope')

  assert(await controller.cleanup('2026-06-01T00:00:00Z') === 'completed', 'Audit E2E cleanup failed')
  assert(controller.lastCleanup()?.deleted === 1, 'Audit E2E cleanup result failed')
  assert(controller.list.snapshot().total === 0, 'Audit E2E cleanup did not refresh list')
  assert((await snapshot()).count === 0, 'Audit E2E cleanup did not persist')
}

try {
  await run()
  console.log('AUDIT_E2E_PROFILE_PASS')
} finally {
  await shutdown().catch(() => undefined)
}
