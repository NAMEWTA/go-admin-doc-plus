<script setup lang="ts">
import { createDesktopFetch, createDesktopPlatform, createDesktopRuntime, createDesktopSessionClient } from '@go-admin-plus/adapter-desktop'
import { ProductWorkspace } from '@go-admin-plus/app-shell/product'

const runtime = createDesktopRuntime()
const platform = createDesktopPlatform()
const fetcher = createDesktopFetch()
const session = createDesktopSessionClient()
</script>

<template>
  <ProductWorkspace host="desktop" :runtime="runtime" :platform="platform" :fetcher="fetcher" :session-client="session" />
</template>
