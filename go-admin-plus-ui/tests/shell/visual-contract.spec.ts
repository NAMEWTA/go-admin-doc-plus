import { readFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const workspaceRoot = fileURLToPath(new URL('../..', import.meta.url))
const source = (path: string) => readFile(`${workspaceRoot}/${path}`, 'utf8')

describe('admin visual contract', () => {
  it('keeps the established shell composition and dimensions', async () => {
    const shell = await source('packages/app-shell/src/product/ProductWorkspace.vue')

    for (const region of [
      'product-shell__sidebar',
      'product-shell__header',
      'product-shell__breadcrumb',
      'product-shell__tags',
      'product-shell__content'
    ]) expect(shell).toContain(region)

    expect(shell).toMatch(/grid-template-columns:\s*210px/)
    expect(shell).toMatch(/height:\s*50px/)
    expect(shell).toMatch(/height:\s*40px/)
    expect(shell).toMatch(/grid-template-columns:\s*54px/)
    expect(shell).toContain('@media (max-width: 760px)')
  })

  it('shares the established management tokens across both hosts', async () => {
    const [theme, web, desktop, uiManifest] = await Promise.all([
      source('packages/ui/src/admin-theme.css'),
      source('apps/admin-web/src/main.ts'),
      source('apps/admin-desktop/src/main.ts'),
      source('packages/ui/package.json')
    ])

    for (const token of ['--ga-brand', '--ga-sidebar-bg', '--ga-bg-container', '--ga-border-light']) {
      expect(theme).toContain(token)
    }
    expect(theme).toContain('.product-shell__content')
    expect(web).toContain("@go-admin/ui/admin-theme.css")
    expect(desktop).toContain("@go-admin/ui/admin-theme.css")
    expect(uiManifest).toContain('"./admin-theme.css"')
  })

  it('preserves the terminal-stage login composition without restoring removed features', async () => {
    const login = await source('packages/web-domains/iam/src/session/LoginPage.vue')

    expect(login).toContain('login-page__stage')
    expect(login).toContain('login-page__terminal')
    expect(login).toContain('GopherMark')
    expect(login).toContain('Tauri 2')
    expect(login).not.toMatch(/captcha|验证码|uuid/i)
  })
})
