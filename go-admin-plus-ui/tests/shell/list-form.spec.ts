import { describe, expect, it, vi } from 'vitest'

import {
  createFormController,
  createListController,
  createRemovalController
} from '@go-admin/ui'

const deferred = <T>() => {
  let resolve!: (value: T) => void
  const promise = new Promise<T>(done => { resolve = done })
  return { promise, resolve }
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
})
