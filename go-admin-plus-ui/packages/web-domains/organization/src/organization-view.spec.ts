import { describe, expect, it } from 'vitest'
import pageSource from './OrganizationPage.vue?raw'
import { organizationViewForPath } from './organization-view'

describe('organizationViewForPath', () => {
  it('derives the organization management view only from an exact product route', () => {
    expect(organizationViewForPath('/organization/departments')).toBe('departments')
    expect(organizationViewForPath('/organization/positions')).toBe('positions')
    expect(organizationViewForPath('/organization/departments/')).toBeNull()
    expect(organizationViewForPath('/organization/unknown')).toBeNull()
  })

  it('keeps the rendered view route-derived without local tab navigation', () => {
    expect(pageSource).toContain('useRoute')
    expect(pageSource).toContain('organizationViewForPath(route.path)')
    expect(pageSource).not.toContain("const tab = ref")
    expect(pageSource).not.toContain('switchTab')
    expect(pageSource).not.toContain('class="tabs"')
  })
})
