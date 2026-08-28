<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { createDesktopFetch, createDesktopRuntime, createDesktopSessionClient, createDesktopTransport } from '@go-admin/adapter-desktop'
import { ProductWorkspace } from '@go-admin/app-shell/product'

const runtime = createDesktopRuntime()
const fetcher = createDesktopFetch()
const session = createDesktopSessionClient()
const nativeE2E = import.meta.env.VITE_GO_ADMIN_NATIVE_E2E === '1'
const nativeBoundary = ref('')
const nativeAuthorization = ref('')
const workspaceKey = ref(0)
const transport = nativeE2E ? createDesktopTransport() : null
let boundaryTimer: number | undefined
let nativeConfirm: typeof window.confirm | undefined

const verifyNativeBoundary = async () => {
  if (!nativeE2E) return
  try {
    const text = document.body.textContent ?? ''
    const location = new URL(window.location.href)
    const databases = typeof window.indexedDB.databases === 'function' ? await window.indexedDB.databases() : []
    const cacheNames = 'caches' in window ? await window.caches.keys() : []
    const opaqueMaterial = text.split(/\s+/).some(value => /^[A-Za-z0-9_-]{43}$/.test(value) ||
      /^[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}$/.test(value))
    const trustedLocation = location.protocol === 'tauri:' || location.hostname === 'tauri.localhost'
    const failures = [
      document.cookie !== '' ? 'cookie' : '',
      window.localStorage.length !== 0 ? 'local-storage' : '',
      window.sessionStorage.length !== 0 ? 'session-storage' : '',
      databases.length !== 0 ? 'indexed-db' : '',
      cacheNames.length !== 0 ? 'cache-storage' : '',
      !trustedLocation ? 'location' : '',
      text.includes('__Host-go-admin-session') || text.includes('csrfToken') || text.includes('X-CSRF-Token') ? 'protected-text' : '',
      opaqueMaterial ? 'opaque-text' : ''
    ].filter(Boolean)
    const safe = failures.length === 0
    nativeBoundary.value = safe && text.includes('Administrator')
      ? 'E2E authenticated boundary verified'
      : safe && text.includes('Sign in') ? 'E2E unauthenticated boundary verified'
        : `E2E boundary blocked: ${failures.join(',')}`
  } catch {
    nativeBoundary.value = ''
  }
}

const nativeControl = async (action: 'scope-self' | 'scope-all' | 'permissions-off' | 'permissions-on' | 'session-revoke') => {
  if (!transport) return
  nativeAuthorization.value = ''
  let stage = 'control-request'
  try {
    const result = await transport.request('/__desktop/test-control', 'POST', { action })
    stage = 'control-response'
    if (result.status !== 204 || result.body !== null) throw new Error('native control failed')
    if (action === 'scope-self' || action === 'scope-all') {
      stage = 'scope-request'
      const response = await fetcher('/api/demo/products?page=1&pageSize=20&sort=sku&direction=ascending')
      stage = 'scope-status'
      if (response.status !== 200) throw new Error('native scope request failed')
      stage = 'scope-body'
      const payload = await response.json() as { rows?: Array<{ sku?: string }> }
      const foreignVisible = payload.rows?.some(row => row.sku === 'E2E-FOREIGN') === true
      stage = 'scope-boundary'
      if (action === 'scope-self' ? foreignVisible : !foreignVisible) throw new Error('native scope boundary failed')
      nativeAuthorization.value = action === 'scope-self' ? 'E2E self scope enforced' : 'E2E all scope restored'
    }
    if (action === 'permissions-off') {
      stage = 'permission-request'
      const response = await fetcher('/api/demo/products?page=1&pageSize=20&sort=sku&direction=ascending')
      stage = 'permission-status'
      if (response.status !== 403) throw new Error('native authorization remained active')
      nativeAuthorization.value = 'E2E authorization denied'
    }
    workspaceKey.value += 1
  } catch {
    nativeAuthorization.value = `E2E control failed: ${stage}`
  }
}

onMounted(() => {
  if (!nativeE2E) return
  nativeConfirm = window.confirm
  window.confirm = () => true
  boundaryTimer = window.setInterval(() => { void verifyNativeBoundary() }, 100)
})
onUnmounted(() => {
  if (boundaryTimer !== undefined) window.clearInterval(boundaryTimer)
  if (nativeConfirm) window.confirm = nativeConfirm
})
</script>

<template>
  <aside v-if="nativeE2E" class="native-e2e" aria-live="polite">
    <span v-if="nativeBoundary">{{ nativeBoundary }}</span>
    <span v-if="nativeAuthorization">{{ nativeAuthorization }}</span>
    <button type="button" @click="nativeControl('scope-self')">E2E scope self</button>
    <button type="button" @click="nativeControl('scope-all')">E2E scope all</button>
    <button type="button" @click="nativeControl('permissions-off')">E2E permissions off</button>
    <button type="button" @click="nativeControl('permissions-on')">E2E permissions on</button>
    <button type="button" @click="nativeControl('session-revoke')">E2E revoke session</button>
  </aside>
  <ProductWorkspace :key="workspaceKey" host="desktop" :runtime="runtime" :fetcher="fetcher" :session-client="session" />
</template>

<style scoped>
.native-e2e { position: fixed; z-index: 10; right: 8px; bottom: 8px; display: flex; max-width: calc(100vw - 16px); flex-wrap: wrap; gap: 4px; padding: 6px; background: #fff; border: 1px solid #9ba6aa; }
.native-e2e span { width: 100%; font-size: 12px; }
.native-e2e button { min-height: 28px; padding: 2px 6px; font-size: 11px; }
</style>
