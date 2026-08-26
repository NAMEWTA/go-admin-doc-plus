import { requestDomainCompat as request } from '../../../../packages/app-core/src'
import type { ApiResponse } from '../../../../packages/app-core/src'

/** Closes a websocket the scheduler opened for a job's log stream. */
export function unWsLogout(id: string | number, group: string) {
  return request<ApiResponse<null>>({
    url: '/wslogout/' + id + '/' + group,
    method: 'get'
  })
}
