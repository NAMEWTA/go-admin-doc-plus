<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { filesPermissions, type FileMetadata, type UploadCandidate } from '@go-admin/domain-files'
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

const chooseFile = (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  selectedUpload.value = file ? { name: file.name, type: file.type, size: file.size, body: file } : null
  localError.value = null
}
const upload = async () => {
  if (!selectedUpload.value) { localError.value = 'Select a file'; return }
  const result = await props.controller.upload(selectedUpload.value)
  if (result === 'invalid') localError.value = 'File is not accepted'
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
      <div><h1 id="files-title">Files</h1><p>{{ snapshot.total }} objects</p></div>
      <button v-if="controller.pendingRepair" type="button" data-testid="files-repair" :disabled="controller.busy" @click="repair">Retry refresh</button>
    </header>
    <p v-if="failure === 'relogin'" role="alert">Session required</p>
    <p v-else-if="failure === 'forbidden'" role="alert">Permission denied</p>
    <p v-else-if="failure === 'unavailable'" role="alert">Files are unavailable</p>
    <p v-else-if="failure === 'conflict'" role="alert">File state changed</p>
    <template v-if="projectionVisible && canRead">
      <form class="files-page__search" @submit.prevent="settle(() => controller.list.search({ search }))">
        <label>Search <input v-model="search" name="search" maxlength="100"></label>
        <button type="submit" :disabled="blocked">Search</button>
        <button type="button" :disabled="blocked" @click="search = ''; settle(() => controller.list.reset())">Reset</button>
      </form>
      <div v-if="canWrite" class="files-page__upload" data-testid="files-upload">
        <label>File <input ref="fileInput" name="file" type="file" accept=".pdf,.jpg,.jpeg,.png,.txt,application/pdf,image/jpeg,image/png,text/plain" :disabled="blocked" @change="chooseFile"></label>
        <button type="button" :disabled="blocked || !selectedUpload" @click="upload">Upload</button>
        <p v-if="localError" role="alert">{{ localError }}</p>
      </div>
      <div class="files-page__actions">
        <button v-if="canDelete" type="button" data-testid="files-delete-selected" :disabled="blocked || selectedRows.length === 0" @click="remove(selectedRows)">Delete selected</button>
      </div>
      <div class="files-page__table">
        <table>
          <thead><tr><th v-if="canDelete" scope="col">Select</th><th scope="col">Name</th><th scope="col">Type</th><th scope="col">Size</th><th scope="col">Updated</th><th scope="col">Actions</th></tr></thead>
          <tbody><tr v-for="row in snapshot.rows" :key="row.id" :data-file-id="row.id">
            <td v-if="canDelete"><input type="checkbox" :checked="snapshot.selectedKeys.includes(row.id)" :aria-label="`Select ${row.originalName}`" :disabled="blocked" @change="toggle(row, ($event.target as HTMLInputElement).checked)"></td>
            <td>{{ row.originalName }}</td><td>{{ row.mediaType }}</td><td>{{ row.sizeBytes }}</td><td>{{ row.updatedAt }}</td>
            <td><button type="button" :disabled="blocked" @click="download(row)">Download</button><button v-if="canDelete" type="button" :disabled="blocked" @click="remove([row])">Delete</button></td>
          </tr></tbody>
        </table>
      </div>
      <nav aria-label="Pagination">
        <button type="button" :disabled="blocked || snapshot.page <= 1" @click="settle(() => controller.list.setPage(snapshot.page - 1))">Previous</button>
        <span>Page {{ snapshot.page }}</span>
        <label>Rows <select :value="snapshot.pageSize" :disabled="blocked" @change="settle(() => controller.list.setPageSize(Number(($event.target as HTMLSelectElement).value)))"><option :value="10">10</option><option :value="20">20</option><option :value="50">50</option></select></label>
        <button type="button" :disabled="blocked || snapshot.page * snapshot.pageSize >= snapshot.total" @click="settle(() => controller.list.setPage(snapshot.page + 1))">Next</button>
      </nav>
    </template>
    <button v-else-if="failure === 'unavailable'" type="button" @click="settle(() => controller.list.refresh())">Retry</button>
  </section>
</template>

<style scoped>
.files-page { display: grid; gap: 16px; color: #17202a; }
.files-page__header, .files-page__search, .files-page__upload, .files-page__actions, nav { display: flex; align-items: end; justify-content: space-between; gap: 12px; }
h1, p { margin: 0; }
label { display: grid; gap: 4px; }
.files-page__table { min-width: 0; overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 8px; border-bottom: 1px solid #dfe6e9; text-align: left; }
td:last-child { display: flex; gap: 8px; }
input, select, button { font: inherit; }
button { min-height: 34px; }
[role="alert"] { padding: 8px; border-left: 3px solid #b42318; background: #fff1f0; }
@media (max-width: 720px) { .files-page__header, .files-page__search, .files-page__upload, .files-page__actions, nav { align-items: stretch; flex-direction: column; } }
</style>
