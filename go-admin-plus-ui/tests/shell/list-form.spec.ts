import { describe, expect, it, vi } from 'vitest'

import {
  createFormController,
  createListController,
  createRemovalController
} from '@go-admin-plus/ui'

const deferred = <T>() => {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((done, fail) => {
    resolve = done
    reject = fail
  })
  return { promise, reject, resolve }
}

describe('shared list interaction', () => {
  it('searches from page one and keeps the selected page when refreshing', async () => {
    const load = vi.fn(async ({ filters, page, pageSize }: {
      filters: { name: string }
      page: number
      pageSize: number
    }) => ({ rows: [{ id: `${filters.name}-${page}` }], total: pageSize }))
    const list = createListController({
      initialFilters: () => ({ name: '' }),
      load,
      pageSize: 20,
      rowKey: (row: { id: string }) => row.id
    })

    await list.setPage(3)
    await list.search({ name: 'Ada' })
    expect(load).toHaveBeenLastCalledWith({ filters: { name: 'Ada' }, page: 1, pageSize: 20 })

    await list.setPage(2)
    await list.refresh()
    expect(load).toHaveBeenLastCalledWith({ filters: { name: 'Ada' }, page: 2, pageSize: 20 })
  })

  it('ignores a stale successful request after the latest result is visible', async () => {
    const first = deferred<{ rows: Array<{ id: string }>, total: number }>()
    const second = deferred<{ rows: Array<{ id: string }>, total: number }>()
    const load = vi.fn()
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(second.promise)
    const list = createListController({
      initialFilters: () => ({ name: '' }),
      load,
      rowKey: (row: { id: string }) => row.id
    })

    const staleRequest = list.search({ name: 'old' })
    const latestRequest = list.search({ name: 'new' })
    second.resolve({ rows: [{ id: 'new' }], total: 1 })
    await latestRequest
    first.resolve({ rows: [{ id: 'old' }], total: 99 })
    await expect(staleRequest).resolves.toBeUndefined()

    expect(list.snapshot()).toMatchObject({
      filters: { name: 'new' },
      loading: false,
      rows: [{ id: 'new' }],
      total: 1
    })
  })

  it('silently discards a stale rejection without changing the latest state', async () => {
    const first = deferred<{ rows: Array<{ id: string }>, total: number }>()
    const second = deferred<{ rows: Array<{ id: string }>, total: number }>()
    const load = vi.fn()
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(second.promise)
    const list = createListController({
      initialFilters: () => ({ name: '' }),
      load,
      rowKey: (row: { id: string }) => row.id
    })

    const staleRequest = list.search({ name: 'old' })
    const latestRequest = list.search({ name: 'new' })
    second.resolve({ rows: [{ id: 'new' }], total: 1 })
    await latestRequest
    first.reject(new Error('stale request failed'))
    await expect(staleRequest).resolves.toBeUndefined()

    expect(list.snapshot()).toMatchObject({
      filters: { name: 'new' },
      loading: false,
      rows: [{ id: 'new' }],
      total: 1
    })
  })

  it('normalizes before transport and keeps the last successful state when validation cancels an older request', async () => {
    const pending = deferred<{ rows: Array<{ id: string }>, total: number }>()
    const load = vi.fn()
      .mockResolvedValueOnce({ rows: [{ id: 'stable' }], total: 1 })
      .mockReturnValueOnce(pending.promise)
    const list = createListController({
      initialFilters: () => ({ name: '' }),
      load,
      normalizeFilters: filters => ({ name: filters.name.trim() }),
      validate: request => {
        if (request.filters.name.length > 4) throw new Error('invalid filters')
      },
      rowKey: (row: { id: string }) => row.id
    })

    await list.search({ name: ' ok ' })
    list.select(list.snapshot().rows)
    const stale = list.search({ name: 'next' })
    await expect(list.search({ name: 'invalid' })).rejects.toThrow('invalid filters')
    expect(load).toHaveBeenCalledTimes(2)
    expect(list.snapshot()).toMatchObject({
      filters: { name: 'ok' },
      loading: false,
      rows: [{ id: 'stable' }],
      selectedKeys: ['stable'],
      total: 1
    })
    pending.resolve({ rows: [{ id: 'stale' }], total: 99 })
    await stale
    expect(list.snapshot()).toMatchObject({ filters: { name: 'ok' }, rows: [{ id: 'stable' }], total: 1 })
  })

  for (const staleOutcome of ['resolve', 'reject'] as const) {
    it(`clears loading after normalization throws and the stale request later ${staleOutcome}s`, async () => {
      const pending = deferred<{ rows: Array<{ id: string }>, total: number }>()
      const load = vi.fn()
        .mockResolvedValueOnce({ rows: [{ id: 'stable' }], total: 1 })
        .mockReturnValueOnce(pending.promise)
      const list = createListController({
        initialFilters: () => ({ name: '' }),
        load,
        normalizeFilters: filters => {
          if (filters.name === 'explode') throw new Error('normalization failed')
          return { name: filters.name.trim() }
        },
        rowKey: (row: { id: string }) => row.id
      })

      await list.search({ name: 'stable' })
      const stale = list.search({ name: 'pending' })
      await expect(list.search({ name: 'explode' })).rejects.toThrow('normalization failed')
      expect(list.snapshot()).toMatchObject({ filters: { name: 'stable' }, loading: false, rows: [{ id: 'stable' }], total: 1 })
      if (staleOutcome === 'resolve') pending.resolve({ rows: [{ id: 'stale' }], total: 99 })
      else pending.reject(new Error('stale request failed'))
      await expect(stale).resolves.toBeUndefined()
      expect(list.snapshot()).toMatchObject({ filters: { name: 'stable' }, loading: false, rows: [{ id: 'stable' }], total: 1 })
    })
  }
})

describe('shared form interaction', () => {
  it('does not send a command when validation fails', async () => {
    const submit = vi.fn()
    const form = createFormController({
      validate: async () => false,
      submit
    })

    await expect(form.run({ name: '' })).resolves.toBe('invalid')
    expect(submit).not.toHaveBeenCalled()
  })

  it('guards validation and submission as one in-flight operation', async () => {
    const gate = deferred<boolean>()
    const submit = vi.fn(async () => undefined)
    const form = createFormController({ validate: () => gate.promise, submit })

    const first = form.run({ name: 'Ada' })
    await expect(form.run({ name: 'Ada' })).resolves.toBe('busy')
    gate.resolve(true)
    await expect(first).resolves.toBe('submitted')
    expect(submit).toHaveBeenCalledTimes(1)
  })

  it('refreshes once after a successful command', async () => {
    const refreshed = vi.fn(async () => undefined)
    const form = createFormController({
      validate: async () => true,
      submit: async () => undefined,
      submitted: refreshed
    })

    await expect(form.run({ name: 'Ada' })).resolves.toBe('submitted')
    expect(refreshed).toHaveBeenCalledTimes(1)
  })

  it('reports refresh failure separately after writing exactly once', async () => {
    const submit = vi.fn(async () => undefined)
    const failed = vi.fn(async () => undefined)
    const form = createFormController({
      validate: async () => true,
      submit,
      submitted: async () => { throw new Error('refresh failed') },
      failed
    })

    await expect(form.run({ name: 'Ada' })).resolves.toBe('refresh-failed')
    expect(submit).toHaveBeenCalledTimes(1)
    expect(failed).not.toHaveBeenCalled()
  })
})

describe('shared destructive interaction', () => {
  it('confirms once, writes once, then refreshes and clears selection', async () => {
    const confirmation = deferred<boolean>()
    const confirm = vi.fn(() => confirmation.promise)
    const execute = vi.fn(async () => undefined)
    const refreshed = vi.fn(async () => undefined)
    const clearSelection = vi.fn()
    const removal = createRemovalController({ confirm, execute, refreshed, clearSelection })

    const first = removal.run(['user-1'])
    await expect(removal.run(['user-1'])).resolves.toBe('busy')
    confirmation.resolve(true)
    await expect(first).resolves.toBe('completed')

    expect(confirm).toHaveBeenCalledTimes(1)
    expect(execute).toHaveBeenCalledTimes(1)
    expect(refreshed).toHaveBeenCalledTimes(1)
    expect(clearSelection).toHaveBeenCalledTimes(1)
  })

  it('does not write or refresh when confirmation is cancelled', async () => {
    const execute = vi.fn()
    const refreshed = vi.fn()
    const removal = createRemovalController({
      confirm: async () => false,
      execute,
      refreshed,
      clearSelection: vi.fn()
    })

    await expect(removal.run(['user-1'])).resolves.toBe('cancelled')
    expect(execute).not.toHaveBeenCalled()
    expect(refreshed).not.toHaveBeenCalled()
  })

  it('reports post-write UI failure separately without repeating the write', async () => {
    const execute = vi.fn(async () => undefined)
    const refreshed = vi.fn(async () => { throw new Error('refresh failed') })
    const failed = vi.fn(async () => undefined)
    const removal = createRemovalController({
      confirm: async () => true,
      execute,
      refreshed,
      clearSelection: vi.fn(),
      failed
    })

    await expect(removal.run(['user-1'])).resolves.toBe('refresh-failed')
    expect(execute).toHaveBeenCalledTimes(1)
    expect(refreshed).toHaveBeenCalledTimes(1)
    expect(failed).not.toHaveBeenCalled()
  })

  it('still refreshes after selection cleanup fails following a successful write', async () => {
    const execute = vi.fn(async () => undefined)
    const refreshed = vi.fn(async () => undefined)
    const failed = vi.fn(async () => undefined)
    const removal = createRemovalController({
      confirm: async () => true,
      execute,
      refreshed,
      clearSelection: () => { throw new Error('selection cleanup failed') },
      failed
    })

    await expect(removal.run(['user-1'])).resolves.toBe('refresh-failed')
    expect(execute).toHaveBeenCalledTimes(1)
    expect(refreshed).toHaveBeenCalledTimes(1)
    expect(failed).not.toHaveBeenCalled()
  })
})
