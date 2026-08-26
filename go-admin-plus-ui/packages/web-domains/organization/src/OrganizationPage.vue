<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import type { Department, DepartmentInput, Position, PositionInput } from '@go-admin/domain-organization'
import { settleOrganizationPageOperation, type OrganizationController } from './organization-controller'

const props = defineProps<{ controller: OrganizationController }>()
const emit = defineEmits<{ sessionRequired: [] }>()
const tab = ref<'departments' | 'positions'>('departments')
const revision = ref(0)
const filters = reactive({ search: '' })
const department = reactive<DepartmentInput>({ key: '', name: '', parentId: '', sortOrder: 0 })
const position = reactive<PositionInput>({ key: '', name: '', departmentId: '', enabled: true })
const editedDepartment = ref<(DepartmentInput & { id: string; protected: boolean }) | null>(null)
const editedPosition = ref<(PositionInput & { id: string; protected: boolean }) | null>(null)
const failureMessage = ref('')
const departments = computed(() => { void revision.value; return props.controller.departments() })
const tree = computed(() => { void revision.value; return props.controller.departmentTree() })
const departmentRows = computed(() => {
  void tree.value
  const byID = new Map(departments.value.map((item) => [item.id, item]))
  return departments.value.map((item) => {
    let depth = 0
    let parent = item.parentId ? byID.get(item.parentId) : undefined
    while (parent) { depth += 1; parent = parent.parentId ? byID.get(parent.parentId) : undefined }
    return { item, depth }
  })
})
const positions = computed(() => { void revision.value; return props.controller.positions.snapshot() })
const blocked = computed(() => { void revision.value; return props.controller.busy || props.controller.hasPendingRepair() })
const can = (permission: string) => { void revision.value; return props.controller.can(permission) }
const surfaceFailure = () => {
  const failure = props.controller.failure()
  failureMessage.value = failure === 'relogin' ? 'Your session must be renewed.'
    : failure === 'forbidden' ? 'You do not have permission for that action.'
      : failure === 'validation' ? 'Review the submitted values.'
        : failure === 'not-found' ? 'The organization record no longer exists.'
          : failure === 'conflict' ? 'The record is protected or still referenced.'
            : failure === 'unavailable' ? 'The organization service is unavailable.' : ''
  if (failure === 'relogin') emit('sessionRequired')
}
const run = (operation: () => Promise<unknown>) => settleOrganizationPageOperation(operation, () => { surfaceFailure(); revision.value += 1 })
const editDepartment = (item: Department) => { editedDepartment.value = { id: item.id, key: item.key, name: item.name, parentId: item.parentId ?? '', sortOrder: item.sortOrder, protected: item.protected } }
const editPosition = (item: Position) => { editedPosition.value = { id: item.id, key: item.key, name: item.name, departmentId: item.departmentId, enabled: item.enabled, protected: item.protected } }
const search = () => run(() => props.controller.positions.search({ ...filters }))
const reset = () => run(async () => { filters.search = ''; await props.controller.positions.reset() })
const submitDepartment = () => run(() => props.controller.createDepartment({ ...department }))
const submitPosition = () => run(() => props.controller.createPosition({ ...position }))
const saveDepartment = () => editedDepartment.value && run(() => props.controller.updateDepartment(editedDepartment.value!.id, {
  key: editedDepartment.value!.key,
  name: editedDepartment.value!.name,
  parentId: editedDepartment.value!.parentId,
  sortOrder: editedDepartment.value!.sortOrder,
}))
const savePosition = () => editedPosition.value && run(() => props.controller.updatePosition(editedPosition.value!.id, {
  key: editedPosition.value!.key,
  name: editedPosition.value!.name,
  departmentId: editedPosition.value!.departmentId,
  enabled: editedPosition.value!.enabled,
}))
onMounted(() => run(async () => {
  if (can('organization.departments.read')) await props.controller.refreshDepartments()
  if (can('organization.positions.read')) await props.controller.positions.refresh()
  if (!can('organization.departments.read')) tab.value = 'positions'
}))
</script>

<template>
  <main class="organization-page">
    <header><h1>Organization</h1></header>
    <p v-if="failureMessage" role="alert">{{ failureMessage }}</p>
    <button v-if="controller.hasPendingRepair()" type="button" data-testid="repair-organization" :disabled="controller.busy" @click="run(() => controller.repairProjection())">Refresh saved changes</button>
    <nav class="tabs" aria-label="Organization views">
      <button v-if="can('organization.departments.read')" type="button" :aria-pressed="tab === 'departments'" @click="tab = 'departments'">Departments</button>
      <button v-if="can('organization.positions.read')" type="button" :aria-pressed="tab === 'positions'" @click="tab = 'positions'">Positions</button>
    </nav>

    <section v-if="tab === 'departments' && can('organization.departments.read')" aria-labelledby="departments-heading">
      <h2 id="departments-heading">Department tree</h2>
      <ul class="tree">
        <li v-for="row in departmentRows" :key="row.item.id" :data-row-key="row.item.key" :style="{ paddingInlineStart: `${row.depth * 24}px` }"><strong>{{ row.item.name }}</strong><span>{{ row.item.key }}</span><button v-if="can('organization.departments.write')" type="button" data-action="edit" @click="editDepartment(row.item)">Edit</button><button v-if="can('organization.departments.delete')" type="button" data-action="delete" :disabled="row.item.protected || blocked" @click="run(() => controller.deleteDepartment(row.item.id))">Delete</button></li>
      </ul>
      <form v-if="can('organization.departments.write')" class="editor" data-testid="create-department" @submit.prevent="submitDepartment">
        <h3>Create department</h3><label>Key<input name="key" v-model.trim="department.key" required minlength="3" maxlength="64" pattern="[a-z0-9][a-z0-9_-]*"></label><label>Name<input name="name" v-model.trim="department.name" required maxlength="100"></label><label>Parent<select name="parentId" v-model="department.parentId" required><option disabled value="">Select</option><option v-for="item in departments" :key="item.id" :value="item.id">{{ item.name }}</option></select></label><label>Order<input name="sortOrder" v-model.number="department.sortOrder" type="number" min="-1000000" max="1000000"></label><button type="submit" :disabled="blocked">Create</button>
      </form>
      <form v-if="editedDepartment && can('organization.departments.write')" class="editor" data-testid="edit-department" @submit.prevent="saveDepartment">
        <h3>Edit {{ editedDepartment.key }}</h3><label>Name<input v-model.trim="editedDepartment.name" required maxlength="100"></label><label>Parent<select v-model="editedDepartment.parentId" required><option v-for="item in departments.filter((candidate) => candidate.id !== editedDepartment?.id)" :key="item.id" :value="item.id">{{ item.name }}</option></select></label><label>Order<input v-model.number="editedDepartment.sortOrder" type="number" min="-1000000" max="1000000"></label><button type="submit" :disabled="editedDepartment.protected || blocked">Save</button>
      </form>
    </section>

    <section v-else-if="tab === 'positions' && can('organization.positions.read')" aria-labelledby="positions-heading">
      <h2 id="positions-heading">Positions</h2>
      <form class="toolbar" data-testid="position-search" @submit.prevent="search"><label>Search<input name="search" v-model.trim="filters.search" maxlength="100"></label><button type="submit">Search</button><button type="button" @click="reset">Reset</button></form>
      <table><thead><tr><th>Key</th><th>Name</th><th>Department</th><th>Status</th><th>Actions</th></tr></thead><tbody><tr v-for="item in positions.rows" :key="item.id" :data-row-key="item.key"><td>{{ item.key }}</td><td>{{ item.name }}</td><td>{{ departments.find((candidate) => candidate.id === item.departmentId)?.name ?? item.departmentId }}</td><td>{{ item.enabled ? 'Enabled' : 'Disabled' }}</td><td><button v-if="can('organization.positions.write')" type="button" data-action="edit" @click="editPosition(item)">Edit</button><button v-if="can('organization.positions.delete')" type="button" data-action="delete" :disabled="item.protected || blocked" @click="run(() => controller.deletePosition(item.id))">Delete</button></td></tr></tbody></table>
      <div class="pagination"><button type="button" :disabled="positions.page <= 1" @click="run(() => controller.positions.setPage(positions.page - 1))">Previous</button><span>Page {{ positions.page }}</span><button type="button" :disabled="positions.page * positions.pageSize >= positions.total" @click="run(() => controller.positions.setPage(positions.page + 1))">Next</button></div>
      <form v-if="can('organization.positions.write')" class="editor" data-testid="create-position" @submit.prevent="submitPosition"><h3>Create position</h3><label>Key<input name="key" v-model.trim="position.key" required minlength="3" maxlength="64" pattern="[a-z0-9][a-z0-9_-]*"></label><label>Name<input name="name" v-model.trim="position.name" required maxlength="100"></label><label>Department<select name="departmentId" v-model="position.departmentId" required><option disabled value="">Select</option><option v-for="item in departments" :key="item.id" :value="item.id">{{ item.name }}</option></select></label><label class="choice"><input name="enabled" v-model="position.enabled" type="checkbox">Enabled</label><button type="submit" :disabled="blocked">Create</button></form>
      <form v-if="editedPosition && can('organization.positions.write')" class="editor" data-testid="edit-position" @submit.prevent="savePosition"><h3>Edit {{ editedPosition.key }}</h3><label>Name<input v-model.trim="editedPosition.name" required maxlength="100"></label><label>Department<select v-model="editedPosition.departmentId" required><option v-for="item in departments" :key="item.id" :value="item.id">{{ item.name }}</option></select></label><label class="choice"><input v-model="editedPosition.enabled" type="checkbox">Enabled</label><button type="submit" :disabled="editedPosition.protected || blocked">Save</button></form>
    </section>
    <section v-else><p>No organization view is available.</p></section>
  </main>
</template>

<style scoped>
.organization-page { display: grid; gap: 20px; max-width: 1120px; padding: 24px; }
h1, h2, h3 { margin: 0; letter-spacing: 0; }
.tabs, .toolbar, .pagination { display: flex; gap: 8px; align-items: end; flex-wrap: wrap; }
.tabs button[aria-pressed="true"] { border-bottom-color: #176b54; color: #176b54; }
section, .editor { display: grid; gap: 16px; }
.tree { display: grid; gap: 6px; margin: 0; padding: 0; list-style: none; }
.tree li { display: grid; grid-template-columns: minmax(160px, 1fr) minmax(100px, 1fr) auto auto; align-items: center; gap: 8px; min-height: 42px; border-bottom: 1px solid #d7dce2; }
.editor { grid-template-columns: repeat(auto-fit, minmax(170px, 1fr)); align-items: end; border-top: 1px solid #d7dce2; padding-top: 16px; }
.editor h3 { grid-column: 1 / -1; }
label { display: grid; gap: 6px; }
.choice { grid-template-columns: auto 1fr; align-items: center; }
input, select, button { min-height: 40px; font: inherit; }
button { border: 1px solid #aab2bc; background: #fff; padding: 6px 12px; }
button:disabled { opacity: .5; }
table { width: 100%; border-collapse: collapse; }
th, td { border-bottom: 1px solid #d7dce2; padding: 10px; text-align: left; }
th { color: #4c5968; font-size: 14px; }
@media (max-width: 640px) { .organization-page { padding: 16px; overflow-x: auto; } table { min-width: 680px; } .tree li { grid-template-columns: minmax(120px, 1fr) auto; } }
</style>
