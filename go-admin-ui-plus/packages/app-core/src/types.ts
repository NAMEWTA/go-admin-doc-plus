import type { ApiEnvelope } from '../../contracts/src'

export type ApiResponse<T = unknown> = ApiEnvelope<T>

export interface PageResult<T> {
  list: T[]
  count: number
  pageIndex?: number
  pageSize?: number
}

export type Id = string | number

export interface PageQuery {
  pageIndex: number
  pageSize: number
}

export type ListApi<TRow, TQuery = Record<string, unknown>> = (
  query: TQuery & PageQuery
) => Promise<ApiResponse<PageResult<TRow>>>

export interface DictOption {
  label: string
  value: string
  [key: string]: unknown
}

export interface BackendMenu {
  path: string
  component: string
  routeKey?: string
  visible: string
  menuName: string
  title: string
  icon?: string
  noCache?: boolean
  children?: BackendMenu[] | null
}
