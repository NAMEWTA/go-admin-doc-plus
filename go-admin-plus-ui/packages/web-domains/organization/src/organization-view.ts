export type OrganizationView = 'departments' | 'positions'

const organizationViews = new Map<string, OrganizationView>([
  ['/organization/departments', 'departments'],
  ['/organization/positions', 'positions'],
])

export const organizationViewForPath = (path: string): OrganizationView | null =>
  organizationViews.get(path) ?? null
