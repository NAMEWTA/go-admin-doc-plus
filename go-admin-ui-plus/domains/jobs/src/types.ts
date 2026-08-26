export interface SysJob {
  jobId?: number
  jobName?: string
  jobGroup?: string
  jobType?: number
  cronExpression?: string
  invokeTarget?: string
  args?: string
  misfirePolicy?: number
  concurrent?: number
  status?: number | string
  entry_id?: number
  createdAt?: string
}

export interface SysJobQuery {
  jobName?: string
  jobGroup?: string
  status?: string
}
