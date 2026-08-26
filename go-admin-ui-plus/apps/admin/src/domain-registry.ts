import { createDomainRegistry } from '../../../packages/app-core/src'
import { systemDomain } from '../../../domains/system/src'
import { jobsDomain } from '../../../domains/jobs/src'
import { demoDomain } from '../../../domains/demo/src'
import { toolsDomain } from '../../../domains/tools/src'
import { monitorDomain } from '../../../domains/monitor/src'

export const domainRegistry = createDomainRegistry([
  systemDomain,
  jobsDomain,
  demoDomain,
  toolsDomain,
  monitorDomain
])
