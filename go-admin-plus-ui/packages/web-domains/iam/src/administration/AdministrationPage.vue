<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import type { Menu, Role, User } from '@go-admin/domain-iam/administration'
import { createUserAndClearPassword, resetPasswordAndClear, settleAdministrationPageOperation, type AdministrationController, type CreateRoleModel, type CreateUserModel } from './administration-controller'

const props = defineProps<{ controller: AdministrationController }>()
const emit = defineEmits<{ sessionRequired: [] }>()
const tab = ref<'users' | 'roles' | 'menus'>('users')
const revision = ref(0)
const filters = reactive({ search: '' })
const user = reactive<CreateUserModel>({ username: '', displayName: '', email: '', password: '' })
const role = reactive<CreateRoleModel>({ key: '', name: '', dataScope: 'self' })
const menu = reactive({ key: '', label: '', path: '', permissionCode: '', sortOrder: 0 })
const editedUser = ref<User | null>(null)
const editedRole = ref<Role | null>(null)
const editedMenu = ref<Menu | null>(null)
const assignedRoles = ref<string[]>([])
const grantedPermissions = ref<string[]>([])
const grantedMenus = ref<string[]>([])
const replacementPassword = ref('')
const failureMessage = ref('')
const mutationBlocked = computed(() => { void revision.value; return props.controller.busy || props.controller.hasPendingRepair() })
const users = computed(() => { void revision.value; return props.controller.users.snapshot() })
const roles = computed(() => { void revision.value; return props.controller.roles() })
const menus = computed(() => { void revision.value; return props.controller.menus() })
const can = (permissionCode: string) => { void revision.value; return props.controller.can(permissionCode) }
const surfaceFailure = () => {
  const failure = props.controller.failure()
  failureMessage.value = failure === 'relogin' ? '登录状态已失效，请重新登录。'
    : failure === 'forbidden' ? '当前账号没有执行该操作的权限。'
      : failure === 'validation' ? '请检查提交内容。'
        : failure === 'conflict' ? '数据已发生变化或受系统保护。'
          : failure === 'unavailable' ? '管理服务暂时不可用。' : ''
  if (failure === 'relogin') emit('sessionRequired')
}
const run = (operation: () => Promise<unknown>) => settleAdministrationPageOperation(operation, () => {
  surfaceFailure(); revision.value += 1
})
const search = () => run(() => props.controller.users.search({ ...filters }))
const reset = () => run(async () => { filters.search = ''; await props.controller.users.reset() })
const submitUser = () => run(() => createUserAndClearPassword(props.controller, { ...user }, () => { user.password = '' }))
const submitRole = () => run(() => props.controller.createRole.run({ ...role }))
const submitMenu = () => run(() => props.controller.createMenu.run({ ...menu }))
const editUser = (value: User) => { editedUser.value = { ...value, roleIds: [...value.roleIds] }; assignedRoles.value = [...value.roleIds]; replacementPassword.value = '' }
const saveUser = () => editedUser.value && run(() => props.controller.updateUser(editedUser.value!, !editedUser.value!.disabled))
const toggleUser = (value: User) => run(() => props.controller.updateUser(value, value.disabled))
const deleteUser = (value: User) => run(() => props.controller.deleteUsers.run([value.id]))
const saveUserRoles = () => editedUser.value && run(() => props.controller.setUserRoles(editedUser.value!.id, assignedRoles.value))
const resetPassword = async () => {
  if (!editedUser.value) return
  await run(() => resetPasswordAndClear(props.controller, editedUser.value!.id, replacementPassword.value, () => { replacementPassword.value = '' }))
}
const editRole = (value: Role) => { editedRole.value = { ...value, permissionCodes: [...value.permissionCodes], menuIds: [...value.menuIds] }; grantedPermissions.value = [...value.permissionCodes]; grantedMenus.value = [...value.menuIds] }
const saveRole = () => editedRole.value && run(() => props.controller.updateRole(editedRole.value!))
const saveRoleGrants = () => editedRole.value && run(() => props.controller.setRoleGrants(editedRole.value!.id, grantedPermissions.value, grantedMenus.value))
const editMenu = (value: Menu) => { editedMenu.value = { ...value } }
const saveMenu = () => editedMenu.value && run(() => props.controller.updateMenu(editedMenu.value!))
const selectUser = (checked: boolean, row: typeof users.value.rows[number]) => {
  const selected = new Set(users.value.selectedKeys)
  if (checked) selected.add(row.id); else selected.delete(row.id)
  props.controller.users.select(users.value.rows.filter((candidate) => selected.has(candidate.id))); revision.value += 1
}
onMounted(() => run(async () => {
  await props.controller.refreshAuthorizationData()
  if (can('iam.users.read')) await props.controller.users.refresh()
  if (!can(`iam.${tab.value}.read`)) tab.value = can('iam.roles.read') ? 'roles' : 'menus'
}))
onBeforeUnmount(() => { user.password = ''; replacementPassword.value = '' })
</script>

<template>
  <main class="administration-page">
    <header><h1>用户与权限</h1></header>
    <p v-if="failureMessage" role="alert">{{ failureMessage }}</p>
    <button v-if="controller.hasPendingRepair()" type="button" data-testid="repair-projection" :disabled="controller.busy" @click="run(() => controller.repairProjection())">刷新已保存数据</button>
    <nav class="tabs" aria-label="用户与权限管理视图">
      <button v-for="value in (['users', 'roles', 'menus'] as const).filter((item) => can(`iam.${item}.read`))" :key="value" type="button" :aria-pressed="tab === value" @click="tab = value">{{ value === 'users' ? '用户管理' : value === 'roles' ? '角色管理' : '菜单管理' }}</button>
    </nav>

    <section v-if="tab === 'users' && can('iam.users.read')" aria-labelledby="users-heading">
      <h2 id="users-heading">用户管理</h2>
      <form class="toolbar" data-testid="user-search" @submit.prevent="search">
        <label>关键字<input v-model.trim="filters.search" maxlength="100" placeholder="用户名、姓名或邮箱"></label>
        <button type="submit">查询</button><button type="button" data-testid="user-search-reset" @click="reset">重置</button>
        <button v-if="can('iam.users.delete')" type="button" data-testid="delete-selected-users" :disabled="users.selectedKeys.length === 0 || controller.deleteUsers.busy || mutationBlocked" @click="run(() => controller.deleteUsers.run(users.selectedKeys))">删除所选</button>
      </form>
      <table><thead><tr><th scope="col">选择</th><th scope="col">用户名</th><th scope="col">姓名</th><th scope="col">邮箱</th><th scope="col">状态</th><th scope="col">操作</th></tr></thead>
        <tbody><tr v-for="row in users.rows" :key="row.id" :data-row-key="row.username"><td><input v-if="can('iam.users.delete')" type="checkbox" :checked="users.selectedKeys.includes(row.id)" :aria-label="`选择 ${row.username}`" @change="selectUser(($event.target as HTMLInputElement).checked, row)"></td><td>{{ row.username }}</td><td>{{ row.displayName }}</td><td>{{ row.email }}</td><td>{{ row.disabled ? '停用' : '启用' }}</td><td><button v-if="can('iam.users.write')" type="button" data-action="edit" @click="editUser(row)">编辑</button><button v-if="can('iam.users.write')" type="button" data-action="toggle" :disabled="mutationBlocked" @click="toggleUser(row)">{{ row.disabled ? '启用' : '停用' }}</button><button v-if="can('iam.users.delete')" type="button" data-action="delete" :disabled="controller.deleteUsers.busy || mutationBlocked" @click="deleteUser(row)">删除</button></td></tr></tbody>
      </table>
      <div class="pagination"><button type="button" :disabled="users.page <= 1" @click="run(() => controller.users.setPage(users.page - 1))">上一页</button><span>第 {{ users.page }} 页</span><button type="button" :disabled="users.page * users.pageSize >= users.total" @click="run(() => controller.users.setPage(users.page + 1))">下一页</button></div>
      <form v-if="can('iam.users.write')" class="editor" data-testid="create-user" @submit.prevent="submitUser"><h3>新增用户</h3><label>用户名<input name="username" v-model.trim="user.username" required minlength="3" maxlength="64"></label><label>姓名<input name="displayName" v-model.trim="user.displayName" required maxlength="80"></label><label>邮箱<input name="email" v-model.trim="user.email" type="email" required></label><label>密码<input name="password" v-model="user.password" type="password" required minlength="12" autocomplete="new-password"></label><button type="submit" :disabled="controller.createUser.busy || mutationBlocked">新增用户</button></form>
      <form v-if="editedUser && can('iam.users.write')" class="editor" data-testid="edit-user" @submit.prevent="saveUser"><h3>编辑 {{ editedUser.username }}</h3><label>姓名<input name="displayName" v-model.trim="editedUser.displayName" required maxlength="80"></label><label>邮箱<input name="email" v-model.trim="editedUser.email" type="email" required></label><button type="submit" :disabled="mutationBlocked">保存用户</button></form>
      <form v-if="editedUser && can('iam.roles.assign')" class="editor" data-testid="assign-user-roles" @submit.prevent="saveUserRoles"><h3>分配角色</h3><label v-for="item in roles" :key="item.id" class="choice"><input v-model="assignedRoles" type="checkbox" :value="item.id" :data-role-key="item.key">{{ item.name }}</label><button type="submit" :disabled="mutationBlocked">保存角色分配</button></form>
      <form v-if="editedUser && can('iam.users.reset-password')" class="editor" data-testid="reset-user-password" @submit.prevent="resetPassword"><h3>重置密码</h3><label>新密码<input name="password" v-model="replacementPassword" type="password" required minlength="12" autocomplete="new-password"></label><button type="submit" :disabled="mutationBlocked">重置密码</button></form>
    </section>

    <section v-else-if="tab === 'roles' && can('iam.roles.read')" aria-labelledby="roles-heading">
      <h2 id="roles-heading">角色管理</h2>
      <table><thead><tr><th>角色标识</th><th>角色名称</th><th>数据范围</th><th>状态</th><th>操作</th></tr></thead><tbody><tr v-for="item in roles" :key="item.id" :data-row-key="item.key"><td>{{ item.key }}</td><td>{{ item.name }}</td><td>{{ item.dataScope === 'self' ? '本人' : '全部' }}</td><td>{{ item.enabled ? '启用' : '停用' }}</td><td><button v-if="can('iam.roles.write')" type="button" data-action="edit" @click="editRole(item)">编辑</button><button v-if="can('iam.roles.delete')" type="button" data-action="delete" :disabled="item.protected || mutationBlocked" @click="run(() => controller.deleteRole(item.id))">删除</button></td></tr></tbody></table>
      <form v-if="can('iam.roles.write')" class="editor" data-testid="create-role" @submit.prevent="submitRole"><h3>新增角色</h3><label>角色标识<input name="key" v-model.trim="role.key" required minlength="3" maxlength="64" pattern="[a-z0-9][a-z0-9_-]*"></label><label>角色名称<input name="name" v-model.trim="role.name" required maxlength="100"></label><label>数据范围<select name="dataScope" v-model="role.dataScope"><option value="self">本人</option><option value="all">全部</option></select></label><button type="submit" :disabled="controller.createRole.busy || mutationBlocked">新增角色</button></form>
      <form v-if="editedRole && can('iam.roles.write')" class="editor" data-testid="edit-role" @submit.prevent="saveRole"><h3>编辑 {{ editedRole.key }}</h3><label>角色名称<input name="name" v-model.trim="editedRole.name" required maxlength="100"></label><label>数据范围<select name="dataScope" v-model="editedRole.dataScope"><option value="self">本人</option><option value="all">全部</option></select></label><label class="choice"><input name="enabled" v-model="editedRole.enabled" type="checkbox">启用</label><button type="submit" :disabled="editedRole.protected || mutationBlocked">保存角色</button></form>
      <form v-if="editedRole && can('iam.roles.assign')" class="editor" data-testid="assign-role-grants" @submit.prevent="saveRoleGrants"><h3>分配权限与菜单</h3><fieldset><legend>权限标识</legend><label v-for="item in controller.permissions()" :key="item.code" class="choice"><input v-model="grantedPermissions" type="checkbox" :value="item.code" :data-permission-code="item.code">{{ item.name }}</label></fieldset><fieldset><legend>菜单</legend><label v-for="item in menus" :key="item.id" class="choice"><input v-model="grantedMenus" type="checkbox" :value="item.id" :data-menu-key="item.key">{{ item.label }}</label></fieldset><button type="submit" :disabled="editedRole.protected || mutationBlocked">保存权限分配</button></form>
    </section>

    <section v-else-if="tab === 'menus' && can('iam.menus.read')" aria-labelledby="menus-heading">
      <h2 id="menus-heading">菜单管理</h2>
      <table><thead><tr><th>菜单标识</th><th>菜单名称</th><th>路由地址</th><th>权限标识</th><th>操作</th></tr></thead><tbody><tr v-for="item in menus" :key="item.id" :data-row-key="item.key"><td>{{ item.key }}</td><td>{{ item.label }}</td><td>{{ item.path }}</td><td>{{ item.permissionCode }}</td><td><button v-if="can('iam.menus.write')" type="button" data-action="edit" @click="editMenu(item)">编辑</button><button v-if="can('iam.menus.delete')" type="button" data-action="delete" :disabled="item.protected || mutationBlocked" @click="run(() => controller.deleteMenu(item.id))">删除</button></td></tr></tbody></table>
      <form v-if="can('iam.menus.write')" class="editor" data-testid="create-menu" @submit.prevent="submitMenu"><h3>新增菜单</h3><label>菜单标识<input name="key" v-model.trim="menu.key" required minlength="3" maxlength="64" pattern="[a-z0-9][a-z0-9_-]*"></label><label>菜单名称<input name="label" v-model.trim="menu.label" required maxlength="80"></label><label>路由地址<input name="path" v-model.trim="menu.path" required pattern="/[a-z0-9/_-]+"></label><label>权限标识<input name="permissionCode" v-model.trim="menu.permissionCode" required minlength="3"></label><label>显示顺序<input name="sortOrder" v-model.number="menu.sortOrder" type="number" min="0" max="100000"></label><button type="submit" :disabled="controller.createMenu.busy || mutationBlocked">新增菜单</button></form>
      <form v-if="editedMenu && can('iam.menus.write')" class="editor" data-testid="edit-menu" @submit.prevent="saveMenu"><h3>编辑 {{ editedMenu.key }}</h3><label>菜单名称<input name="label" v-model.trim="editedMenu.label" required maxlength="80"></label><label>路由地址<input name="path" v-model.trim="editedMenu.path" required pattern="/[a-z0-9/_-]+"></label><label>权限标识<input name="permissionCode" v-model.trim="editedMenu.permissionCode" required minlength="3"></label><label>显示顺序<input name="sortOrder" v-model.number="editedMenu.sortOrder" type="number" min="0" max="100000"></label><button type="submit" :disabled="editedMenu.protected || mutationBlocked">保存菜单</button></form>
    </section>
    <section v-else aria-live="polite"><p>当前账号没有可用的管理视图。</p></section>
  </main>
</template>

<style scoped>
.administration-page { display: grid; gap: 20px; max-width: 1200px; padding: 24px; }
h1, h2, h3 { margin: 0; letter-spacing: 0; }
.tabs, .toolbar, .pagination { display: flex; align-items: end; gap: 8px; flex-wrap: wrap; }
.tabs button[aria-pressed="true"] { border-bottom-color: var(--ga-brand); color: var(--ga-brand); }
section, .editor { display: grid; gap: 16px; }
.editor { grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); align-items: end; border-top: 1px solid var(--ga-border-light); padding-top: 16px; }
.editor h3 { grid-column: 1 / -1; }
label { display: grid; gap: 6px; }
input, select, button { min-height: 40px; font: inherit; }
button { border: 1px solid var(--ga-border); background: var(--ga-bg-container); padding: 6px 12px; }
button:disabled { opacity: .5; }
table { width: 100%; border-collapse: collapse; }
th, td { border-bottom: 1px solid var(--ga-border-light); padding: 10px; text-align: left; }
th { color: var(--ga-text-2); font-size: 14px; }
@media (max-width: 640px) { .administration-page { padding: 16px; overflow-x: auto; } table { min-width: 680px; } }
</style>
