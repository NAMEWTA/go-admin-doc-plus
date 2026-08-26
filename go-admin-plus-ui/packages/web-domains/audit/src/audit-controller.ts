import { AuditRequestError, type AuditClient, type AuditFact, type AuditFailure, type AuditFilters, type CleanupResult } from '@go-admin/domain-audit'
import {
  createListController,
  createRemovalController,
  type ListController,
  type RemovalRunResult,
} from '@go-admin/ui'

export interface AuditController {
  readonly list: ListController<AuditFilters, AuditFact, string>
  detail(id: string): Promise<AuditFact>
  cleanup(before: string): Promise<RemovalRunResult>
  lastCleanup(): CleanupResult | null
  lastFailure(): AuditFailure | null
}

export const createAuditController = (
  client: AuditClient,
  confirmCleanup: (count: number) => Promise<boolean>,
): AuditController => {
  const list = createListController<AuditFilters, AuditFact, string>({
    initialFilters: () => ({}),
    load: async (request) => {
      const page = await client.list({ filters: request.filters, page: request.page, pageSize: request.pageSize })
      return { rows: page.records, total: page.total }
    },
    rowKey: (fact) => fact.id,
    pageSize: 20,
  })
  let cleanupResult: CleanupResult | null = null
	let cleanupFailure: AuditFailure | null = null
  const removal = createRemovalController<string>({
    confirm: confirmCleanup,
    execute: async ([before]) => {
      if (!before) throw new Error('Cleanup boundary is required')
			try {
				cleanupResult = await client.cleanup(before)
			} catch (error) {
				cleanupFailure = error instanceof AuditRequestError ? error.category : 'unavailable'
				throw error
			}
    },
    clearSelection: list.clearSelection,
    refreshed: list.refresh,
  })
	let cleanupNeedsRefresh = false
  return {
    list,
    detail: (id) => client.detail(id),
	async cleanup(before) {
		if (cleanupNeedsRefresh) {
			cleanupFailure = null
			try {
				await list.refresh()
				cleanupNeedsRefresh = false
				return 'completed'
			} catch {
				return 'refresh-failed'
			}
		}
		cleanupResult = null
		cleanupFailure = null
		const result = await removal.run([before])
		if (result === 'refresh-failed') cleanupNeedsRefresh = true
		return result
	},
    lastCleanup: () => cleanupResult,
		lastFailure: () => cleanupFailure,
  }
}
