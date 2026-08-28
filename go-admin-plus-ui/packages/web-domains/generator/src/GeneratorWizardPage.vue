<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { ColumnDraft } from '@go-admin/domain-generator'
import type { GeneratorController } from './generator-controller'

const props = defineProps<{ controller: GeneratorController }>()
const emit = defineEmits<{ sessionRequired: []; forbidden: [] }>()
const revision = ref(0), search = ref(''), selectedFile = ref(0), confirmed = ref(false)
const snapshot = computed(() => { void revision.value; return props.controller.tables.snapshot() })
const step = computed(() => { void revision.value; return props.controller.step })
const draft = computed(() => { void revision.value; return props.controller.draft })
const preview = computed(() => { void revision.value; return props.controller.previewValue })
const failure = computed(() => { void revision.value; return props.controller.failure() })
const settle = async (operation: () => Promise<unknown>) => {
  try { await operation() } catch { /* controller retains the stable failure */ }
  finally { revision.value += 1; if (props.controller.failure() === 'relogin') emit('sessionRequired'); if (props.controller.failure() === 'forbidden') emit('forbidden') }
}
const updateNames = (key: 'module'|'entity'|'plural', value: string) => {
  if (!draft.value) return
  props.controller.setNames(key === 'module' ? value : draft.value.module, key === 'entity' ? value : draft.value.entity, key === 'plural' ? value : draft.value.plural); revision.value += 1
}
const updateColumn = (column: ColumnDraft, changes: Partial<Omit<ColumnDraft, 'name'>>) => { props.controller.configureColumn(column.name, changes); revision.value += 1 }
const write = () => settle(async () => { const result = await props.controller.confirmWrite(confirmed.value); if (result !== 'completed') confirmed.value = false })
onMounted(() => { void settle(() => props.controller.tables.refresh()) })
</script>

<template>
  <section class="generator-wizard" aria-labelledby="generator-title">
    <header><div><h1 id="generator-title">代码生成</h1><p>{{ step === 'source' ? '选择数据表' : step === 'configure' ? '配置字段' : step === 'preview' ? '预览代码' : '生成完成' }}</p></div><button v-if="step !== 'source'" type="button" :disabled="controller.busy" @click="controller.reset(); confirmed = false; revision += 1">重新开始</button></header>
    <p v-if="failure" role="alert" :data-failure="failure">代码生成操作未能完成，请检查配置后重试。</p>

    <div v-if="step === 'source' && controller.projectionVisible" class="wizard-panel">
      <form class="toolbar" @submit.prevent="settle(() => controller.tables.search({ search }))"><label>数据表<input v-model="search" placeholder="请输入表名"></label><button :disabled="controller.busy">搜索</button><button type="button" :disabled="controller.busy" @click="search = ''; settle(() => controller.tables.reset())">重置</button></form>
      <table><thead><tr><th>Schema</th><th>表名</th><th>操作</th></tr></thead><tbody><tr v-for="table in snapshot.rows" :key="`${table.schema}.${table.name}`"><td>{{ table.schema }}</td><td>{{ table.name }}</td><td><button type="button" :disabled="controller.busy" @click="settle(() => controller.select(table))">配置</button></td></tr></tbody></table>
    </div>

    <form v-else-if="step === 'configure' && draft" class="wizard-panel" @submit.prevent="settle(() => controller.createPreview())">
      <div class="names"><label>模块<input :value="draft.module" @input="updateNames('module', ($event.target as HTMLInputElement).value)"></label><label>实体<input :value="draft.entity" @input="updateNames('entity', ($event.target as HTMLInputElement).value)"></label><label>复数路由<input :value="draft.plural" @input="updateNames('plural', ($event.target as HTMLInputElement).value)"></label></div>
      <table><thead><tr><th>数据列</th><th>字段</th><th>生成</th><th>搜索</th><th>排序</th></tr></thead><tbody><tr v-for="column in draft.columns" :key="column.name"><td>{{ column.name }}</td><td><input :value="column.field" @input="updateColumn(column, { field: ($event.target as HTMLInputElement).value })"></td><td><input type="checkbox" :checked="column.include" @change="updateColumn(column, { include: ($event.target as HTMLInputElement).checked })"></td><td><input type="checkbox" :checked="column.searchable" :disabled="!column.include" @change="updateColumn(column, { searchable: ($event.target as HTMLInputElement).checked })"></td><td><input type="checkbox" :checked="column.sortable" :disabled="!column.include" @change="updateColumn(column, { sortable: ($event.target as HTMLInputElement).checked })"></td></tr></tbody></table>
      <footer><button type="submit" :disabled="controller.busy">预览</button></footer>
    </form>

    <div v-else-if="step === 'preview' && preview" class="wizard-panel preview-grid">
      <nav aria-label="生成文件"><button v-for="(file, index) in preview.files" :key="file.path" type="button" :aria-current="selectedFile === index ? 'page' : undefined" @click="selectedFile = index">{{ file.path }}</button></nav>
      <pre><code>{{ preview.files[selectedFile]?.content }}</code></pre>
      <footer><label><input v-model="confirmed" type="checkbox">我已检查隔离目录中的生成结果</label><button type="button" :disabled="controller.busy || !confirmed" @click="write">生成代码</button></footer>
    </div>

    <div v-else-if="step === 'complete' && controller.result" class="wizard-panel"><h2>生成完成</h2><p>{{ controller.result.directory }}</p><ul><li v-for="file in controller.result.files" :key="file">{{ file }}</li></ul></div>
  </section>
</template>

<style scoped>
.generator-wizard { display: grid; gap: 16px; color: var(--ga-text-1); }
header, footer, .names { display: flex; align-items: end; justify-content: space-between; gap: 12px; flex-wrap: wrap; }
h1, h2, p { margin: 0; }
.wizard-panel { display: grid; gap: 14px; }
label { display: grid; gap: 4px; }
.preview-grid { grid-template-columns: minmax(220px, 32%) minmax(0, 1fr); }
.preview-grid nav { display: grid; align-content: start; gap: 4px; max-height: 520px; overflow: auto; }
.preview-grid nav button { text-align: left; overflow-wrap: anywhere; }
.preview-grid pre { margin: 0; padding: 12px; max-height: 520px; overflow: auto; background: var(--ga-bg-subtle); border: 1px solid var(--ga-border-light); }
.preview-grid footer { grid-column: 1 / -1; }
@media (max-width: 720px) { .preview-grid { grid-template-columns: 1fr; } .preview-grid footer { grid-column: auto; } }
</style>
