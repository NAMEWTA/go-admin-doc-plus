<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'

import type { AccountProfile, SessionClient, SessionState } from '@go-admin/domain-iam/session'
import { createSessionController } from '@go-admin/domain-iam/session'
import type { ShellRuntimePort } from '@go-admin/platform'
import { createBrowserFilesClient } from '@go-admin/adapter-browser'
import { AuditPage, createAuditController, createWebAuditClient } from '@go-admin/web-domain-audit'
import { createDemoController, createWebDemoClient, DemoProductsPage } from '@go-admin/web-domain-demo'
import { createFilesController, FilesPage } from '@go-admin/web-domain-files'
import { createGeneratorController, createWebGeneratorClient, GeneratorWizardPage } from '@go-admin/web-domain-generator'
import { AccountPage, createWebSessionClient, LoginPage } from '@go-admin/web-domain-iam/session'
import { AdministrationPage, createAdministrationController, createWebAdministrationClient } from '@go-admin/web-domain-iam/administration'
import { createOrganizationController, createWebOrganizationClient, OrganizationPage } from '@go-admin/web-domain-organization'
import { createSchedulerController, createWebSchedulerClient, SchedulerPage } from '@go-admin/web-domain-scheduler'
import { createSettingsController, createWebSettingsClient, SettingsPage } from '@go-admin/web-domain-settings'

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

const capability = {
  can: (permission: string) => permissions.value.has(permission),
  scope: () => dataScope.value
}
const confirmRemoval = async (count: number) => window.confirm(`Delete ${count} item${count === 1 ? '' : 's'}?`)
const confirmSetting = async (kind: string) => window.confirm(`Delete this ${kind}?`)

const administration = createAdministrationController(createWebAdministrationClient(fetcher), confirmRemoval)
const audit = createAuditController(createWebAuditClient(fetcher), confirmRemoval)
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
const navigate = (next: string) => {
  window.history.pushState({}, '', next)
  path.value = next
  resolveView()
}
const resolveView = () => {
  if (sessionState.value.status !== 'authenticated') { view.value = 'login'; return }
  if (path.value === '/account') { view.value = 'account'; return }
  if (path.value === '/' || path.value === '/login') {
    const first = routes.value[0]?.path
    if (first) { replacePath(first); view.value = 'workspace' } else view.value = 'forbidden'
    return
  }
  if (!route.value) { view.value = 'not-found'; return }
  view.value = routes.value.some(candidate => candidate.path === path.value) ? 'workspace' : 'forbidden'
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
const restore = async () => {
  view.value = 'loading'
  await session.restore()
  if (session.state().status === 'authenticated') await loadRuntime()
  else view.value = 'login'
}
const handlePopState = () => { path.value = window.location.pathname; resolveView() }
const unsubscribe = session.subscribe(state => { sessionState.value = state })

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
    <header class="product-shell__header">
      <button class="product-shell__brand" type="button" @click="navigate('/')">Go Admin Plus</button>
      <div class="product-shell__identity">
        <span>{{ host === 'desktop' ? 'Desktop' : 'Web' }}</span>
        <button v-if="profile" type="button" @click="navigate('/account')">{{ profile.displayName }}</button>
      </div>
    </header>

    <section v-if="view === 'loading'" class="product-shell__state" aria-live="polite">
      <span class="product-shell__spinner" aria-hidden="true" />
      <p>正在加载</p>
    </section>

    <LoginPage v-else-if="view === 'login'" :controller="session" @authenticated="authenticated" />

    <section v-else-if="view === 'workspace'" class="product-shell__workspace">
      <nav class="product-shell__navigation" aria-label="主导航">
        <button v-for="item in routes" :key="item.key" type="button" :aria-current="path === item.path ? 'page' : undefined" @click="navigate(item.path)">
          {{ item.label }}
        </button>
      </nav>
      <div class="product-shell__content">
        <AdministrationPage v-if="module === 'iam'" :controller="administration" @session-required="requireSession" />
        <AuditPage v-else-if="module === 'audit'" :controller="audit" @relogin="requireSession" />
        <OrganizationPage v-else-if="module === 'organization'" :controller="organization" @session-required="requireSession" />
        <SettingsPage v-else-if="module === 'settings'" :controller="settings" @session-required="requireSession" @forbidden="forbid" />
        <GeneratorWizardPage v-else-if="module === 'generator'" :controller="generator" @session-required="requireSession" @forbidden="forbid" />
        <SchedulerPage v-else-if="module === 'scheduler'" :controller="scheduler" @session-required="requireSession" />
        <DemoProductsPage v-else-if="module === 'demo'" :controller="demo" @session-required="requireSession" @forbidden="forbid" />
        <FilesPage v-else-if="module === 'files'" :controller="files" @session-required="requireSession" />
      </div>
    </section>

    <AccountPage v-else-if="view === 'account' && profile" :controller="session" :profile="profile" @signed-out="signedOut" />

    <section v-else class="product-shell__state">
      <p class="product-shell__code">{{ view === 'forbidden' ? '403' : view === 'not-found' ? '404' : 'RUNTIME' }}</p>
      <h1>{{ view === 'forbidden' ? '无权访问' : view === 'not-found' ? '页面不存在' : '服务暂不可用' }}</h1>
      <button type="button" @click="view === 'unavailable' ? restore() : navigate('/')">{{ view === 'unavailable' ? '重试' : '返回工作台' }}</button>
    </section>
  </main>
</template>

<style scoped>
.product-shell { min-height: 100vh; background: #f5f7f8; color: #17202a; }
.product-shell__header { display: flex; min-height: 56px; align-items: center; justify-content: space-between; padding: 0 20px; border-bottom: 1px solid #d9e0e3; background: #fff; }
.product-shell__brand { border: 0; background: transparent; color: #17202a; font: 700 16px/1.2 system-ui, sans-serif; }
.product-shell__identity { display: flex; align-items: center; gap: 12px; }
.product-shell__workspace { display: grid; min-height: calc(100vh - 57px); grid-template-columns: 220px minmax(0, 1fr); }
.product-shell__navigation { display: grid; align-content: start; gap: 2px; padding: 12px; border-right: 1px solid #d9e0e3; background: #fff; }
.product-shell__navigation button { min-height: 38px; padding: 0 10px; border: 0; background: transparent; text-align: left; }
.product-shell__navigation button[aria-current='page'] { background: #e7f0ec; color: #135c43; font-weight: 700; }
.product-shell__content { min-width: 0; padding: 20px; overflow: auto; }
.product-shell__state { display: grid; min-height: calc(100vh - 57px); place-content: center; justify-items: center; gap: 12px; }
.product-shell__spinner { width: 28px; height: 28px; border: 3px solid #ccd5d8; border-top-color: #135c43; border-radius: 50%; animation: spin .8s linear infinite; }
.product-shell__code { font-weight: 700; color: #6d777b; }
button { cursor: pointer; font: inherit; }
@keyframes spin { to { transform: rotate(360deg); } }
@media (max-width: 760px) {
  .product-shell__workspace { grid-template-columns: 1fr; }
  .product-shell__navigation { grid-auto-flow: column; grid-auto-columns: minmax(140px, max-content); overflow-x: auto; border-right: 0; border-bottom: 1px solid #d9e0e3; }
}
</style>
