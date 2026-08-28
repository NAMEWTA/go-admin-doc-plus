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
  failureMessage.value = failure === 'relogin' ? '会话已失效，请重新登录。'
    : failure === 'forbidden' ? '没有执行该操作的权限。'
      : failure === 'validation' ? '请检查提交内容。'
        : failure === 'not-found' ? '组织记录已不存在。'
          : failure === 'conflict' ? '该记录受保护或仍被引用。'
            : failure === 'unavailable' ? '组织服务暂不可用。' : ''
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
    <header><h1>组织管理</h1></header>
    <p v-if="failureMessage" role="alert">{{ failureMessage }}</p>
    <button v-if="controller.hasPendingRepair()" type="button" data-testid="repair-organization" :disabled="controller.busy" @click="run(() => controller.repairProjection())">刷新已保存的变更</button>
    <nav class="tabs" aria-label="Organization views">
      <button v-if="can('organization.departments.read')" type="button" :aria-pressed="tab === 'departments'" @click="tab = 'departments'">部门管理</button>
      <button v-if="can('organization.positions.read')" type="button" :aria-pressed="tab === 'positions'" @click="tab = 'positions'">岗位管理</button>
    </nav>

    <section v-if="tab === 'departments' && can('organization.departments.read')" aria-labelledby="departments-heading">
      <h2 id="departments-heading">部门列表</h2>
      <ul class="tree">
        <li v-for="row in departmentRows" :key="row.item.id" :data-row-key="row.item.key" :style="{ paddingInlineStart: `${row.depth * 24}px` }"><strong>{{ row.item.name }}</strong><span>{{ row.item.key }}</span><button v-if="can('organization.departments.write')" type="button" data-action="edit" @click="editDepartment(row.item)">修改</button><button v-if="can('organization.departments.delete')" type="button" data-action="delete" :disabled="row.item.protected || blocked" @click="run(() => controller.deleteDepartment(row.item.id))">删除</button></li>
      </ul>
      <form v-if="can('organization.departments.write')" class="editor" data-testid="create-department" @submit.prevent="submitDepartment">
        <h3>新增部门</h3><label>部门编码<input name="key" v-model.trim="department.key" required minlength="3" maxlength="64" pattern="[a-z0-9][a-z0-9_-]*"></label><label>部门名称<input name="name" v-model.trim="department.name" required maxlength="100"></label><label>上级部门<select name="parentId" v-model="department.parentId" required><option disabled value="">请选择</option><option v-for="item in departments" :key="item.id" :value="item.id">{{ item.name }}</option></select></label><label>显示排序<input name="sortOrder" v-model.number="department.sortOrder" type="number" min="-1000000" max="1000000"></label><button type="submit" :disabled="blocked">新增</button>
      </form>
      <form v-if="editedDepartment && can('organization.departments.write')" class="editor" data-testid="edit-department" @submit.prevent="saveDepartment">
        <h3>修改部门 {{ editedDepartment.key }}</h3><label>部门名称<input v-model.trim="editedDepartment.name" required maxlength="100"></label><label>上级部门<select v-model="editedDepartment.parentId" required><option v-for="item in departments.filter((candidate) => candidate.id !== editedDepartment?.id)" :key="item.id" :value="item.id">{{ item.name }}</option></select></label><label>显示排序<input v-model.number="editedDepartment.sortOrder" type="number" min="-1000000" max="1000000"></label><button type="submit" :disabled="editedDepartment.protected || blocked">保存</button>
      </form>
    </section>

    <section v-else-if="tab === 'positions' && can('organization.positions.read')" aria-labelledby="positions-heading">
      <h2 id="positions-heading">岗位列表</h2>
      <form class="toolbar" data-testid="position-search" @submit.prevent="search"><label>岗位搜索<input name="search" v-model.trim="filters.search" maxlength="100" placeholder="请输入岗位编码或名称"></label><button type="submit">搜索</button><button type="button" @click="reset">重置</button></form>
      <table><thead><tr><th>岗位编码</th><th>岗位名称</th><th>所属部门</th><th>状态</th><th>操作</th></tr></thead><tbody><tr v-for="item in positions.rows" :key="item.id" :data-row-key="item.key"><td>{{ item.key }}</td><td>{{ item.name }}</td><td>{{ departments.find((candidate) => candidate.id === item.departmentId)?.name ?? item.departmentId }}</td><td>{{ item.enabled ? '启用' : '停用' }}</td><td><button v-if="can('organization.positions.write')" type="button" data-action="edit" @click="editPosition(item)">修改</button><button v-if="can('organization.positions.delete')" type="button" data-action="delete" :disabled="item.protected || blocked" @click="run(() => controller.deletePosition(item.id))">删除</button></td></tr></tbody></table>
      <div class="pagination"><button type="button" :disabled="positions.page <= 1" @click="run(() => controller.positions.setPage(positions.page - 1))">上一页</button><span>第 {{ positions.page }} 页</span><button type="button" :disabled="positions.page * positions.pageSize >= positions.total" @click="run(() => controller.positions.setPage(positions.page + 1))">下一页</button></div>
      <form v-if="can('organization.positions.write')" class="editor" data-testid="create-position" @submit.prevent="submitPosition"><h3>新增岗位</h3><label>岗位编码<input name="key" v-model.trim="position.key" required minlength="3" maxlength="64" pattern="[a-z0-9][a-z0-9_-]*"></label><label>岗位名称<input name="name" v-model.trim="position.name" required maxlength="100"></label><label>所属部门<select name="departmentId" v-model="position.departmentId" required><option disabled value="">请选择</option><option v-for="item in departments" :key="item.id" :value="item.id">{{ item.name }}</option></select></label><label class="choice"><input name="enabled" v-model="position.enabled" type="checkbox">启用</label><button type="submit" :disabled="blocked">新增</button></form>
      <form v-if="editedPosition && can('organization.positions.write')" class="editor" data-testid="edit-position" @submit.prevent="savePosition"><h3>修改岗位 {{ editedPosition.key }}</h3><label>岗位名称<input v-model.trim="editedPosition.name" required maxlength="100"></label><label>所属部门<select v-model="editedPosition.departmentId" required><option v-for="item in departments" :key="item.id" :value="item.id">{{ item.name }}</option></select></label><label class="choice"><input v-model="editedPosition.enabled" type="checkbox">启用</label><button type="submit" :disabled="editedPosition.protected || blocked">保存</button></form>
    </section>
    <section v-else><p>当前没有可访问的组织管理视图。</p></section>
  </main>
</template>

<style scoped>
.organization-page { display: grid; gap: 20px; max-width: 1120px; padding: 24px; }
h1, h2, h3 { margin: 0; letter-spacing: 0; }
.tabs, .toolbar, .pagination { display: flex; gap: 8px; align-items: end; flex-wrap: wrap; }
.tabs button[aria-pressed="true"] { border-bottom-color: var(--ga-brand); color: var(--ga-brand); }
section, .editor { display: grid; gap: 16px; }
.tree { display: grid; gap: 6px; margin: 0; padding: 0; list-style: none; }
.tree li { display: grid; grid-template-columns: minmax(160px, 1fr) minmax(100px, 1fr) auto auto; align-items: center; gap: 8px; min-height: 42px; border-bottom: 1px solid var(--ga-border-light); }
.editor { grid-template-columns: repeat(auto-fit, minmax(170px, 1fr)); align-items: end; border-top: 1px solid var(--ga-border-light); padding-top: 16px; }
.editor h3 { grid-column: 1 / -1; }
label { display: grid; gap: 6px; }
.choice { grid-template-columns: auto 1fr; align-items: center; }
input, select, button { min-height: 40px; font: inherit; }
button { border: 1px solid var(--ga-border); background: var(--ga-bg-container); padding: 6px 12px; }
button:disabled { opacity: .5; }
table { width: 100%; border-collapse: collapse; }
th, td { border-bottom: 1px solid var(--ga-border-light); padding: 10px; text-align: left; }
th { color: var(--ga-text-2); font-size: 14px; }
@media (max-width: 640px) { .organization-page { padding: 16px; overflow-x: auto; } table { min-width: 680px; } .tree li { grid-template-columns: minmax(120px, 1fr) auto; } }
</style>
