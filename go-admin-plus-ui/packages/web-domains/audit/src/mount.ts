import { createApp } from 'vue'
import AuditPage from './AuditPage.vue'
import type { AuditController } from './audit-controller'

export const mountAuditPage = (element: Element | string, controller: AuditController) =>
  createApp(AuditPage, { controller }).mount(element)
