<script setup lang="ts">
import { createBrowserSessionFetch, createWebRuntime } from '@go-admin-plus/adapter-browser'
import { ProductWorkspace } from '@go-admin-plus/app-shell/product'

const fetcher = createBrowserSessionFetch()
const runtime = createWebRuntime(fetcher)
</script>

<template>
  <ProductWorkspace host="web" :runtime="runtime" :fetcher="fetcher" />
</template>
