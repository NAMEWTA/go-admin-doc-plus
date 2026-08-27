<script setup lang="ts">
import { createWebRuntime } from '@go-admin/adapter-browser'
import { ProductWorkspace } from '@go-admin/app-shell/product'

const runtime = createWebRuntime()
</script>

<template>
  <ProductWorkspace host="web" :runtime="runtime" />
</template>
