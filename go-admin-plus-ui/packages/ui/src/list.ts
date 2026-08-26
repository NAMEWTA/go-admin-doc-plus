export interface SortState {
  readonly key: string
  readonly direction: 'ascending' | 'descending'
}

export interface ListRequest<TFilters> {
  readonly filters: TFilters
  readonly page: number
  readonly pageSize: number
  readonly sort?: SortState
}

export interface ListResult<TRow> {
  readonly rows: ReadonlyArray<TRow>
  readonly total: number
}

export interface ListSnapshot<TFilters, TRow, TKey> extends ListRequest<TFilters> {
  readonly rows: ReadonlyArray<TRow>
  readonly total: number
  readonly selectedKeys: ReadonlyArray<TKey>
  readonly loading: boolean
}

export interface ListController<TFilters, TRow, TKey> {
  snapshot(): ListSnapshot<TFilters, TRow, TKey>
  refresh(): Promise<void>
  search(filters: TFilters): Promise<void>
  reset(): Promise<void>
  setPage(page: number): Promise<void>
  setPageSize(pageSize: number): Promise<void>
  setSort(sort?: SortState): Promise<void>
  select(rows: ReadonlyArray<TRow>): void
  clearSelection(): void
}

export interface ListControllerOptions<TFilters, TRow, TKey> {
  readonly initialFilters: () => TFilters
  readonly load: (request: ListRequest<TFilters>) => Promise<ListResult<TRow>>
  readonly rowKey: (row: TRow) => TKey
  readonly pageSize?: number
}

const requirePositiveInteger = (value: number, name: string) => {
  if (!Number.isSafeInteger(value) || value < 1) throw new RangeError(`${name} must be a positive integer`)
}

/** Owns deterministic filters, paging, sorting and selection for a management list. */
export const createListController = <TFilters extends object, TRow, TKey>(
  options: ListControllerOptions<TFilters, TRow, TKey>
): ListController<TFilters, TRow, TKey> => {
  let filters = options.initialFilters()
  let page = 1
  let pageSize = options.pageSize ?? 20
  let sort: SortState | undefined
  let rows: ReadonlyArray<TRow> = []
  let total = 0
  let selectedKeys: ReadonlyArray<TKey> = []
  let loading = false
  let requestSequence = 0
  requirePositiveInteger(pageSize, 'pageSize')

  const snapshot = (): ListSnapshot<TFilters, TRow, TKey> => ({
    filters: { ...filters },
    page,
    pageSize,
    ...(sort ? { sort: { ...sort } } : {}),
    rows: [...rows],
    total,
    selectedKeys: [...selectedKeys],
    loading
  })

  const refresh = async () => {
    const sequence = ++requestSequence
    loading = true
    try {
      const result = await options.load({
        filters: { ...filters },
        page,
        pageSize,
        ...(sort ? { sort: { ...sort } } : {})
      })
      if (sequence !== requestSequence) return
      rows = [...result.rows]
      total = result.total
    } finally {
      if (sequence === requestSequence) loading = false
    }
  }

  return {
    snapshot,
    refresh,
    async search(nextFilters) {
      filters = { ...nextFilters }
      page = 1
      await refresh()
    },
    async reset() {
      filters = options.initialFilters()
      page = 1
      sort = undefined
      selectedKeys = []
      await refresh()
    },
    async setPage(nextPage) {
      requirePositiveInteger(nextPage, 'page')
      page = nextPage
      await refresh()
    },
    async setPageSize(nextPageSize) {
      requirePositiveInteger(nextPageSize, 'pageSize')
      pageSize = nextPageSize
      page = 1
      await refresh()
    },
    async setSort(nextSort) {
      sort = nextSort ? { ...nextSort } : undefined
      page = 1
      await refresh()
    },
    select(nextRows) {
      selectedKeys = nextRows.map(options.rowKey)
    },
    clearSelection() {
      selectedKeys = []
    }
  }
}
