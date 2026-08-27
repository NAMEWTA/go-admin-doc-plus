<script setup lang="ts">
import { createDesktopFetch, createDesktopRuntime, createDesktopSessionClient } from '@go-admin/adapter-desktop'
import { ProductWorkspace } from '@go-admin/app-shell/product'

const runtime = createDesktopRuntime()
const fetcher = createDesktopFetch()
const session = createDesktopSessionClient()
</script>

<template>
  <ProductWorkspace host="desktop" :runtime="runtime" :fetcher="fetcher" :session-client="session" />
</template>
