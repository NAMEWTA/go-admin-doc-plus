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
    expect(shell).toMatch(/product-shell__content\s*\{[^}]*padding:\s*12px/s)
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
    for (const selector of ['.toolbar', '.filters', '.editor', '.pagination', '[data-action="delete"]']) {
      expect(theme).toContain(selector)
    }
    expect(web).toContain("@go-admin/ui/admin-theme.css")
    expect(desktop).toContain("@go-admin/ui/admin-theme.css")
    expect(uiManifest).toContain('"./admin-theme.css"')
  })

  it('keeps the retained administration surface localized and on shared tokens', async () => {
    const [manifest, administration, ...domainPages] = await Promise.all([
      source('packages/app-shell/src/product/manifest.ts'),
      source('packages/web-domains/iam/src/administration/AdministrationPage.vue'),
      source('packages/web-domains/audit/src/AuditPage.vue'),
      source('packages/web-domains/demo/src/DemoProductsPage.vue'),
      source('packages/web-domains/files/src/FilesPage.vue'),
      source('packages/web-domains/generator/src/GeneratorWizardPage.vue'),
      source('packages/web-domains/organization/src/OrganizationPage.vue'),
      source('packages/web-domains/scheduler/src/SchedulerPage.vue'),
      source('packages/web-domains/settings/src/SettingsPage.vue')
    ])

    for (const label of ['用户管理', '角色管理', '菜单管理', '部门管理', '参数设置', '代码生成', '任务调度', '文件管理']) {
      expect(manifest).toContain(`label: '${label}'`)
    }
    for (const text of ['用户与权限', '新增用户', '角色管理', '菜单管理', '当前账号没有可用的管理视图']) {
      expect(administration).toContain(text)
    }
    expect(administration).not.toMatch(/>\s*(Identity and access|Users|Roles|Menus|Create user|Create role|Create menu)\s*</)

    const obsoleteColors = ['#17202a', '#176b54', '#d7dce2', '#dfe6e9', '#aab2bc', '#4c5968', '#4a5561']
    for (const page of [administration, ...domainPages]) {
      for (const color of obsoleteColors) expect(page).not.toContain(color)
    }
  })

  it('preserves the terminal-stage login composition without restoring removed features', async () => {
    const login = await source('packages/web-domains/iam/src/session/LoginPage.vue')

    expect(login).toContain('login-page__stage')
    expect(login).toContain('login-page__terminal')
    expect(login).toContain('GopherMark')
    expect(login).toContain('Tauri 2')
    expect(login).toContain('http://localhost:5173')
    expect(login).not.toContain('http://localhost:3000')
    expect(login).not.toMatch(/captcha|验证码|uuid/i)
  })
})
