<script setup lang="ts">
import { createBrowserSessionFetch, createWebRuntime } from '@go-admin/adapter-browser'
import { ProductWorkspace } from '@go-admin/app-shell/product'

const fetcher = createBrowserSessionFetch()
const runtime = createWebRuntime(fetcher)
</script>

<template>
  <ProductWorkspace host="web" :runtime="runtime" :fetcher="fetcher" />
</template>
