<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { filesPermissions, type FileMetadata, type FileQuery, type UploadCandidate } from '@go-admin-plus/domain-files'
import type { PlatformPort } from '@go-admin-plus/platform'
import { AppPage, EmptyState, Pagination, QueryBar, StatusTag, TableToolbar } from '@go-admin-plus/ui/components'
import type { FilesController } from './files-controller'

const props = defineProps<{ controller: FilesController; platform: PlatformPort }>()
const emit = defineEmits<{ sessionRequired: [] }>()
const revision = ref(0), search = ref(''), selectedUpload = ref<UploadCandidate | null>(null), fileInput = ref<HTMLInputElement | null>(null), localError = ref<string | null>(null)
const snapshot = computed(() => { void revision.value; return props.controller.list.snapshot() })
const failure = computed(() => { void revision.value; return props.controller.failure() })
const failureReference = computed(() => { void revision.value; return props.controller.failureTraceId() })
const projectionVisible = computed(() => { void revision.value; return props.controller.projectionVisible })
const blocked = computed(() => { void revision.value; return props.controller.busy || props.controller.pendingRepair })
const pageBusy = computed(() => snapshot.value.loading && !projectionVisible.value)
const uploadBlocked = computed(() => blocked.value || failure.value === 'quota' || failure.value === 'capacity')
const can = (permission: typeof filesPermissions[keyof typeof filesPermissions]) => { void revision.value; return props.controller.can(permission) }
const canRead = computed(() => can(filesPermissions.read)), canWrite = computed(() => can(filesPermissions.write)), canDelete = computed(() => can(filesPermissions.delete))
const selectedRows = computed(() => snapshot.value.rows.filter(row => snapshot.value.selectedKeys.includes(row.id)))
const failureMessage = computed(() => {
  const current = failure.value
  return current ? ({ relogin: '会话已失效，请重新登录。', forbidden: '没有执行该文件操作的权限。', validation: '请检查查询或上传内容。', content: '文件大小、类型或内容不符合上传规则。', 'not-found': '文件已不存在，请刷新列表。', conflict: '文件状态已变化，请刷新后重试。', quota: '当前账号文件容量已用尽。请下载或删除文件释放空间。', capacity: '存储空间处于保护状态，暂时停止上传。下载和删除仍可使用。', unavailable: '文件服务暂不可用。' } as const)[current] : ''
})
type FileSortKey = FileQuery['sort']
const currentSort = computed(() => snapshot.value.sort ?? { key: 'createdAt', direction: 'descending' as const })
const sortDirection = (key: FileSortKey) => currentSort.value.key === key ? currentSort.value.direction : 'none'
const sortMarker = (key: FileSortKey) => currentSort.value.key === key ? currentSort.value.direction === 'ascending' ? '↑' : '↓' : ''
const formatDate = (value: string) => new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'medium' }).format(new Date(value))
const formatBytes = (value: number) => value < 1024 ? `${value} B` : value < 1024 * 1024 ? `${(value / 1024).toFixed(1)} KB` : `${(value / 1024 / 1024).toFixed(1)} MB`
const settle = async (operation: () => Promise<unknown>) => { try { await operation() } catch {} finally { revision.value += 1; if (props.controller.failure() === 'relogin') emit('sessionRequired') } }
const refresh = () => settle(() => props.controller.pendingRepair ? props.controller.repairProjection() : props.controller.list.refresh())
const searchFiles = () => settle(() => props.controller.list.search({ search: search.value }))
const resetSearch = () => { search.value = ''; void settle(() => props.controller.list.reset()) }
const sortBy = (key: FileSortKey) => settle(() => props.controller.list.setSort({ key, direction: currentSort.value.key === key && currentSort.value.direction === 'ascending' ? 'descending' : 'ascending' }))
const chooseFile = (event: Event) => { const file = (event.target as HTMLInputElement).files?.[0]; selectedUpload.value = file ? { name: file.name, type: file.type, size: file.size, body: file } : null; localError.value = null }
const chooseHostFile = async () => {
  localError.value = null
  if (!props.platform.listCapabilities().has('file-open')) { localError.value = '当前运行环境不支持选择文件'; return }
  try { const file = await props.platform.pickFile(); selectedUpload.value = file ? { name: file.name, type: file.mediaType, size: file.bytes.length, body: new Blob([file.bytes.slice().buffer], { type: file.mediaType }) } : null }
  catch { selectedUpload.value = null; localError.value = '选择文件失败' }
}
const clearUpload = () => { selectedUpload.value = null; localError.value = null; if (fileInput.value) fileInput.value.value = '' }
const upload = async () => {
  if (!selectedUpload.value) { localError.value = '请选择文件'; return }
  const result = await props.controller.upload(selectedUpload.value)
  if (result === 'invalid') localError.value = '不支持该文件或文件不符合上传规则'
  if (result === 'completed') { props.controller.takeCompletion(); clearUpload() }
  revision.value += 1
  if (props.controller.failure() === 'relogin') emit('sessionRequired')
}
const remove = (rows: ReadonlyArray<FileMetadata>) => settle(async () => { if (await props.controller.remove(rows) === 'completed') props.controller.takeCompletion() })
const toggle = (row: FileMetadata, checked: boolean) => { const rows = checked ? [...selectedRows.value, row] : selectedRows.value.filter(candidate => candidate.id !== row.id); props.controller.list.select(rows); revision.value += 1 }
const download = async (row: FileMetadata) => {
  const blob = await props.controller.download(row); revision.value += 1
  if (!blob) return
  if (props.platform.runtime === 'desktop') {
    if (!props.platform.listCapabilities().has('file-save')) { localError.value = '当前运行环境不支持保存文件'; return }
    try { await props.platform.saveFile({ name: row.originalName, mediaType: row.mediaType, bytes: new Uint8Array(await blob.arrayBuffer()) }) } catch { localError.value = '保存文件失败' }
    return
  }
  const url = URL.createObjectURL(blob)
  try { const anchor = document.createElement('a'); anchor.href = url; anchor.download = row.originalName; anchor.click() } finally { URL.revokeObjectURL(url) }
}
onMounted(() => { void settle(() => props.controller.list.refresh()) })
</script>

<template>
  <AppPage title="文件管理" description="上传、下载和释放当前账号的文件空间" :busy="pageBusy">
    <template #actions><StatusTag :tone="failure === 'quota' || failure === 'capacity' ? 'warning' : 'info'" :label="`${snapshot.total} 个文件`" /></template>
    <p v-if="failure" class="page-alert" role="alert" :data-failure="failure">{{ failureMessage }}<span v-if="failureReference"> 参考编号：{{ failureReference }}</span></p>
    <template v-if="projectionVisible && canRead">
      <QueryBar :busy="blocked" :reset-disabled="!search" @search="searchFiles" @reset="resetSearch"><label>文件名称<input v-model="search" name="search" maxlength="100" placeholder="请输入文件名称"></label></QueryBar>
      <section v-if="canWrite" class="upload-panel" data-testid="files-upload" :aria-disabled="uploadBlocked">
        <div><strong>上传文件</strong><p>支持 PDF、JPG、PNG、TXT，单个文件不超过 10 MB</p></div>
        <label v-if="platform.runtime === 'web'">选择文件<input ref="fileInput" name="file" type="file" accept=".pdf,.jpg,.jpeg,.png,.txt,application/pdf,image/jpeg,image/png,text/plain" :disabled="uploadBlocked" @change="chooseFile"></label>
        <div v-else class="host-picker"><button type="button" :disabled="uploadBlocked" @click="chooseHostFile">选择文件</button><span>{{ selectedUpload?.name ?? '未选择文件' }}</span></div>
        <button type="button" :disabled="uploadBlocked || !selectedUpload" @click="upload">上传</button><p v-if="localError" role="alert">{{ localError }}</p>
      </section>
      <TableToolbar :selected-count="selectedRows.length" :busy="blocked" @refresh="refresh"><button v-if="canDelete" type="button" data-testid="files-delete-selected" :disabled="blocked || selectedRows.length === 0" @click="remove(selectedRows)">批量删除</button></TableToolbar>
      <div class="table-scroll" role="region" aria-label="文件列表">
        <table v-if="snapshot.rows.length > 0"><thead><tr><th v-if="canDelete" scope="col">选择</th><th scope="col" :aria-sort="sortDirection('name')"><button type="button" :disabled="blocked" @click="sortBy('name')">文件名称 {{ sortMarker('name') }}</button></th><th scope="col">类型</th><th scope="col" :aria-sort="sortDirection('sizeBytes')"><button type="button" :disabled="blocked" @click="sortBy('sizeBytes')">大小 {{ sortMarker('sizeBytes') }}</button></th><th scope="col" :aria-sort="sortDirection('createdAt')"><button type="button" :disabled="blocked" @click="sortBy('createdAt')">创建时间 {{ sortMarker('createdAt') }}</button></th><th scope="col">操作</th></tr></thead><tbody><tr v-for="row in snapshot.rows" :key="row.id" :data-file-id="row.id"><td v-if="canDelete"><input type="checkbox" :checked="snapshot.selectedKeys.includes(row.id)" :aria-label="`选择 ${row.originalName}`" :disabled="blocked" @change="toggle(row, ($event.target as HTMLInputElement).checked)"></td><td>{{ row.originalName }}</td><td>{{ row.mediaType }}</td><td>{{ formatBytes(row.sizeBytes) }}</td><td>{{ formatDate(row.createdAt) }}</td><td class="row-actions"><button type="button" :disabled="blocked" @click="download(row)">下载</button><button v-if="canDelete" type="button" :disabled="blocked" @click="remove([row])">删除</button></td></tr></tbody></table>
        <EmptyState v-else title="暂无文件" />
      </div>
      <Pagination :page="snapshot.page" :page-size="snapshot.pageSize" :total="snapshot.total" :disabled="blocked" @update:page="settle(() => controller.list.setPage($event))" @update:page-size="settle(() => controller.list.setPageSize($event))" />
    </template>
    <EmptyState v-else-if="!pageBusy" :title="failureMessage || '暂无可用文件数据'" action-label="重试" @action="refresh" />
  </AppPage>
</template>

<style scoped>
.page-alert{margin:0;padding:10px 12px;border-left:3px solid var(--ga-danger);background:var(--ga-danger-soft)}.upload-panel{display:grid;grid-template-columns:minmax(220px,1fr) minmax(220px,1fr) auto;align-items:end;gap:12px;padding:14px;border:1px solid var(--ga-border-light);background:var(--ga-bg-subtle)}.upload-panel p{margin:4px 0 0;color:var(--ga-text-2)}label{display:grid;gap:4px}.host-picker{display:flex;align-items:center;gap:8px;min-width:0}.host-picker span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.table-scroll{min-width:0;overflow:auto}.row-actions{display:flex;gap:8px;white-space:nowrap}@media(max-width:760px){.upload-panel{grid-template-columns:1fr;align-items:stretch}}
</style>
