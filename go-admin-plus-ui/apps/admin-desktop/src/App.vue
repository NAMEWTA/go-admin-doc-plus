<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'

import { createDesktopDemoClient, createDesktopRuntime, createDesktopSession, createDesktopTransport } from '@go-admin/adapter-desktop'
import { createShellNavigator } from '@go-admin/app-shell'
import { DemoRequestError, demoPermissions, type DemoPermissionCode } from '@go-admin/domain-demo'
import { createDemoController, DemoProductsPage } from '@go-admin/web-domain-demo'

type View = 'loading' | 'login' | 'workspace' | 'forbidden' | 'unavailable'

const runtime = createDesktopRuntime()
const session = createDesktopSession()
const transport = createDesktopTransport()
const demoClient = createDesktopDemoClient(transport)
const nativeE2E = import.meta.env.VITE_GO_ADMIN_NATIVE_E2E === '1'
const permissions = ref<ReadonlySet<string>>(new Set())
const dataScope = ref<'self' | 'all' | null>(null)
const view = ref<View>('loading')
const busy = ref(false)
const loginError = ref('')
const logoutError = ref('')
const nativeBoundaryStage = ref<'' | 'startup' | 'unauthenticated' | 'authenticated'>('')
const nativeAuthorization = ref('')
const credentials = reactive({ username: '', password: '' })
const capabilities = { can: (permission: DemoPermissionCode) => permissions.value.has(permission) }
const controller = createDemoController(
  demoClient,
  async count => window.confirm(`Delete ${count} product${count === 1 ? '' : 's'}?`),
  capabilities
)
const canReadDemo = computed(() => dataScope.value === 'all' && permissions.value.has(demoPermissions.read))
const demoPath = '/demo/products'

const shellRuntime = {
  async loadIdentity(request?: { readonly signal?: AbortSignal }) {
    const identity = await runtime.loadIdentity(request)
    if (identity.kind === 'unauthenticated') {
      permissions.value = new Set()
      dataScope.value = null
    } else {
      permissions.value = new Set(identity.permissions)
      dataScope.value = identity.dataScope ?? null
    }
    await verifyNativeBoundary(identity.kind)
    return identity
  },
  loadNavigation: (request?: { readonly signal?: AbortSignal }) => runtime.loadNavigation(request)
}
const navigator = createShellNavigator(shellRuntime, {
  setLoading(loading) {
    if (loading) view.value = 'loading'
  },
  commit(_path, state) {
    if (state.kind === 'authenticated') view.value = 'workspace'
    else if (state.kind === 'unauthenticated') view.value = 'login'
    else if (state.kind === 'adapter-failed') view.value = 'unavailable'
    else view.value = 'forbidden'
  }
})
const refreshIdentity = () => navigator.navigate(demoPath)

const verifyNativeBoundary = async (stage: 'startup' | 'unauthenticated' | 'authenticated') => {
  if (!nativeE2E) return
  nativeBoundaryStage.value = ''
  try {
    const locationText = window.location.href.toLowerCase()
    const exposedText = document.body.textContent ?? ''
    const indexedDatabases = typeof window.indexedDB.databases === 'function' ? await window.indexedDB.databases() : []
    const cacheNames = 'caches' in window ? await window.caches.keys() : []
    const opaqueMaterial = exposedText.split(/\s+/).some(value => /^[A-Za-z0-9_-]{43,}$/.test(value) ||
      /^[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}$/.test(value))
    const safe = document.cookie === '' && window.localStorage.length === 0 &&
      window.sessionStorage.length === 0 && !locationText.includes('127.0.0.1') &&
      !locationText.startsWith('http:') && !locationText.startsWith('https:') &&
      !exposedText.includes('__Host-go-admin-session') && !exposedText.includes('csrfToken') &&
      !exposedText.includes('X-CSRF-Token') && !opaqueMaterial && indexedDatabases.length === 0 && cacheNames.length === 0
    if (safe) nativeBoundaryStage.value = stage
  } catch {
    nativeBoundaryStage.value = ''
  }
}

const nativeControl = async (action: 'scope-self' | 'scope-all' | 'permissions-off' | 'permissions-on' | 'session-revoke') => {
  if (!nativeE2E || busy.value) return
  busy.value = true
  nativeAuthorization.value = ''
  try {
    const result = await transport.request('/__desktop/test-control', 'POST', { action })
    if (result.status !== 204 || result.body !== null) throw new Error('native control failed')
    if (action === 'scope-self' || action === 'permissions-off') {
      try {
        await demoClient.list({ search: '', page: 1, pageSize: 20, sort: 'sku', direction: 'ascending' })
        throw new Error('authorization remained active')
      } catch (error) {
        if (!(error instanceof DemoRequestError) || error.category !== 'forbidden') throw error
        nativeAuthorization.value = 'E2E authorization denied'
      }
    }
    await refreshIdentity()
  } catch {
    nativeAuthorization.value = 'E2E control failed'
  } finally {
    busy.value = false
  }
}

const login = async () => {
  if (busy.value) return
  busy.value = true
  loginError.value = ''
  logoutError.value = ''
  try {
    await session.login(credentials.username.trim(), credentials.password)
    credentials.password = ''
    await refreshIdentity()
  } catch {
    credentials.password = ''
    loginError.value = '登录失败，请检查凭据。'
  } finally {
    busy.value = false
  }
}

const logout = async () => {
  if (busy.value) return
  busy.value = true
  logoutError.value = ''
  try {
    await session.logout()
    permissions.value = new Set()
    dataScope.value = null
    credentials.password = ''
    loginError.value = ''
    view.value = 'login'
  } catch {
    credentials.password = ''
    logoutError.value = '退出失败，请重试。'
  } finally {
    busy.value = false
  }
}

const requireSession = () => {
  permissions.value = new Set()
  dataScope.value = null
  view.value = 'login'
}
const forbid = () => {
  permissions.value = new Set()
  dataScope.value = null
  view.value = 'forbidden'
}

let stopped = false
const initialize = async () => {
  while (!stopped) {
    await refreshIdentity()
    if (view.value !== 'unavailable') return
    await new Promise(resolve => window.setTimeout(resolve, 100))
  }
}
onMounted(() => {
  void verifyNativeBoundary('startup')
  void initialize()
})
onUnmounted(() => { stopped = true; navigator.invalidate() })
</script>

<template>
  <main class="shell" :data-shell-state="view">
    <header class="shell__header">
      <strong>Go Admin Plus</strong>
      <div><span>Desktop</span><button v-if="view === 'workspace'" type="button" :disabled="busy" @click="logout">退出</button></div>
    </header>

    <aside v-if="nativeE2E" class="native-e2e" aria-live="polite">
      <span v-if="nativeBoundaryStage">E2E {{ nativeBoundaryStage }} boundary verified</span>
      <span v-if="nativeAuthorization">{{ nativeAuthorization }}</span>
      <button type="button" :disabled="busy" @click="nativeControl('scope-self')">E2E scope self</button>
      <button type="button" :disabled="busy" @click="nativeControl('scope-all')">E2E scope all</button>
      <button type="button" :disabled="busy" @click="nativeControl('permissions-off')">E2E permissions off</button>
      <button type="button" :disabled="busy" @click="nativeControl('permissions-on')">E2E permissions on</button>
      <button type="button" :disabled="busy" @click="nativeControl('session-revoke')">E2E revoke session</button>
    </aside>

    <section v-if="view === 'loading'" class="shell__state" aria-live="polite"><span class="spinner" aria-hidden="true" /><p>正在加载</p></section>

    <section v-else-if="view === 'login'" class="shell__state">
      <form class="login" @submit.prevent="login">
        <p class="eyebrow">SESSION</p><h1>登录</h1>
        <p v-if="loginError" role="alert">{{ loginError }}</p>
        <label>用户名<input v-model="credentials.username" name="username" autocomplete="username" minlength="3" maxlength="64" required></label>
        <label>密码<input v-model="credentials.password" name="password" type="password" autocomplete="current-password" minlength="12" maxlength="128" required></label>
        <button type="submit" :disabled="busy">登录</button>
      </form>
    </section>

    <section v-else-if="view === 'workspace' && canReadDemo" class="workspace">
      <nav><strong>业务管理</strong></nav>
      <div><p v-if="logoutError" role="alert">{{ logoutError }}</p><DemoProductsPage :controller="controller" @session-required="requireSession" @forbidden="forbid" /></div>
    </section>

    <section v-else-if="view === 'forbidden'" class="shell__state"><p class="eyebrow">403</p><h1>无权访问</h1><button type="button" @click="refreshIdentity">重新检查</button></section>
    <section v-else class="shell__state"><p class="eyebrow">RUNTIME</p><h1>服务暂不可用</h1><button type="button" @click="refreshIdentity">重试</button></section>
  </main>
</template>
