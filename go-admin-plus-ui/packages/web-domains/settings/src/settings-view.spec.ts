import { describe, expect, it } from 'vitest'
import pageSource from './SettingsPage.vue?raw'
import { settingsViewForPath } from './settings-view'

describe('settingsViewForPath', () => {
  it('maps only exact settings routes', () => {
    expect(settingsViewForPath('/settings/values')).toBe('values')
    expect(settingsViewForPath('/settings/dictionaries')).toBe('dictionaries')
    expect(settingsViewForPath('/settings/values/')).toBeNull()
    expect(settingsViewForPath('/settings/unknown')).toBeNull()
  })

  it('keeps the page route-derived without a local tab truth', () => {
    expect(pageSource).toContain('useRoute')
    expect(pageSource).toContain('settingsViewForPath(route.path)')
    expect(pageSource).not.toContain("tab=ref<'business'|'ui'|'dictionaries'>")
    expect(pageSource).not.toContain('switchTab')
  })
})
