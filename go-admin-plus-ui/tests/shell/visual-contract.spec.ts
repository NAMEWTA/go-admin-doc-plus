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
    expect(shell).toMatch(/product-shell__header\s*\{[^}]*height:\s*50px/s)
    expect(shell).toMatch(/product-shell__nav-heading\s*\{[^}]*min-height:\s*50px/s)
    expect(shell).toMatch(/product-shell__navigation button\s*\{[^}]*min-height:\s*50px/s)
    expect(shell).toMatch(/product-shell__tags\s*\{[^}]*height:\s*34px/s)
    expect(shell).toMatch(/product-shell__content\s*\{[^}]*height:\s*calc\(100vh - 84px\)/s)
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
    for (const token of ['--ga-info', '--ga-bg-elevated', '--ga-text-inverse', '--ga-sidebar-light-bg']) {
      expect(theme).toContain(token)
    }
    expect(theme).toContain('.product-shell__content')
    for (const selector of ['.toolbar', '.filters', '.editor', '.pagination', '[data-action="delete"]']) {
      expect(theme).toContain(selector)
    }
    for (const selector of ['.management-toolbar', '.management-dialog-backdrop', '.management-dialog__header', '.management-dialog__body', '.management-dialog__footer']) {
      expect(theme).toContain(selector)
    }
    expect(web).toContain("@go-admin/ui/admin-theme.css")
    expect(desktop).toContain("@go-admin/ui/admin-theme.css")
    expect(uiManifest).toContain('"./admin-theme.css"')
  })

  it('does not let domain-scoped styles override the shared compact page surface', async () => {
    const pages = await Promise.all([
      source('packages/web-domains/iam/src/administration/AdministrationPage.vue'),
      source('packages/web-domains/audit/src/AuditPage.vue'),
      source('packages/web-domains/organization/src/OrganizationPage.vue'),
      source('packages/web-domains/settings/src/SettingsPage.vue'),
      source('packages/web-domains/generator/src/GeneratorWizardPage.vue'),
      source('packages/web-domains/scheduler/src/SchedulerPage.vue'),
      source('packages/web-domains/demo/src/DemoProductsPage.vue'),
      source('packages/web-domains/files/src/FilesPage.vue'),
      source('packages/web-domains/iam/src/session/AccountPage.vue')
    ])

    for (const page of pages) {
      expect(page).not.toMatch(/\.(?:administration-page|audit-page|organization-page|settings-page|generator-wizard|scheduler-page|demo-products|files-page|account-page)\s*\{[^}]*(?:max-width|padding):/s)
      expect(page).not.toMatch(/(?:input|select|button)[^{]*\{[^}]*min-height:\s*40px/s)
      expect(page).not.toMatch(/table\s*\{[^}]*border-collapse:\s*collapse/s)
    }
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
    expect(login).toContain("passwordVisible ? 'text' : 'password'")
    expect(login).toContain("passwordVisible ? '隐藏密码' : '显示密码'")
    expect(login).toContain('login-page__line--7')
    expect(login).toMatch(/login-page__terminal\s*\{[^}]*flex:\s*0 1 620px[^}]*border-radius:\s*10px/s)
    expect(login).toContain('@keyframes line-in')
    expect(login).toContain('@media (max-width: 1100px)')
    expect(login).not.toMatch(/captcha|验证码|uuid/i)
  })

  it('keeps the retained account fields in the established summary and tab composition', async () => {
    const account = await source('packages/web-domains/iam/src/session/AccountPage.vue')

    expect(account).toContain('account-page__workspace')
    expect(account).toContain('account-page__summary')
    expect(account).toContain('role="tablist"')
    expect(account).toContain("activeTab = ref<'profile' | 'password'>('profile')")
    expect(account).toContain("confirmPassword: ''")
    expect(account).toContain('passwordsMatch(password.newPassword, password.confirmPassword)')
    expect(account).toContain('两次输入的密码不一致。')
    expect(account).toMatch(/grid-template-columns:\s*minmax\(220px, 1fr\) minmax\(0, 3fr\)/)
    for (const retained of ['用户名称', '用户昵称', '用户邮箱', '头像元数据', '基本资料', '修改密码']) {
      expect(account).toContain(retained)
    }
    expect(account).not.toMatch(/手机号码|所属部门|所属角色|创建日期/)
  })

  it('keeps management creation and editing in explicit dialogs', async () => {
    const pages = await Promise.all([
      source('packages/web-domains/iam/src/administration/AdministrationPage.vue'),
      source('packages/web-domains/organization/src/OrganizationPage.vue'),
      source('packages/web-domains/demo/src/DemoProductsPage.vue'),
      source('packages/web-domains/settings/src/SettingsPage.vue'),
      source('packages/web-domains/scheduler/src/SchedulerPage.vue')
    ])

    const openControls = [
      'open-create-user',
      'open-create-department',
      'open-product-form',
      'open-setting-form',
      'open-scheduler-definition-form'
    ]
    pages.forEach((page, index) => {
      expect(page).toContain(`data-testid="${openControls[index]}"`)
      expect(page).toContain('management-dialog-backdrop')
      expect(page).toContain('role="dialog"')
      expect(page).not.toMatch(/class="[^"]*\beditor\b[^"]*"/)
    })
  })

  it('keeps paused browser and desktop drivers synchronized with localized UI labels', async () => {
    const [iam, organization, scheduler, audit, desktop] = await Promise.all([
      source('tests/e2e/iam/administration/browser-driver.ts'),
      source('tests/e2e/organization/browser-driver.ts'),
      source('tests/e2e/scheduler/browser-driver.ts'),
      source('tests/e2e/audit/browser-driver.ts'),
      source('tests/e2e/desktop/run.mjs')
    ])

    expect(iam).toContain("users: '用户管理'")
    expect(organization).toContain("openTab('部门管理'")
    expect(scheduler).toContain("button.textContent === '执行记录'")
    expect(audit).toContain("includes('会话已失效，请重新登录')")
    for (const label of ['账号', '密码', '登录', '产品示例', '退出登录']) expect(desktop).toContain(`'${label}'`)

    const drivers = [iam, organization, scheduler, audit, desktop].join('\n')
    for (const staleLabel of ["'Users'", "'Departments'", "'Executions'", "'Sign in again'", "'Sign out'"]) {
      expect(drivers).not.toContain(staleLabel)
    }
  })
})
