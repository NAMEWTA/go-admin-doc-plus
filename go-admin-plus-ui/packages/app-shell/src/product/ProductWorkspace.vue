<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'

import type { AccountProfile, SessionClient, SessionState } from '@go-admin-plus/domain-iam/session'
import { createSessionController } from '@go-admin-plus/domain-iam/session'
import type { ShellRuntimePort } from '@go-admin-plus/platform'
import { createBrowserFilesClient } from '@go-admin-plus/adapter-browser'
import { AuditPage, createAuditController, createWebAuditClient } from '@go-admin-plus/web-domain-audit'
import { createDemoController, createWebDemoClient, DemoProductsPage } from '@go-admin-plus/web-domain-demo'
import { createFilesController, FilesPage } from '@go-admin-plus/web-domain-files'
import { createGeneratorController, createWebGeneratorClient, GeneratorWizardPage } from '@go-admin-plus/web-domain-generator'
import { AccountPage, createWebSessionClient, LoginPage } from '@go-admin-plus/web-domain-iam/session'
import { AdministrationPage, createAdministrationController, createWebAdministrationClient } from '@go-admin-plus/web-domain-iam/administration'
import { createOrganizationController, createWebOrganizationClient, OrganizationPage } from '@go-admin-plus/web-domain-organization'
import { createSchedulerController, createWebSchedulerClient, SchedulerPage } from '@go-admin-plus/web-domain-scheduler'
import { createSettingsController, createWebSettingsClient, SettingsPage, type SettingsRemovalKind } from '@go-admin-plus/web-domain-settings'

import { productRoutesFor, type ProductHost } from './manifest'

const props = defineProps<{
  host: ProductHost
  runtime: ShellRuntimePort
  fetcher?: typeof globalThis.fetch
  sessionClient?: SessionClient
}>()

type View = 'loading' | 'login' | 'workspace' | 'account' | 'forbidden' | 'not-found' | 'unavailable'

const fetcher = props.fetcher ?? globalThis.fetch
const session = createSessionController(props.sessionClient ?? createWebSessionClient(fetcher))
const sessionState = ref<SessionState>(session.state())
const permissions = ref<ReadonlySet<string>>(new Set())
const dataScope = ref<'self' | 'all' | null>(null)
const navigationPaths = ref<ReadonlySet<string>>(new Set())
const path = ref(window.location.pathname)
const view = ref<View>('loading')
const sidebarCollapsed = ref(false)
const mobileNavigationOpen = ref(false)
const accountMenuOpen = ref(false)
const visitedPaths = ref<string[]>([])

const capability = {
  can: (permission: string) => permissions.value.has(permission),
  scope: () => dataScope.value
}
const confirmRemoval = async (count: number) => window.confirm(count === 1 ? '确定删除该记录吗？' : `确定删除所选的 ${count} 条记录吗？`)
const confirmAuditCleanup = async () => window.confirm('确定清理所选日期之前且符合保留策略的审计日志吗？')
const confirmSetting = async (kind: SettingsRemovalKind) => {
  const labels: Record<SettingsRemovalKind, string> = { setting: '参数', dictionary: '字典', 'dictionary-item': '字典项' }
  return window.confirm(`确定删除该${labels[kind]}吗？`)
}

const administration = createAdministrationController(createWebAdministrationClient(fetcher), confirmRemoval)
const audit = createAuditController(createWebAuditClient(fetcher), confirmAuditCleanup)
const organization = createOrganizationController(createWebOrganizationClient(fetcher), capability, confirmRemoval)
const settings = createSettingsController(createWebSettingsClient(fetcher), confirmSetting, capability)
const generator = createGeneratorController(createWebGeneratorClient(fetcher), capability)
const scheduler = createSchedulerController(createWebSchedulerClient(fetcher), capability, confirmRemoval)
const demo = createDemoController(createWebDemoClient(fetcher), confirmRemoval, capability)
const files = createFilesController(createBrowserFilesClient(fetcher), confirmRemoval, capability)

const routes = computed(() => productRoutesFor(props.host).filter(route =>
  navigationPaths.value.has(route.path) && permissions.value.has(route.permission)
))
const profile = computed<AccountProfile | null>(() => sessionState.value.profile)
const route = computed(() => productRoutesFor(props.host).find(candidate => candidate.path === path.value))
const routeGroups = computed(() => {
  const labels: Record<string, string> = {
    iam: '权限管理',
    audit: '审计管理',
    organization: '组织管理',
    settings: '系统设置',
    generator: '开发工具',
    scheduler: '任务调度',
    demo: '示例业务',
    files: '文件管理'
  }
  const groups = new Map<string, { key: string; label: string; mark: string; routes: typeof routes.value }>()
  for (const item of routes.value) {
    const key = item.path.split('/')[1] ?? 'system'
    const existing = groups.get(key)
    if (existing) existing.routes = [...existing.routes, item]
    else groups.set(key, { key, label: labels[key] ?? key, mark: key.slice(0, 2).toUpperCase(), routes: [item] })
  }
  return [...groups.values()]
})
const visitedRoutes = computed(() => visitedPaths.value
  .map(visited => productRoutesFor(props.host).find(candidate => candidate.path === visited))
  .filter(candidate => candidate !== undefined))
const currentTitle = computed(() => path.value === '/account' ? '个人中心' : route.value?.label ?? '工作台')
const profileMark = computed(() => (profile.value?.displayName.trim().charAt(0) || 'A').toUpperCase())
const module = computed(() => {
  const current = route.value
  if (!current) return null
  if (current.path.startsWith('/iam/')) return 'iam'
  return current.path.split('/')[1] ?? null
})

const replacePath = (next: string) => {
  window.history.replaceState({}, '', next)
  path.value = next
}
const rememberRoute = (next: string) => {
  if (!productRoutesFor(props.host).some(candidate => candidate.path === next)) return
  visitedPaths.value = [...visitedPaths.value.filter(visited => visited !== next), next]
}
const navigate = (next: string) => {
  window.history.pushState({}, '', next)
  path.value = next
  rememberRoute(next)
  mobileNavigationOpen.value = false
  accountMenuOpen.value = false
  resolveView()
}
const resolveView = () => {
  if (sessionState.value.status !== 'authenticated') { view.value = 'login'; return }
  if (path.value === '/account') { view.value = 'account'; return }
  if (path.value === '/' || path.value === '/login') {
    const first = routes.value[0]?.path
    if (first) { replacePath(first); rememberRoute(first); view.value = 'workspace' } else view.value = 'forbidden'
    return
  }
  if (!route.value) { view.value = 'not-found'; return }
  if (routes.value.some(candidate => candidate.path === path.value)) {
    rememberRoute(path.value)
    view.value = 'workspace'
  } else view.value = 'forbidden'
}
const loadRuntime = async () => {
  view.value = 'loading'
  try {
    const identity = await props.runtime.loadIdentity()
    if (identity.kind === 'unauthenticated') {
      permissions.value = new Set()
      navigationPaths.value = new Set()
      view.value = 'login'
      return
    }
    permissions.value = new Set(identity.permissions)
    dataScope.value = 'dataScope' in identity && (identity.dataScope === 'self' || identity.dataScope === 'all')
      ? identity.dataScope
      : 'all'
    const navigation = await props.runtime.loadNavigation()
    navigationPaths.value = new Set(navigation.map(entry => entry.path))
    resolveView()
  } catch {
    view.value = 'unavailable'
  }
}
const authenticated = async () => { await loadRuntime() }
const requireSession = () => {
  permissions.value = new Set()
  navigationPaths.value = new Set()
  replacePath('/login')
  view.value = 'login'
}
const forbid = () => { view.value = 'forbidden' }
const signedOut = () => requireSession()
const signOut = async () => {
  accountMenuOpen.value = false
  if (!window.confirm('确定注销并退出系统吗？')) return
  await session.logout()
  if (session.state().status === 'unauthenticated') signedOut()
}
const closeTag = (target: string) => {
  if (visitedPaths.value.length <= 1) return
  const index = visitedPaths.value.indexOf(target)
  visitedPaths.value = visitedPaths.value.filter(visited => visited !== target)
  if (target !== path.value) return
  const fallback = visitedPaths.value[Math.min(index, visitedPaths.value.length - 1)] ?? routes.value[0]?.path
  if (fallback) navigate(fallback)
}
const restore = async () => {
  view.value = 'loading'
  await session.restore()
  if (session.state().status === 'authenticated') await loadRuntime()
  else view.value = 'login'
}
const handlePopState = () => { path.value = window.location.pathname; rememberRoute(path.value); resolveView() }
const unsubscribe = session.subscribe(state => {
  sessionState.value = state
  if (state.status === 'unauthenticated' && view.value !== 'loading') requireSession()
})

onMounted(() => {
  window.addEventListener('popstate', handlePopState)
  void restore()
})
onUnmounted(() => {
  unsubscribe()
  window.removeEventListener('popstate', handlePopState)
})
</script>

<template>
  <main class="product-shell" :data-host="host" :data-shell-state="view">
    <section v-if="view === 'loading'" class="product-shell__state" aria-live="polite">
      <span class="product-shell__spinner" aria-hidden="true" />
      <p>正在加载</p>
    </section>

    <LoginPage v-else-if="view === 'login'" :controller="session" @authenticated="authenticated" />

    <section v-else-if="view === 'workspace' || view === 'account'" class="product-shell__workspace" :class="{ 'is-collapsed': sidebarCollapsed }">
      <button v-if="mobileNavigationOpen" class="product-shell__drawer-backdrop" type="button" aria-label="关闭导航" @click="mobileNavigationOpen = false" />
      <aside class="product-shell__sidebar" :class="{ 'is-mobile-open': mobileNavigationOpen }">
        <button class="product-shell__brand" type="button" title="Go Admin Plus" @click="navigate('/')">
          <span class="product-shell__brand-mark">G</span>
          <span class="product-shell__brand-name">Go Admin Plus</span>
        </button>
        <nav class="product-shell__navigation" aria-label="主导航">
          <section v-for="group in routeGroups" :key="group.key" class="product-shell__nav-group">
            <p class="product-shell__nav-heading"><span class="product-shell__nav-mark">{{ group.mark }}</span><span>{{ group.label }}</span></p>
            <button v-for="item in group.routes" :key="item.key" type="button" :title="item.label" :aria-current="path === item.path ? 'page' : undefined" @click="navigate(item.path)">
              <span class="product-shell__nav-dot" aria-hidden="true" />
              <span class="product-shell__nav-label">{{ item.label }}</span>
            </button>
          </section>
        </nav>
      </aside>

      <div class="product-shell__main">
        <header class="product-shell__header">
          <button class="product-shell__hamburger" type="button" aria-label="切换导航" @click="mobileNavigationOpen = !mobileNavigationOpen; sidebarCollapsed = !sidebarCollapsed">
            <span /><span /><span />
          </button>
          <nav class="product-shell__breadcrumb" aria-label="面包屑">
            <button type="button" @click="navigate('/')">工作台</button><span>/</span><strong>{{ currentTitle }}</strong>
          </nav>
          <div class="product-shell__identity">
            <span class="product-shell__host">{{ host === 'desktop' ? 'Desktop' : 'Web' }}</span>
            <button v-if="profile" class="product-shell__profile" type="button" :aria-expanded="accountMenuOpen" @click="accountMenuOpen = !accountMenuOpen">
              <span class="product-shell__avatar">{{ profileMark }}</span><span>{{ profile.displayName }}</span><span aria-hidden="true">⌄</span>
            </button>
            <div v-if="accountMenuOpen" class="product-shell__account-menu">
              <button type="button" @click="navigate('/account')">个人中心</button>
              <button type="button" @click="signOut">退出登录</button>
            </div>
          </div>
        </header>

        <div class="product-shell__tags" aria-label="已访问页面">
          <div v-for="item in visitedRoutes" :key="item.path" class="product-shell__tag" :class="{ 'is-active': path === item.path }">
            <button type="button" :aria-current="path === item.path ? 'page' : undefined" @click="navigate(item.path)">{{ item.label }}</button>
            <button v-if="visitedRoutes.length > 1" class="product-shell__tag-close" type="button" :aria-label="`关闭 ${item.label}`" @click="closeTag(item.path)">×</button>
          </div>
        </div>

        <div class="product-shell__content">
          <AdministrationPage v-if="view === 'workspace' && module === 'iam'" :controller="administration" @session-required="requireSession" />
          <AuditPage v-else-if="view === 'workspace' && module === 'audit'" :controller="audit" @relogin="requireSession" />
          <OrganizationPage v-else-if="view === 'workspace' && module === 'organization'" :controller="organization" @session-required="requireSession" />
          <SettingsPage v-else-if="view === 'workspace' && module === 'settings'" :controller="settings" @session-required="requireSession" @forbidden="forbid" />
          <GeneratorWizardPage v-else-if="view === 'workspace' && module === 'generator'" :controller="generator" @session-required="requireSession" @forbidden="forbid" />
          <SchedulerPage v-else-if="view === 'workspace' && module === 'scheduler'" :controller="scheduler" @session-required="requireSession" />
          <DemoProductsPage v-else-if="view === 'workspace' && module === 'demo'" :controller="demo" @session-required="requireSession" @forbidden="forbid" />
          <FilesPage v-else-if="view === 'workspace' && module === 'files'" :controller="files" @session-required="requireSession" />
          <AccountPage v-else-if="view === 'account' && profile" :controller="session" :profile="profile" @signed-out="signedOut" />
        </div>
      </div>
    </section>

    <section v-else class="product-shell__state">
      <p class="product-shell__code">{{ view === 'forbidden' ? '403' : view === 'not-found' ? '404' : 'RUNTIME' }}</p>
      <h1>{{ view === 'forbidden' ? '无权访问' : view === 'not-found' ? '页面不存在' : '服务暂不可用' }}</h1>
      <button type="button" @click="view === 'unavailable' ? restore() : navigate('/')">{{ view === 'unavailable' ? '重试' : '返回工作台' }}</button>
    </section>
  </main>
</template>

<style scoped>
.product-shell { min-height: 100vh; color: var(--ga-text-1); background: var(--ga-bg-body); }
.product-shell__workspace { display: grid; min-height: 100vh; grid-template-columns: 210px minmax(0, 1fr); transition: grid-template-columns .28s; }
.product-shell__sidebar { position: fixed; inset: 0 auto 0 0; z-index: 20; width: 210px; overflow: hidden; color: var(--ga-sidebar-text); background: var(--ga-sidebar-bg); transition: width .28s, transform .28s; }
.product-shell__brand { display: flex; width: 100%; height: 64px; align-items: center; justify-content: center; gap: 10px; padding: 0 12px; color: #fff; background: linear-gradient(135deg, color-mix(in srgb, var(--ga-brand), #000 45%), var(--ga-brand)); border: 0; cursor: pointer; }
.product-shell__brand-mark { display: grid; width: 30px; height: 30px; flex: 0 0 30px; place-items: center; border: 1px solid rgb(255 255 255 / 35%); border-radius: 6px; font-size: 17px; font-weight: 700; }
.product-shell__brand-name { overflow: hidden; font-size: 15px; font-weight: 700; white-space: nowrap; }
.product-shell__navigation { height: calc(100vh - 64px); padding: 8px 0 24px; overflow: auto; }
.product-shell__nav-group { margin: 0; padding: 0; }
.product-shell__nav-heading { display: flex; min-height: 50px; align-items: center; gap: 12px; margin: 0; padding: 0 20px; color: rgb(255 255 255 / 88%); font-size: 13px; font-weight: 600; }
.product-shell__nav-mark { display: grid; width: 22px; height: 22px; flex: 0 0 22px; place-items: center; color: rgb(255 255 255 / 72%); border: 1px solid rgb(255 255 255 / 20%); border-radius: 4px; font-size: 9px; }
.product-shell__navigation button { display: flex; width: 100%; min-height: 50px; align-items: center; gap: 12px; padding: 0 20px 0 33px; overflow: hidden; color: var(--ga-sidebar-text); background: var(--ga-sidebar-bg-sub); border: 0; cursor: pointer; text-align: left; }
.product-shell__navigation button:hover { color: #fff; background: var(--ga-sidebar-hover); }
.product-shell__navigation button[aria-current='page'] { color: #fff; background: var(--ga-sidebar-hover); box-shadow: inset 3px 0 var(--ga-brand); }
.product-shell__nav-dot { width: 6px; height: 6px; flex: 0 0 6px; border: 1px solid currentColor; border-radius: 50%; }
.product-shell__nav-label { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.product-shell__main { min-width: 0; grid-column: 2; }
.product-shell__header { position: relative; z-index: 10; display: flex; height: 50px; align-items: center; padding: 0 20px 0 0; background: var(--ga-bg-container); border-bottom: 1px solid var(--ga-border-light); box-shadow: var(--ga-shadow-sm); }
.product-shell__header::after { position: absolute; right: 0; bottom: 0; left: 0; height: 2px; background: linear-gradient(90deg, var(--ga-brand), color-mix(in srgb, var(--ga-brand), white 30%)); content: ''; opacity: .7; }
.product-shell__hamburger { display: grid; width: 50px; height: 100%; place-content: center; gap: 4px; padding: 0; background: transparent; border: 0; cursor: pointer; }
.product-shell__hamburger:hover { background: var(--ga-bg-hover); }
.product-shell__hamburger span { display: block; width: 18px; height: 2px; background: var(--ga-brand); }
.product-shell__breadcrumb { display: flex; min-width: 0; align-items: center; gap: 10px; font-size: 13px; }
.product-shell__breadcrumb button { padding: 0; color: var(--ga-text-2); background: transparent; border: 0; cursor: pointer; }
.product-shell__breadcrumb span { color: var(--ga-text-3); }
.product-shell__breadcrumb strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.product-shell__identity { position: relative; display: flex; align-items: center; gap: 12px; margin-left: auto; }
.product-shell__host { padding: 3px 8px; color: var(--ga-text-3); background: var(--ga-bg-subtle); border: 1px solid var(--ga-border-light); border-radius: 4px; font-size: 11px; font-weight: 600; }
.product-shell__profile { display: flex; height: 42px; align-items: center; gap: 7px; padding: 0 4px; color: var(--ga-text-2); background: transparent; border: 0; cursor: pointer; }
.product-shell__profile:hover { color: var(--ga-brand); background: var(--ga-bg-hover); }
.product-shell__avatar { display: grid; width: 32px; height: 32px; place-items: center; color: #fff; background: var(--ga-brand); border: 2px solid color-mix(in srgb, var(--ga-brand), white 70%); border-radius: 50%; font-size: 12px; font-weight: 700; }
.product-shell__account-menu { position: absolute; top: 45px; right: 0; z-index: 30; display: grid; width: 128px; padding: 4px 0; background: var(--ga-bg-container); border: 1px solid var(--ga-border); border-radius: 4px; box-shadow: var(--ga-shadow-lg); }
.product-shell__account-menu button { min-height: 34px; padding: 0 14px; color: var(--ga-text-2); background: transparent; border: 0; cursor: pointer; text-align: left; }
.product-shell__account-menu button:hover { color: var(--ga-brand); background: var(--ga-bg-hover); }
.product-shell__tags { display: flex; height: 34px; align-items: end; gap: 4px; padding: 0 4px; overflow-x: auto; background: var(--ga-bg-container); border-bottom: 1px solid var(--ga-border-light); box-shadow: var(--ga-shadow-sm); }
.product-shell__tag { display: flex; height: 32px; flex: 0 0 auto; align-items: center; padding: 0 5px 0 10px; color: var(--ga-text-2); background: var(--ga-bg-subtle); border: 1px solid var(--ga-border); border-bottom: 0; border-radius: 3px 3px 0 0; font-size: 12px; }
.product-shell__tag.is-active { color: var(--ga-brand); background: var(--ga-bg-container); border-color: color-mix(in srgb, var(--ga-brand), white 50%); font-weight: 500; }
.product-shell__tag > button { height: 100%; padding: 0; color: inherit; background: transparent; border: 0; cursor: pointer; }
.product-shell__tag-close { display: grid; width: 18px; height: 18px !important; margin-left: 4px; place-items: center; border-radius: 50% !important; line-height: 1; }
.product-shell__tag-close:hover { color: #fff; background: var(--ga-text-3); }
.product-shell__content { min-width: 0; height: calc(100vh - 84px); padding: 12px; overflow: auto; }
.product-shell__state { display: grid; min-height: 100vh; place-content: center; justify-items: center; gap: 12px; padding: 24px; text-align: center; }
.product-shell__spinner { width: 28px; height: 28px; border: 3px solid var(--ga-border); border-top-color: var(--ga-brand); border-radius: 50%; animation: spin .8s linear infinite; }
.product-shell__code { font-weight: 700; color: var(--ga-text-3); }
.product-shell__state button { min-height: 36px; padding: 0 14px; color: #fff; background: var(--ga-brand); border: 1px solid var(--ga-brand); border-radius: var(--ga-radius); cursor: pointer; }
.product-shell__workspace.is-collapsed { grid-template-columns: 54px minmax(0, 1fr); }
.is-collapsed .product-shell__sidebar { width: 54px; }
.is-collapsed .product-shell__brand { padding: 0; }
.is-collapsed :is(.product-shell__brand-name, .product-shell__nav-heading > span:last-child, .product-shell__nav-label) { display: none; }
.is-collapsed .product-shell__nav-heading { justify-content: center; padding: 0; }
.is-collapsed .product-shell__navigation button { justify-content: center; padding: 0; }
.product-shell__drawer-backdrop { display: none; }
@keyframes spin { to { transform: rotate(360deg); } }
@media (max-width: 760px) {
  .product-shell__workspace, .product-shell__workspace.is-collapsed { grid-template-columns: 1fr; }
  .product-shell__sidebar, .is-collapsed .product-shell__sidebar { width: 210px; transform: translateX(-100%); }
  .product-shell__sidebar.is-mobile-open { transform: translateX(0); }
  .product-shell__main { grid-column: 1; }
  .product-shell__drawer-backdrop { position: fixed; inset: 0; z-index: 15; display: block; padding: 0; background: rgb(0 0 0 / 30%); border: 0; }
  .product-shell__header { padding-right: 10px; }
  .product-shell__breadcrumb strong { max-width: 110px; }
  .product-shell__host, .product-shell__profile > span:nth-child(2) { display: none; }
  .product-shell__content { height: calc(100vh - 84px); padding: 0; }
  .is-collapsed :is(.product-shell__brand-name, .product-shell__nav-heading > span:last-child, .product-shell__nav-label) { display: initial; }
  .is-collapsed .product-shell__brand { padding: 0 12px; }
  .is-collapsed .product-shell__nav-heading { justify-content: flex-start; padding: 0 20px; }
  .is-collapsed .product-shell__navigation button { justify-content: flex-start; padding: 0 20px 0 33px; }
}
@media (prefers-reduced-motion: reduce) { .product-shell__workspace, .product-shell__sidebar { transition: none; } }
</style>
