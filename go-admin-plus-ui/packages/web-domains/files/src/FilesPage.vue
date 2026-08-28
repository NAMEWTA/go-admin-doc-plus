<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { filesPermissions, type FileMetadata, type FileQuery, type UploadCandidate } from '@go-admin-plus/domain-files'
import type { FilesController } from './files-controller'

const props = defineProps<{ controller: FilesController }>()
const emit = defineEmits<{ sessionRequired: [] }>()
const revision = ref(0)
const search = ref('')
const selectedUpload = ref<UploadCandidate | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const localError = ref<string | null>(null)
const settle = async (operation: () => Promise<unknown>) => {
  try { await operation() } catch { /* Stable failure state belongs to the controller. */ }
  finally {
    revision.value += 1
    if (props.controller.failure() === 'relogin') emit('sessionRequired')
  }
}
const can = (permission: typeof filesPermissions[keyof typeof filesPermissions]) => {
  void revision.value
  return props.controller.can(permission)
}
const snapshot = computed(() => { void revision.value; return props.controller.list.snapshot() })
const failure = computed(() => { void revision.value; return props.controller.failure() })
const projectionVisible = computed(() => { void revision.value; return props.controller.projectionVisible })
const blocked = computed(() => { void revision.value; return props.controller.busy || props.controller.pendingRepair })
const canRead = computed(() => can(filesPermissions.read))
const canWrite = computed(() => can(filesPermissions.write))
const canDelete = computed(() => can(filesPermissions.delete))
const selectedRows = computed(() => snapshot.value.rows.filter(row => snapshot.value.selectedKeys.includes(row.id)))
type FileSortKey = FileQuery['sort']
const currentSort = computed(() => snapshot.value.sort ?? { key: 'createdAt', direction: 'descending' as const })
const sortDirection = (key: FileSortKey) => currentSort.value.key === key ? currentSort.value.direction : 'none'
const sortMarker = (key: FileSortKey) => currentSort.value.key === key ? currentSort.value.direction === 'ascending' ? '↑' : '↓' : ''
const sortBy = (key: FileSortKey) => settle(() => props.controller.list.setSort({ key, direction: currentSort.value.key === key && currentSort.value.direction === 'ascending' ? 'descending' : 'ascending' }))
const formatDate = (value: string) => new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'medium' }).format(new Date(value))

const chooseFile = (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  selectedUpload.value = file ? { name: file.name, type: file.type, size: file.size, body: file } : null
  localError.value = null
}
const upload = async () => {
  if (!selectedUpload.value) { localError.value = '请选择文件'; return }
  const result = await props.controller.upload(selectedUpload.value)
  if (result === 'invalid') localError.value = '不支持该文件或文件不符合上传规则'
  if (result === 'completed') { props.controller.takeCompletion(); clearUpload() }
  revision.value += 1
  if (props.controller.failure() === 'relogin') emit('sessionRequired')
}
const repair = async () => {
  const result = await props.controller.repairProjection()
  if (result === 'completed' && props.controller.takeCompletion() === 'upload') clearUpload()
  revision.value += 1
}
const clearUpload = () => {
  selectedUpload.value = null
  localError.value = null
  if (fileInput.value) fileInput.value.value = ''
}
const remove = async (rows: ReadonlyArray<FileMetadata>) => settle(async () => {
  const result = await props.controller.remove(rows)
  if (result === 'completed') props.controller.takeCompletion()
})
const toggle = (row: FileMetadata, checked: boolean) => {
  const rows = checked ? [...selectedRows.value, row] : selectedRows.value.filter(candidate => candidate.id !== row.id)
  props.controller.list.select(rows)
  revision.value += 1
}
const download = async (row: FileMetadata) => {
  const blob = await props.controller.download(row)
  revision.value += 1
  if (!blob) return
  const url = URL.createObjectURL(blob)
  try {
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = row.originalName
    anchor.click()
  } finally { URL.revokeObjectURL(url) }
}
onMounted(() => settle(() => props.controller.list.refresh()))
</script>

<template>
  <section class="files-page" aria-labelledby="files-title">
    <header class="files-page__header">
      <div><h1 id="files-title">文件管理</h1><p>共 {{ snapshot.total }} 个文件</p></div>
      <button v-if="controller.pendingRepair" type="button" data-testid="files-repair" :disabled="controller.busy" @click="repair">重试刷新</button>
    </header>
    <p v-if="failure === 'relogin'" role="alert">会话已失效，请重新登录。</p>
    <p v-else-if="failure === 'forbidden'" role="alert">没有文件操作权限。</p>
    <p v-else-if="failure === 'unavailable'" role="alert">文件服务暂不可用。</p>
    <p v-else-if="failure === 'conflict'" role="alert">文件状态已发生变化，请刷新后重试。</p>
    <template v-if="projectionVisible && canRead">
      <form class="files-page__search" @submit.prevent="settle(() => controller.list.search({ search }))">
        <label>文件名称<input v-model="search" name="search" maxlength="100" placeholder="请输入文件名称"></label>
        <button type="submit" :disabled="blocked">搜索</button>
        <button type="button" :disabled="blocked" @click="search = ''; settle(() => controller.list.reset())">重置</button>
      </form>
      <div v-if="canWrite" class="files-page__upload" data-testid="files-upload">
        <label>选择文件<input ref="fileInput" name="file" type="file" accept=".pdf,.jpg,.jpeg,.png,.txt,application/pdf,image/jpeg,image/png,text/plain" :disabled="blocked" @change="chooseFile"></label>
        <button type="button" :disabled="blocked || !selectedUpload" @click="upload">上传</button>
        <p v-if="localError" role="alert">{{ localError }}</p>
      </div>
      <div class="files-page__actions">
        <button v-if="canDelete" type="button" data-testid="files-delete-selected" :disabled="blocked || selectedRows.length === 0" @click="remove(selectedRows)">批量删除</button>
      </div>
      <div class="files-page__table">
        <table>
          <thead><tr><th v-if="canDelete" scope="col">选择</th><th scope="col" :aria-sort="sortDirection('name')"><button type="button" :disabled="blocked" @click="sortBy('name')">文件名称 <span aria-hidden="true">{{ sortMarker('name') }}</span></button></th><th scope="col">类型</th><th scope="col" :aria-sort="sortDirection('sizeBytes')"><button type="button" :disabled="blocked" @click="sortBy('sizeBytes')">大小 <span aria-hidden="true">{{ sortMarker('sizeBytes') }}</span></button></th><th scope="col" :aria-sort="sortDirection('createdAt')"><button type="button" :disabled="blocked" @click="sortBy('createdAt')">创建时间 <span aria-hidden="true">{{ sortMarker('createdAt') }}</span></button></th><th scope="col">操作</th></tr></thead>
          <tbody><tr v-for="row in snapshot.rows" :key="row.id" :data-file-id="row.id">
            <td v-if="canDelete"><input type="checkbox" :checked="snapshot.selectedKeys.includes(row.id)" :aria-label="`选择 ${row.originalName}`" :disabled="blocked" @change="toggle(row, ($event.target as HTMLInputElement).checked)"></td>
            <td>{{ row.originalName }}</td><td>{{ row.mediaType }}</td><td>{{ row.sizeBytes }}</td><td>{{ formatDate(row.createdAt) }}</td>
            <td><button type="button" :disabled="blocked" @click="download(row)">下载</button><button v-if="canDelete" type="button" :disabled="blocked" @click="remove([row])">删除</button></td>
          </tr></tbody>
        </table>
      </div>
      <nav aria-label="分页">
        <button type="button" :disabled="blocked || snapshot.page <= 1" @click="settle(() => controller.list.setPage(snapshot.page - 1))">上一页</button>
        <span>第 {{ snapshot.page }} 页</span>
        <label>每页<select :value="snapshot.pageSize" :disabled="blocked" @change="settle(() => controller.list.setPageSize(Number(($event.target as HTMLSelectElement).value)))"><option :value="10">10</option><option :value="20">20</option><option :value="50">50</option></select></label>
        <button type="button" :disabled="blocked || snapshot.page * snapshot.pageSize >= snapshot.total" @click="settle(() => controller.list.setPage(snapshot.page + 1))">下一页</button>
      </nav>
    </template>
    <button v-else-if="failure === 'unavailable'" type="button" @click="settle(() => controller.list.refresh())">重试</button>
  </section>
</template>

<style scoped>
.files-page { display: grid; gap: 16px; color: var(--ga-text-1); }
.files-page__header { display: flex; align-items: end; justify-content: space-between; gap: 12px; }
h1, p { margin: 0; }
label { display: grid; gap: 4px; }
.files-page__table { min-width: 0; overflow-x: auto; }
td:last-child { display: flex; gap: 8px; }
@media (max-width: 720px) { .files-page__header, .files-page__search, .files-page__upload, .files-page__actions, nav { align-items: stretch; flex-direction: column; } }
</style>
