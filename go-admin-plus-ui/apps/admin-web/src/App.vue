<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'

import { resolveShellState } from '@go-admin/app-shell'
import type { ShellState } from '@go-admin/app-shell'

import { createWebRuntime } from './runtime/web-runtime'

const runtime = createWebRuntime()
const state = ref<ShellState>()
const loading = ref(true)

const currentPath = () => window.location.pathname

const navigate = async () => {
  loading.value = true
  const resolved = await resolveShellState(runtime, currentPath())
  if (resolved.kind === 'unauthenticated' && currentPath() !== resolved.redirectTo) {
    window.history.replaceState({}, '', resolved.redirectTo)
  }
  state.value = resolved
  loading.value = false
}

onMounted(() => {
  window.addEventListener('popstate', navigate)
  void navigate()
})
onUnmounted(() => window.removeEventListener('popstate', navigate))
</script>

<template>
  <main class="shell" :data-shell-state="loading ? 'loading' : state?.kind">
    <header class="shell__header">
      <a class="shell__brand" href="/">Go Admin Plus</a>
    </header>

    <section v-if="loading" class="shell__state" aria-live="polite">
      <span class="shell__spinner" aria-hidden="true" />
      <p>正在加载</p>
    </section>

    <section v-else-if="state?.kind === 'authenticated'" class="shell__workspace">
      <nav class="shell__nav" aria-label="主导航">
        <strong>工作台</strong>
      </nav>
      <div class="shell__content">
        <p class="shell__eyebrow">ADMIN WEB</p>
        <h1>管理工作台</h1>
      </div>
    </section>

    <section v-else-if="state?.kind === 'unauthenticated'" class="shell__state">
      <p class="shell__code">SESSION</p>
      <h1>登录</h1>
      <p>请登录以继续。</p>
    </section>

    <section v-else-if="state?.kind === 'unauthorized'" class="shell__state">
      <p class="shell__code">403</p>
      <h1>无权访问</h1>
      <a href="/">返回工作台</a>
    </section>

    <section v-else-if="state?.kind === 'not-found'" class="shell__state">
      <p class="shell__code">404</p>
      <h1>页面不存在</h1>
      <a href="/">返回工作台</a>
    </section>

    <section v-else class="shell__state">
      <p class="shell__code">RUNTIME</p>
      <h1>服务暂不可用</h1>
      <button type="button" @click="navigate">重试</button>
    </section>
  </main>
</template>
