<script setup lang="ts">
import { createBrowserSessionFetch, createWebPlatform, createWebRuntime } from '@go-admin-plus/adapter-browser'
import { ProductWorkspace } from '@go-admin-plus/app-shell/product'

const fetcher = createBrowserSessionFetch()
const runtime = createWebRuntime(fetcher)
const platform = createWebPlatform()
</script>

<template>
  <ProductWorkspace host="web" :runtime="runtime" :platform="platform" :fetcher="fetcher" />
</template>
