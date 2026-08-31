<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import type { Department, DepartmentInput, Position, PositionInput } from '@go-admin-plus/domain-organization'
import { settleOrganizationPageOperation, type OrganizationController } from './organization-controller'
import { organizationViewForPath } from './organization-view'

const props = defineProps<{ controller: OrganizationController }>()
const emit = defineEmits<{ sessionRequired: [] }>()
const route = useRoute()
const view = computed(() => organizationViewForPath(route.path))
const revision = ref(0)
const filters = reactive({ search: '' })
const department = reactive<DepartmentInput>({ key: '', name: '', parentId: '', sortOrder: 0 })
const position = reactive<PositionInput>({ key: '', name: '', departmentId: '', enabled: true })
const createDepartmentOpen = ref(false)
const createPositionOpen = ref(false)
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
  failureMessage.value = failure === 'relogin' ? '会话已失效，请重新登录。'
    : failure === 'forbidden' ? '没有执行该操作的权限。'
      : failure === 'validation' ? '请检查提交内容。'
        : failure === 'not-found' ? '组织记录已不存在。'
          : failure === 'conflict' ? '该记录受保护或仍被引用。'
            : failure === 'unavailable' ? '组织服务暂不可用。' : ''
  if (failure === 'relogin') emit('sessionRequired')
}
const run = <T>(operation: () => Promise<T>) => settleOrganizationPageOperation(operation, () => { surfaceFailure(); revision.value += 1 })
const resetDepartmentForm = () => { Object.assign(department, { key: '', name: '', parentId: '', sortOrder: 0 }) }
const openCreateDepartment = () => { resetDepartmentForm(); createDepartmentOpen.value = true }
const closeCreateDepartment = () => { createDepartmentOpen.value = false; resetDepartmentForm() }
const resetPositionForm = () => { Object.assign(position, { key: '', name: '', departmentId: '', enabled: true }) }
const openCreatePosition = () => { resetPositionForm(); createPositionOpen.value = true }
const closeCreatePosition = () => { createPositionOpen.value = false; resetPositionForm() }
const closeDialogs = () => {
  closeCreateDepartment(); closeCreatePosition()
  editedDepartment.value = null; editedPosition.value = null
}
const editDepartment = (item: Department) => { editedDepartment.value = { id: item.id, key: item.key, name: item.name, parentId: item.parentId ?? '', sortOrder: item.sortOrder, protected: item.protected } }
const editPosition = (item: Position) => { editedPosition.value = { id: item.id, key: item.key, name: item.name, departmentId: item.departmentId, enabled: item.enabled, protected: item.protected } }
const search = () => run(() => props.controller.positions.search({ ...filters }))
const reset = () => run(async () => { filters.search = ''; await props.controller.positions.reset() })
const submitDepartment = async () => {
  if (await run(() => props.controller.createDepartment({ ...department })) === 'completed') {
    closeCreateDepartment()
  }
}
const submitPosition = async () => {
  if (await run(() => props.controller.createPosition({ ...position })) === 'completed') {
    closeCreatePosition()
  }
}
const saveDepartment = async () => {
  if (!editedDepartment.value) return
  const result = await run(() => props.controller.updateDepartment(editedDepartment.value!.id, {
    key: editedDepartment.value!.key,
    name: editedDepartment.value!.name,
    parentId: editedDepartment.value!.parentId,
    sortOrder: editedDepartment.value!.sortOrder,
  }))
  if (result === 'completed') editedDepartment.value = null
}
const savePosition = async () => {
  if (!editedPosition.value) return
  const result = await run(() => props.controller.updatePosition(editedPosition.value!.id, {
    key: editedPosition.value!.key,
    name: editedPosition.value!.name,
    departmentId: editedPosition.value!.departmentId,
    enabled: editedPosition.value!.enabled,
  }))
  if (result === 'completed') editedPosition.value = null
}
const loadView = () => run(async () => {
  if (can('organization.departments.read')) await props.controller.refreshDepartments()
  if (view.value === 'positions' && can('organization.positions.read')) await props.controller.positions.refresh()
})
onMounted(loadView)
watch(view, () => { closeDialogs(); void loadView() })
</script>

<template>
  <main class="organization-page">
    <header><h1>组织管理</h1></header>
    <p v-if="failureMessage" role="alert">{{ failureMessage }}</p>
    <button v-if="controller.hasPendingRepair()" type="button" data-testid="repair-organization" :disabled="controller.busy" @click="run(() => controller.repairProjection())">刷新已保存的变更</button>
    <section v-if="view === 'departments' && can('organization.departments.read')" aria-labelledby="departments-heading">
      <h2 id="departments-heading">部门列表</h2>
      <div class="toolbar management-toolbar"><button v-if="can('organization.departments.write')" type="button" data-testid="open-create-department" @click="openCreateDepartment">新增</button></div>
      <ul class="tree">
        <li v-for="row in departmentRows" :key="row.item.id" :data-row-key="row.item.key" :style="{ paddingInlineStart: `${row.depth * 24}px` }"><strong>{{ row.item.name }}</strong><span>{{ row.item.key }}</span><button v-if="can('organization.departments.write')" type="button" data-action="edit" @click="editDepartment(row.item)">修改</button><button v-if="can('organization.departments.delete')" type="button" data-action="delete" :disabled="row.item.protected || blocked" @click="run(() => controller.deleteDepartment(row.item.id))">删除</button></li>
      </ul>
      <div v-if="createDepartmentOpen && can('organization.departments.write')" class="management-dialog-backdrop" @click.self="closeCreateDepartment" @keydown.esc="closeCreateDepartment"><form class="management-dialog" data-testid="create-department" role="dialog" aria-modal="true" aria-labelledby="create-department-title" @submit.prevent="submitDepartment"><header class="management-dialog__header"><h3 id="create-department-title">新增部门</h3><button type="button" aria-label="关闭" @click="closeCreateDepartment">×</button></header><div class="management-dialog__body"><label>部门编码<input name="key" v-model.trim="department.key" autofocus required minlength="3" maxlength="64" pattern="[a-z0-9][a-z0-9_-]*"></label><label>部门名称<input name="name" v-model.trim="department.name" required maxlength="100"></label><label>上级部门<select name="parentId" v-model="department.parentId" required><option disabled value="">请选择</option><option v-for="item in departments" :key="item.id" :value="item.id">{{ item.name }}</option></select></label><label>显示排序<input name="sortOrder" v-model.number="department.sortOrder" type="number" min="-1000000" max="1000000"></label></div><footer class="management-dialog__footer"><button type="button" @click="closeCreateDepartment">取消</button><button type="submit" :disabled="blocked">确定</button></footer></form></div>
      <div v-if="editedDepartment && can('organization.departments.write')" class="management-dialog-backdrop" @click.self="editedDepartment = null" @keydown.esc="editedDepartment = null"><form class="management-dialog" data-testid="edit-department" role="dialog" aria-modal="true" aria-labelledby="edit-department-title" @submit.prevent="saveDepartment"><header class="management-dialog__header"><h3 id="edit-department-title">修改部门 {{ editedDepartment.key }}</h3><button type="button" aria-label="关闭" @click="editedDepartment = null">×</button></header><div class="management-dialog__body"><label>部门名称<input v-model.trim="editedDepartment.name" required maxlength="100"></label><label>上级部门<select v-model="editedDepartment.parentId" required><option v-for="item in departments.filter((candidate) => candidate.id !== editedDepartment?.id)" :key="item.id" :value="item.id">{{ item.name }}</option></select></label><label>显示排序<input v-model.number="editedDepartment.sortOrder" type="number" min="-1000000" max="1000000"></label></div><footer class="management-dialog__footer"><button type="button" @click="editedDepartment = null">取消</button><button type="submit" :disabled="editedDepartment.protected || blocked">保存</button></footer></form></div>
    </section>

    <section v-else-if="view === 'positions' && can('organization.positions.read')" aria-labelledby="positions-heading">
      <h2 id="positions-heading">岗位列表</h2>
      <form class="toolbar" data-testid="position-search" @submit.prevent="search"><label>岗位搜索<input name="search" v-model.trim="filters.search" maxlength="100" placeholder="请输入岗位编码或名称"></label><button type="submit">搜索</button><button type="button" @click="reset">重置</button><button v-if="can('organization.positions.write')" type="button" data-testid="open-create-position" @click="openCreatePosition">新增</button></form>
      <table><thead><tr><th>岗位编码</th><th>岗位名称</th><th>所属部门</th><th>状态</th><th>操作</th></tr></thead><tbody><tr v-for="item in positions.rows" :key="item.id" :data-row-key="item.key"><td>{{ item.key }}</td><td>{{ item.name }}</td><td>{{ departments.find((candidate) => candidate.id === item.departmentId)?.name ?? item.departmentId }}</td><td>{{ item.enabled ? '启用' : '停用' }}</td><td><button v-if="can('organization.positions.write')" type="button" data-action="edit" @click="editPosition(item)">修改</button><button v-if="can('organization.positions.delete')" type="button" data-action="delete" :disabled="item.protected || blocked" @click="run(() => controller.deletePosition(item.id))">删除</button></td></tr></tbody></table>
      <div class="pagination" data-testid="organization-positions-pagination"><span>共 {{ positions.total }} 条</span><button type="button" :disabled="blocked || positions.page <= 1" @click="run(() => controller.positions.setPage(positions.page - 1))">上一页</button><span>第 {{ positions.page }} 页</span><label>每页<select :value="positions.pageSize" :disabled="blocked" @change="run(() => controller.positions.setPageSize(Number(($event.target as HTMLSelectElement).value)))"><option :value="10">10</option><option :value="20">20</option><option :value="50">50</option></select></label><button type="button" :disabled="blocked || positions.page * positions.pageSize >= positions.total" @click="run(() => controller.positions.setPage(positions.page + 1))">下一页</button></div>
      <div v-if="createPositionOpen && can('organization.positions.write')" class="management-dialog-backdrop" @click.self="closeCreatePosition" @keydown.esc="closeCreatePosition"><form class="management-dialog" data-testid="create-position" role="dialog" aria-modal="true" aria-labelledby="create-position-title" @submit.prevent="submitPosition"><header class="management-dialog__header"><h3 id="create-position-title">新增岗位</h3><button type="button" aria-label="关闭" @click="closeCreatePosition">×</button></header><div class="management-dialog__body"><label>岗位编码<input name="key" v-model.trim="position.key" autofocus required minlength="3" maxlength="64" pattern="[a-z0-9][a-z0-9_-]*"></label><label>岗位名称<input name="name" v-model.trim="position.name" required maxlength="100"></label><label>所属部门<select name="departmentId" v-model="position.departmentId" required><option disabled value="">请选择</option><option v-for="item in departments" :key="item.id" :value="item.id">{{ item.name }}</option></select></label><label class="choice"><input name="enabled" v-model="position.enabled" type="checkbox">启用</label></div><footer class="management-dialog__footer"><button type="button" @click="closeCreatePosition">取消</button><button type="submit" :disabled="blocked">确定</button></footer></form></div>
      <div v-if="editedPosition && can('organization.positions.write')" class="management-dialog-backdrop" @click.self="editedPosition = null" @keydown.esc="editedPosition = null"><form class="management-dialog" data-testid="edit-position" role="dialog" aria-modal="true" aria-labelledby="edit-position-title" @submit.prevent="savePosition"><header class="management-dialog__header"><h3 id="edit-position-title">修改岗位 {{ editedPosition.key }}</h3><button type="button" aria-label="关闭" @click="editedPosition = null">×</button></header><div class="management-dialog__body"><label>岗位名称<input v-model.trim="editedPosition.name" required maxlength="100"></label><label>所属部门<select v-model="editedPosition.departmentId" required><option v-for="item in departments" :key="item.id" :value="item.id">{{ item.name }}</option></select></label><label class="choice"><input v-model="editedPosition.enabled" type="checkbox">启用</label></div><footer class="management-dialog__footer"><button type="button" @click="editedPosition = null">取消</button><button type="submit" :disabled="editedPosition.protected || blocked">保存</button></footer></form></div>
    </section>
    <section v-else><p>当前没有可访问的组织管理视图。</p></section>
  </main>
</template>

<style scoped>
.organization-page { display: grid; gap: 20px; }
h1, h2, h3 { margin: 0; letter-spacing: 0; }
.toolbar, .pagination { display: flex; gap: 8px; align-items: end; flex-wrap: wrap; }
section { display: grid; gap: 16px; }
.tree { display: grid; gap: 6px; margin: 0; padding: 0; list-style: none; }
.tree li { display: grid; grid-template-columns: minmax(160px, 1fr) minmax(100px, 1fr) auto auto; align-items: center; gap: 8px; min-height: 42px; border-bottom: 1px solid var(--ga-border-light); }
label { display: grid; gap: 6px; }
.choice { grid-template-columns: auto 1fr; align-items: center; }
@media (max-width: 640px) { .organization-page { overflow-x: auto; } table { min-width: 680px; } .tree li { grid-template-columns: minmax(120px, 1fr) auto; } }
</style>
