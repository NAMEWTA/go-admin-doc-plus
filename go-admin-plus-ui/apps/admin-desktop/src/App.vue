<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'

import { createDesktopDemoClient, createDesktopRuntime, createDesktopSession } from '@go-admin/adapter-desktop'
import { demoPermissions, type DemoPermissionCode } from '@go-admin/domain-demo'
import { createDemoController, DemoProductsPage } from '@go-admin/web-domain-demo'

type View = 'loading' | 'login' | 'workspace' | 'forbidden' | 'unavailable'

const runtime = createDesktopRuntime()
const session = createDesktopSession()
const permissions = ref<ReadonlySet<string>>(new Set())
const dataScope = ref<'self' | 'all' | null>(null)
const view = ref<View>('loading')
const busy = ref(false)
const loginError = ref('')
const credentials = reactive({ username: '', password: '' })
const capabilities = { can: (permission: DemoPermissionCode) => permissions.value.has(permission) }
const controller = createDemoController(
  createDesktopDemoClient(),
  async count => window.confirm(`Delete ${count} product${count === 1 ? '' : 's'}?`),
  capabilities
)
const canReadDemo = computed(() => dataScope.value === 'all' && permissions.value.has(demoPermissions.read))

const refreshIdentity = async () => {
  view.value = 'loading'
  try {
    const identity = await runtime.loadIdentity()
    if (identity.kind === 'unauthenticated') {
      permissions.value = new Set()
      dataScope.value = null
      view.value = 'login'
      return
    }
    permissions.value = new Set(identity.permissions)
    dataScope.value = identity.dataScope ?? null
    view.value = dataScope.value === 'all' && permissions.value.has(demoPermissions.read) ? 'workspace' : 'forbidden'
  } catch {
    permissions.value = new Set()
    dataScope.value = null
    view.value = 'unavailable'
  }
}

const login = async () => {
  if (busy.value) return
  busy.value = true
  loginError.value = ''
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
  try {
    const result = await session.logout()
    permissions.value = new Set()
    dataScope.value = null
    credentials.password = ''
    loginError.value = result.remoteRevoked ? '' : '本地凭据已清除，远端会话将按策略失效。'
    view.value = 'login'
  } catch {
    permissions.value = new Set()
    dataScope.value = null
    credentials.password = ''
    view.value = 'unavailable'
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

onMounted(() => { void refreshIdentity() })
</script>

<template>
  <main class="shell" :data-shell-state="view">
    <header class="shell__header">
      <strong>Go Admin Plus</strong>
      <div><span>Desktop</span><button v-if="view === 'workspace'" type="button" :disabled="busy" @click="logout">退出</button></div>
    </header>

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
      <nav><strong>业务管理</strong><span>产品</span></nav>
      <div><DemoProductsPage :controller="controller" @session-required="requireSession" @forbidden="forbid" /></div>
    </section>

    <section v-else-if="view === 'forbidden'" class="shell__state"><p class="eyebrow">403</p><h1>无权访问</h1><button type="button" @click="refreshIdentity">重新检查</button></section>
    <section v-else class="shell__state"><p class="eyebrow">RUNTIME</p><h1>服务暂不可用</h1><button type="button" @click="refreshIdentity">重试</button></section>
  </main>
</template>
