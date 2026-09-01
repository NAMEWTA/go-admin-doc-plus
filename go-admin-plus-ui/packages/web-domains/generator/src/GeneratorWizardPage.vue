<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { ColumnDraft } from '@go-admin-plus/domain-generator'
import { AppPage, EmptyState, Pagination, QueryBar, StatusTag, TableToolbar } from '@go-admin-plus/ui/components'
import type { GeneratorController, WizardStep } from './generator-controller'

const props = defineProps<{ controller: GeneratorController }>()
const emit = defineEmits<{ sessionRequired: []; forbidden: [] }>()
const revision = ref(0), search = ref(''), selectedFile = ref(0), confirmed = ref(false)
const snapshot = computed(() => { void revision.value; return props.controller.tables.snapshot() })
const step = computed(() => { void revision.value; return props.controller.step })
const draft = computed(() => { void revision.value; return props.controller.draft })
const preview = computed(() => { void revision.value; return props.controller.previewValue })
const failure = computed(() => { void revision.value; return props.controller.failure() })
const failureReference = computed(() => { void revision.value; return props.controller.failureTraceId() })
const pageBusy = computed(() => props.controller.busy && !props.controller.projectionVisible)
const stepOrder: WizardStep[] = ['source', 'configure', 'preview', 'complete']
const stepLabel: Record<WizardStep, string> = { source: '选择数据表', configure: '配置字段', preview: '预览与门禁', complete: '生成完成' }
const failureMessage = computed(() => {
  const current = failure.value
  return current ? ({ relogin: '会话已失效，请重新登录。', forbidden: '没有执行该生成操作的权限。', validation: '请检查生成配置和确认项。', 'not-found': '数据表或预览令牌已失效，请重新开始。', conflict: '生成目录状态已变化，请重新预览。', gate: '生成输出未通过必需门禁。预览已保留，请修改配置后重新预览。', unavailable: '代码生成服务暂不可用。' } as const)[current] : ''
})
const stepTone = (value: WizardStep) => stepOrder.indexOf(value) < stepOrder.indexOf(step.value) ? 'success' : value === step.value ? failure.value === 'gate' ? 'danger' : 'info' : 'neutral'
const settle = async (operation: () => Promise<unknown>) => { try { await operation() } catch {} finally { revision.value += 1; if (props.controller.failure() === 'relogin') emit('sessionRequired'); if (props.controller.failure() === 'forbidden') emit('forbidden') } }
const refresh = () => settle(() => props.controller.tables.refresh())
const searchTables = () => settle(() => props.controller.tables.search({ search: search.value }))
const resetSearch = () => { search.value = ''; void settle(() => props.controller.tables.reset()) }
const updateNames = (key: 'module'|'entity'|'plural', value: string) => { if (!draft.value) return; props.controller.setNames(key === 'module' ? value : draft.value.module, key === 'entity' ? value : draft.value.entity, key === 'plural' ? value : draft.value.plural); revision.value += 1 }
const updateColumn = (column: ColumnDraft, changes: Partial<Omit<ColumnDraft, 'name'>>) => { props.controller.configureColumn(column.name, changes); revision.value += 1 }
const restart = () => { props.controller.reset(); selectedFile.value = 0; confirmed.value = false; revision.value += 1 }
const createPreview = () => settle(async () => { if (await props.controller.createPreview() === 'completed') { selectedFile.value = 0; confirmed.value = false } })
const write = () => settle(async () => { const result = await props.controller.confirmWrite(confirmed.value); if (result !== 'completed') confirmed.value = false })
const returnToConfiguration = () => { props.controller.returnToConfiguration(); confirmed.value = false; selectedFile.value = 0; revision.value += 1 }
onMounted(() => { void refresh() })
</script>

<template>
  <AppPage title="代码生成" description="从数据表配置、预览并写入当前架构模块" :busy="pageBusy">
    <template #actions><button v-if="step !== 'source'" type="button" :disabled="controller.busy" @click="restart">重新开始</button></template>
    <ol class="steps" aria-label="生成步骤"><li v-for="item in stepOrder" :key="item"><StatusTag :tone="stepTone(item)" :label="stepLabel[item]" /></li></ol>
    <p v-if="failure" class="page-alert" role="alert" :data-failure="failure">{{ failureMessage }}<span v-if="failureReference"> 参考编号：{{ failureReference }}</span></p>
    <section v-if="step === 'source' && controller.projectionVisible" class="wizard-panel">
      <QueryBar :busy="controller.busy" :reset-disabled="!search" @search="searchTables" @reset="resetSearch"><label>数据表<input v-model="search" maxlength="100" placeholder="请输入 schema 或表名"></label></QueryBar>
      <TableToolbar :busy="controller.busy" @refresh="refresh"><span>共 {{ snapshot.total }} 张表</span></TableToolbar>
      <div class="table-scroll"><table v-if="snapshot.rows.length"><thead><tr><th>Schema</th><th>表名</th><th>操作</th></tr></thead><tbody><tr v-for="table in snapshot.rows" :key="`${table.schema}.${table.name}`"><td>{{ table.schema }}</td><td>{{ table.name }}</td><td><button type="button" :disabled="controller.busy" @click="settle(() => controller.select(table))">配置</button></td></tr></tbody></table><EmptyState v-else title="未找到可生成的数据表" /></div>
      <Pagination data-testid="generator-table-pagination" :page="snapshot.page" :page-size="snapshot.pageSize" :total="snapshot.total" :disabled="controller.busy" @update:page="settle(() => controller.tables.setPage($event))" @update:page-size="settle(() => controller.tables.setPageSize($event))" />
    </section>
    <form v-else-if="step === 'configure' && draft" class="wizard-panel" @submit.prevent="createPreview">
      <div class="names"><label>模块<input name="module" :value="draft.module" required minlength="2" maxlength="32" pattern="[a-z][a-z0-9]{1,31}" @input="updateNames('module', ($event.target as HTMLInputElement).value)"></label><label>实体<input name="entity" :value="draft.entity" required maxlength="64" pattern="[A-Z][A-Za-z0-9]{0,63}" @input="updateNames('entity', ($event.target as HTMLInputElement).value)"></label><label>复数路由<input name="plural" :value="draft.plural" required minlength="2" maxlength="64" pattern="[a-z][a-z0-9-]{1,63}" @input="updateNames('plural', ($event.target as HTMLInputElement).value)"></label></div>
      <div class="table-scroll"><table><thead><tr><th>数据列</th><th>字段</th><th>生成</th><th>搜索</th><th>排序</th></tr></thead><tbody><tr v-for="column in draft.columns" :key="column.name"><td>{{ column.name }}</td><td><input :name="`field-${column.name}`" :value="column.field" required maxlength="64" pattern="[A-Z][A-Za-z0-9]{0,63}" @input="updateColumn(column, { field: ($event.target as HTMLInputElement).value })"></td><td><input type="checkbox" :checked="column.include" :aria-label="`生成 ${column.name}`" @change="updateColumn(column, { include: ($event.target as HTMLInputElement).checked })"></td><td><input type="checkbox" :checked="column.searchable" :aria-label="`搜索 ${column.name}`" :disabled="!column.include" @change="updateColumn(column, { searchable: ($event.target as HTMLInputElement).checked })"></td><td><input type="checkbox" :checked="column.sortable" :aria-label="`排序 ${column.name}`" :disabled="!column.include" @change="updateColumn(column, { sortable: ($event.target as HTMLInputElement).checked })"></td></tr></tbody></table></div>
      <footer><button type="submit" :disabled="controller.busy">生成预览</button></footer>
    </form>
    <section v-else-if="step === 'preview' && preview" class="wizard-panel preview-grid">
      <header class="gate-status"><div><h2>输出预览</h2><p>{{ preview.files.length }} 个文件，预览有效至 {{ new Date(preview.expiresAt).toLocaleString() }}</p></div><StatusTag :tone="failure === 'gate' ? 'danger' : 'warning'" :label="failure === 'gate' ? '必需门禁未通过' : '等待必需门禁'" /></header>
      <nav aria-label="生成文件"><button v-for="(file, index) in preview.files" :key="file.path" type="button" :aria-current="selectedFile === index ? 'page' : undefined" @click="selectedFile = index">{{ file.path }}</button></nav>
      <pre aria-label="生成文件内容"><code>{{ preview.files[selectedFile]?.content }}</code></pre>
      <footer><button type="button" :disabled="controller.busy" @click="returnToConfiguration">返回配置</button><label class="confirmation"><input v-model="confirmed" type="checkbox">我已检查隔离目录中的生成结果</label><button type="button" :disabled="controller.busy || !confirmed" @click="write">执行门禁并写入</button></footer>
    </section>
    <section v-else-if="step === 'complete' && controller.result" class="wizard-panel complete"><StatusTag tone="success" label="必需门禁已通过" /><h2>生成完成</h2><p class="directory">{{ controller.result.directory }}</p><ul><li v-for="file in controller.result.files" :key="file">{{ file }}</li></ul></section>
    <EmptyState v-else-if="step === 'source' && !pageBusy" :title="failureMessage || '暂无数据表'" action-label="重试" @action="refresh" />
  </AppPage>
</template>

<style scoped>
.steps{display:flex;gap:8px;flex-wrap:wrap;margin:0;padding:0;list-style:none}.page-alert{margin:0;padding:10px 12px;border-left:3px solid var(--ga-danger);background:var(--ga-danger-soft)}.wizard-panel{display:grid;gap:14px}.names{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:12px}label{display:grid;gap:4px}.table-scroll{min-width:0;overflow:auto}footer,.gate-status{display:flex;align-items:center;justify-content:space-between;gap:12px}.gate-status h2,.gate-status p,.complete h2,.complete p{margin:0}.preview-grid{grid-template-columns:minmax(200px,30%) minmax(0,1fr)}.preview-grid .gate-status,.preview-grid footer{grid-column:1/-1}.preview-grid nav{display:grid;align-content:start;gap:4px;max-height:520px;overflow:auto}.preview-grid nav button{text-align:left;overflow-wrap:anywhere}.preview-grid pre{margin:0;padding:12px;max-height:520px;overflow:auto;background:var(--ga-bg-subtle);border:1px solid var(--ga-border-light);white-space:pre}.confirmation{display:flex;align-items:center;grid-template-columns:auto 1fr}.directory{overflow-wrap:anywhere}@media(max-width:760px){.names,.preview-grid{grid-template-columns:1fr}.preview-grid .gate-status,.preview-grid footer{grid-column:auto}.gate-status,footer{align-items:stretch;flex-direction:column}}
</style>
