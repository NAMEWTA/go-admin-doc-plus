<script setup lang="ts">
import { createDesktopPlatform } from '@go-admin-plus/adapter-desktop'
import { ProductWorkspace, type SessionClient } from '@go-admin-plus/app-shell/product'
import type { ShellRuntimePort } from '@go-admin-plus/platform'

defineProps<{
  runtime: ShellRuntimePort
  fetcher: typeof globalThis.fetch
  session: SessionClient
}>()
const platform = createDesktopPlatform()
</script>

<template>
  <ProductWorkspace host="desktop" :runtime="runtime" :platform="platform" :fetcher="fetcher" :session-client="session" />
</template>
