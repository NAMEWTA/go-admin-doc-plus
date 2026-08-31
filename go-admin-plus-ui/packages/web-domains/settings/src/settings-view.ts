export type SettingsView = 'values' | 'dictionaries'

const settingsViews: Readonly<Record<string, SettingsView>> = {
  '/settings/values': 'values',
  '/settings/dictionaries': 'dictionaries',
}

export const settingsViewForPath = (path: string): SettingsView | null => settingsViews[path] ?? null
