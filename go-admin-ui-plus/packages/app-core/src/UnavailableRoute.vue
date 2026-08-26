<template>
  <main class="route-unavailable" role="alert">
    <h1>页面不可用</h1>
    <p>菜单配置无效或该页面尚未安装。</p>
    <code v-if="identity">{{ identity }}</code>
  </main>
</template>

<script setup lang="ts">
defineOptions({ name: 'UnavailableDomainRoute' })

defineProps<{
  identity?: string
}>()
</script>

<style scoped>
.route-unavailable {
  min-height: 320px;
  display: grid;
  place-content: center;
  justify-items: center;
  gap: 8px;
  color: var(--el-text-color-regular);
}

h1,
p {
  margin: 0;
}

h1 {
  font-size: 24px;
}

code {
  color: var(--el-text-color-secondary);
}
</style>
