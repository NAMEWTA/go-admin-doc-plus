import { spawn } from 'node:child_process'
import { join } from 'node:path'
import { chromium } from '@playwright/test'

const repoRoot = new URL('../../..', import.meta.url).pathname
const port = 9531
const baseUrl = `http://127.0.0.1:${port}`

const leaf = (path, routeKey, component) => ({
  path, routeKey, component, visible: '0', menuName: (routeKey ?? path).replaceAll(/[./:]/g, '-'),
  title: routeKey, noCache: false, children: null
})

const menuTree = [
  { path: '/admin', component: 'Layout', visible: '0', menuName: 'system-root', title: 'System', children: [
    leaf('sys-user', 'system.user', '/wrong/user'),
    leaf('sys-menu', 'system.menu', '/wrong/menu'),
    leaf('sys-role', 'system.role', '/wrong/role'),
    leaf('sys-dept', 'system.department', '/wrong/dept'),
    leaf('sys-post', 'system.post', '/wrong/post'),
    leaf('dict', 'system.dictionary', '/wrong/dict'),
    leaf('dict/data/:dictId', 'system.dictionary-data', '/wrong/dict-data'),
    leaf('sys-config', 'system.config', '/wrong/config'),
    leaf('sys-config/set', 'system.config-settings', '/wrong/config-settings'),
    leaf('sys-api', 'system.api', '/wrong/api'),
    leaf('sys-login-log', 'system.login-log', '/wrong/login-log'),
    leaf('sys-oper-log', 'system.operation-log', '/wrong/operation-log')
  ]},
  { path: '/schedule', component: 'Layout', visible: '0', menuName: 'jobs-root', title: 'Jobs', children: [
    leaf('manage', 'jobs.schedule', '/wrong/schedule'),
    leaf('log', 'jobs.log', '/wrong/log')
  ]},
  { path: '/demo', component: 'Layout', visible: '0', menuName: 'demo-root', title: 'Demo', children: [
    leaf('product', undefined, '/demo/product/index')
  ]},
  { path: '/dev-tools', component: 'Layout', visible: '0', menuName: 'tools-root', title: 'Tools', children: [
    leaf('swagger', 'tools.swagger', '/wrong/swagger'),
    leaf('gen', 'tools.generator', '/wrong/generator'),
    leaf('editTable', 'tools.generator-edit', '/wrong/generator-edit'),
    leaf('build', 'tools.form-builder', '/wrong/form-builder')
  ]},
  { path: '/sys-tools', component: 'Layout', visible: '0', menuName: 'monitor-root', title: 'Monitor', children: [
    leaf('monitor', 'monitor.server', '/wrong/monitor')
  ]},
  { path: '/diagnostics', component: 'Layout', visible: '1', menuName: 'diagnostics-root', title: 'Diagnostics', children: [
    leaf('unknown', 'demo.unknown', '/demo/product/index')
  ]}
]

const routes = [
  '/admin/sys-user', '/admin/sys-menu', '/admin/sys-role', '/admin/sys-dept',
  '/admin/sys-post', '/admin/dict', '/admin/dict/data/1', '/admin/sys-config',
  '/admin/sys-config/set', '/admin/sys-api', '/admin/sys-login-log', '/admin/sys-oper-log',
  '/schedule/manage', '/schedule/log', '/demo/product', '/dev-tools/swagger',
  '/dev-tools/gen', '/dev-tools/editTable?tableId=1', '/dev-tools/build', '/sys-tools/monitor'
]

const envelope = data => ({ code: 200, data, msg: 'ok' })

const mockData = pathname => {
  if (pathname === '/api/v1/getinfo') return {
    roles: ['admin'], name: 'route-check', avatar: '', introduction: '', permissions: ['*:*:*']
  }
  if (pathname === '/api/v1/menurole') return menuTree
  if (pathname === '/api/v1/app-config' || pathname === '/api/v1/set-config') return {}
  if (pathname.includes('option-select') || pathname.includes('Treeselect') || pathname === '/api/v1/deptTree') return []
  if (pathname === '/api/v1/dept' || pathname === '/api/v1/menu') return []
  if (pathname === '/api/v1/server-monitor') return {}
  if (pathname.includes('/sys/tables/info/')) return { info: {}, list: [] }
  if (pathname === '/api/v1/gen/tabletree') return []
  return { list: [], count: 0 }
}

const waitForServer = async() => {
  for (let attempt = 0; attempt < 120; attempt++) {
    try {
      const response = await fetch(baseUrl)
      if (response.ok) return
    } catch {}
    await new Promise(resolve => setTimeout(resolve, 250))
  }
  throw new Error(`Vite did not start at ${baseUrl}`)
}

const server = spawn(join(repoRoot, 'node_modules/.bin/vite'), [
  '--config', join(repoRoot, 'vite.config.mjs'), '--host', '127.0.0.1', '--port', String(port)
], { cwd: repoRoot, stdio: 'inherit', detached: process.platform !== 'win32' })

let browser
try {
  await waitForServer()
  browser = await chromium.launch()
  const context = await browser.newContext()
  await context.addCookies([{ name: 'Admin-Token', value: 'route-check-token', url: baseUrl }])
  const page = await context.newPage()
  const pageErrors = []
  page.on('pageerror', error => pageErrors.push(error.message))
  await page.route('**/*', async route => {
    const url = new URL(route.request().url())
    if (!url.pathname.startsWith('/api/') && !url.pathname.startsWith('/wslogout/')) {
      await route.continue()
      return
    }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(envelope(mockData(url.pathname)))
    })
  })

  for (const path of routes) {
    await page.goto(`${baseUrl}/#${path}`)
    await page.locator('.app-main').waitFor({ state: 'visible' })
    await page.locator('.app-main .page-container, .app-main .el-card, .app-main .job-log').first()
      .waitFor({ state: 'visible' })
    if (await page.locator('.route-unavailable').count()) throw new Error(`${path}: resolved as unavailable`)
    console.log(`Route OK: ${path}`)
  }

  await page.goto(`${baseUrl}/#/diagnostics/unknown`)
  await page.locator('.route-unavailable').waitFor({ state: 'visible' })
  if (await page.locator('.pro-table').count()) throw new Error('unknown routeKey fell back to its legacy component')
  if (pageErrors.length) throw new Error(`page errors: ${pageErrors.join(' | ')}`)
  console.log(`Route fail-closed OK: /diagnostics/unknown`)
  console.log(`Domain navigation: ${routes.length} routes passed`)
} finally {
  await browser?.close()
  if (process.platform === 'win32') server.kill()
  else if (server.pid) process.kill(-server.pid, 'SIGTERM')
}
