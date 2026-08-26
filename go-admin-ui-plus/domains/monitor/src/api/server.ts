import { requestDomainCompat as request } from '../../../../packages/app-core/src'
import type { ApiResponse } from '../../../../packages/app-core/src'

/** Host metrics behind the 服务监控 screen. Shape follows whatever the Go side sends. */
export function getServer() {
  return request<ApiResponse<Record<string, unknown>>>({
    url: '/api/v1/server-monitor',
    method: 'get'
  })
}
