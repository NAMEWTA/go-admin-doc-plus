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
const users = computed(() => { void revision.value; return props.controller.users.snapshot() })
const roles = computed(() => { void revision.value; return props.controller.roles() })
const menus = computed(() => { void revision.value; return props.controller.menus() })
const surfaceFailure = () => {
  const failure = props.controller.failure()
  failureMessage.value = failure === 'relogin' ? 'Your session must be renewed.'
    : failure === 'forbidden' ? 'You do not have permission for that action.'
      : failure === 'validation' ? 'Review the submitted values.'
        : failure === 'conflict' ? 'The resource changed or is protected.'
          : failure === 'unavailable' ? 'The administration service is unavailable.' : ''
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
  if (props.controller.can('iam.users.read')) await props.controller.users.refresh()
  if (!props.controller.can(`iam.${tab.value}.read`)) tab.value = props.controller.can('iam.roles.read') ? 'roles' : 'menus'
}))
onBeforeUnmount(() => { user.password = ''; replacementPassword.value = '' })
</script>

<template>
  <main class="administration-page">
    <header><h1>Identity and access</h1></header>
    <p v-if="failureMessage" role="alert">{{ failureMessage }}</p>
    <button v-if="controller.hasPendingRepair()" type="button" data-testid="repair-projection" :disabled="controller.busy" @click="run(() => controller.repairProjection())">Refresh saved changes</button>
    <nav class="tabs" aria-label="IAM administration views">
      <button v-for="value in (['users', 'roles', 'menus'] as const).filter((item) => controller.can(`iam.${item}.read`))" :key="value" type="button" :aria-pressed="tab === value" @click="tab = value">{{ value }}</button>
    </nav>

    <section v-if="tab === 'users' && controller.can('iam.users.read')" aria-labelledby="users-heading">
      <h2 id="users-heading">Users</h2>
      <form class="toolbar" data-testid="user-search" @submit.prevent="search">
        <label>Search<input v-model.trim="filters.search" maxlength="100"></label>
        <button type="submit">Search</button><button type="button" data-testid="user-search-reset" @click="reset">Reset</button>
        <button v-if="controller.can('iam.users.delete')" type="button" data-testid="delete-selected-users" :disabled="users.selectedKeys.length === 0 || controller.deleteUsers.busy" @click="run(() => controller.deleteUsers.run(users.selectedKeys))">Delete selected</button>
      </form>
      <table><thead><tr><th scope="col">Select</th><th scope="col">Username</th><th scope="col">Name</th><th scope="col">Email</th><th scope="col">Status</th><th scope="col">Actions</th></tr></thead>
        <tbody><tr v-for="row in users.rows" :key="row.id" :data-row-key="row.username"><td><input v-if="controller.can('iam.users.delete')" type="checkbox" :checked="users.selectedKeys.includes(row.id)" :aria-label="`Select ${row.username}`" @change="selectUser(($event.target as HTMLInputElement).checked, row)"></td><td>{{ row.username }}</td><td>{{ row.displayName }}</td><td>{{ row.email }}</td><td>{{ row.disabled ? 'Disabled' : 'Enabled' }}</td><td><button v-if="controller.can('iam.users.write')" type="button" data-action="edit" @click="editUser(row)">Edit</button><button v-if="controller.can('iam.users.write')" type="button" data-action="toggle" :disabled="controller.busy" @click="toggleUser(row)">{{ row.disabled ? 'Enable' : 'Disable' }}</button><button v-if="controller.can('iam.users.delete')" type="button" data-action="delete" :disabled="controller.deleteUsers.busy" @click="deleteUser(row)">Delete</button></td></tr></tbody>
      </table>
      <div class="pagination"><button type="button" :disabled="users.page <= 1" @click="run(() => controller.users.setPage(users.page - 1))">Previous</button><span>Page {{ users.page }}</span><button type="button" :disabled="users.page * users.pageSize >= users.total" @click="run(() => controller.users.setPage(users.page + 1))">Next</button></div>
      <form v-if="controller.can('iam.users.write')" class="editor" data-testid="create-user" @submit.prevent="submitUser"><h3>Create user</h3><label>Username<input name="username" v-model.trim="user.username" required minlength="3" maxlength="64"></label><label>Display name<input name="displayName" v-model.trim="user.displayName" required maxlength="80"></label><label>Email<input name="email" v-model.trim="user.email" type="email" required></label><label>Password<input name="password" v-model="user.password" type="password" required minlength="12" autocomplete="new-password"></label><button type="submit" :disabled="controller.createUser.busy">Create user</button></form>
      <form v-if="editedUser && controller.can('iam.users.write')" class="editor" data-testid="edit-user" @submit.prevent="saveUser"><h3>Edit {{ editedUser.username }}</h3><label>Display name<input name="displayName" v-model.trim="editedUser.displayName" required maxlength="80"></label><label>Email<input name="email" v-model.trim="editedUser.email" type="email" required></label><button type="submit" :disabled="controller.busy">Save user</button></form>
      <form v-if="editedUser && controller.can('iam.roles.assign')" class="editor" data-testid="assign-user-roles" @submit.prevent="saveUserRoles"><h3>Assign roles</h3><label v-for="item in roles" :key="item.id" class="choice"><input v-model="assignedRoles" type="checkbox" :value="item.id" :data-role-key="item.key">{{ item.name }}</label><button type="submit" :disabled="controller.busy">Save role assignments</button></form>
      <form v-if="editedUser && controller.can('iam.users.reset-password')" class="editor" data-testid="reset-user-password" @submit.prevent="resetPassword"><h3>Reset password</h3><label>Replacement password<input name="password" v-model="replacementPassword" type="password" required minlength="12" autocomplete="new-password"></label><button type="submit" :disabled="controller.busy">Reset password</button></form>
    </section>

    <section v-else-if="tab === 'roles' && controller.can('iam.roles.read')" aria-labelledby="roles-heading">
      <h2 id="roles-heading">Roles</h2>
      <table><thead><tr><th>Key</th><th>Name</th><th>Data scope</th><th>Status</th><th>Actions</th></tr></thead><tbody><tr v-for="item in roles" :key="item.id" :data-row-key="item.key"><td>{{ item.key }}</td><td>{{ item.name }}</td><td>{{ item.dataScope }}</td><td>{{ item.enabled ? 'Enabled' : 'Disabled' }}</td><td><button v-if="controller.can('iam.roles.write')" type="button" data-action="edit" @click="editRole(item)">Edit</button><button v-if="controller.can('iam.roles.delete')" type="button" data-action="delete" :disabled="item.protected || controller.busy" @click="run(() => controller.deleteRole(item.id))">Delete</button></td></tr></tbody></table>
      <form v-if="controller.can('iam.roles.write')" class="editor" data-testid="create-role" @submit.prevent="submitRole"><h3>Create role</h3><label>Key<input name="key" v-model.trim="role.key" required minlength="3" maxlength="64" pattern="[a-z0-9][a-z0-9_-]*"></label><label>Name<input name="name" v-model.trim="role.name" required maxlength="100"></label><label>Data scope<select name="dataScope" v-model="role.dataScope"><option value="self">Self</option><option value="all">All</option></select></label><button type="submit" :disabled="controller.createRole.busy">Create role</button></form>
      <form v-if="editedRole && controller.can('iam.roles.write')" class="editor" data-testid="edit-role" @submit.prevent="saveRole"><h3>Edit {{ editedRole.key }}</h3><label>Name<input name="name" v-model.trim="editedRole.name" required maxlength="100"></label><label>Data scope<select name="dataScope" v-model="editedRole.dataScope"><option value="self">Self</option><option value="all">All</option></select></label><label class="choice"><input name="enabled" v-model="editedRole.enabled" type="checkbox">Enabled</label><button type="submit" :disabled="editedRole.protected || controller.busy">Save role</button></form>
      <form v-if="editedRole && controller.can('iam.roles.assign')" class="editor" data-testid="assign-role-grants" @submit.prevent="saveRoleGrants"><h3>Permission and menu grants</h3><fieldset><legend>Permission Codes</legend><label v-for="item in controller.permissions()" :key="item.code" class="choice"><input v-model="grantedPermissions" type="checkbox" :value="item.code" :data-permission-code="item.code">{{ item.name }}</label></fieldset><fieldset><legend>Menus</legend><label v-for="item in menus" :key="item.id" class="choice"><input v-model="grantedMenus" type="checkbox" :value="item.id" :data-menu-key="item.key">{{ item.label }}</label></fieldset><button type="submit" :disabled="editedRole.protected || controller.busy">Save grants</button></form>
    </section>

    <section v-else-if="tab === 'menus' && controller.can('iam.menus.read')" aria-labelledby="menus-heading">
      <h2 id="menus-heading">Menus</h2>
      <table><thead><tr><th>Key</th><th>Label</th><th>Path</th><th>Permission</th><th>Actions</th></tr></thead><tbody><tr v-for="item in menus" :key="item.id" :data-row-key="item.key"><td>{{ item.key }}</td><td>{{ item.label }}</td><td>{{ item.path }}</td><td>{{ item.permissionCode }}</td><td><button v-if="controller.can('iam.menus.write')" type="button" data-action="edit" @click="editMenu(item)">Edit</button><button v-if="controller.can('iam.menus.delete')" type="button" data-action="delete" :disabled="item.protected || controller.busy" @click="run(() => controller.deleteMenu(item.id))">Delete</button></td></tr></tbody></table>
      <form v-if="controller.can('iam.menus.write')" class="editor" data-testid="create-menu" @submit.prevent="submitMenu"><h3>Create menu</h3><label>Key<input name="key" v-model.trim="menu.key" required minlength="3" maxlength="64" pattern="[a-z0-9][a-z0-9_-]*"></label><label>Label<input name="label" v-model.trim="menu.label" required maxlength="80"></label><label>Path<input name="path" v-model.trim="menu.path" required pattern="/[a-z0-9/_-]+"></label><label>Permission code<input name="permissionCode" v-model.trim="menu.permissionCode" required minlength="3"></label><label>Order<input name="sortOrder" v-model.number="menu.sortOrder" type="number" min="0" max="100000"></label><button type="submit" :disabled="controller.createMenu.busy">Create menu</button></form>
      <form v-if="editedMenu && controller.can('iam.menus.write')" class="editor" data-testid="edit-menu" @submit.prevent="saveMenu"><h3>Edit {{ editedMenu.key }}</h3><label>Label<input name="label" v-model.trim="editedMenu.label" required maxlength="80"></label><label>Path<input name="path" v-model.trim="editedMenu.path" required pattern="/[a-z0-9/_-]+"></label><label>Permission code<input name="permissionCode" v-model.trim="editedMenu.permissionCode" required minlength="3"></label><label>Order<input name="sortOrder" v-model.number="editedMenu.sortOrder" type="number" min="0" max="100000"></label><button type="submit" :disabled="editedMenu.protected || controller.busy">Save menu</button></form>
    </section>
    <section v-else aria-live="polite"><p>No administration view is available.</p></section>
  </main>
</template>

<style scoped>
.administration-page { display: grid; gap: 20px; max-width: 1200px; padding: 24px; }
h1, h2, h3 { margin: 0; letter-spacing: 0; }
.tabs, .toolbar, .pagination { display: flex; align-items: end; gap: 8px; flex-wrap: wrap; }
.tabs button[aria-pressed="true"] { border-bottom-color: #176b54; color: #176b54; }
section, .editor { display: grid; gap: 16px; }
.editor { grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); align-items: end; border-top: 1px solid #d7dce2; padding-top: 16px; }
.editor h3 { grid-column: 1 / -1; }
label { display: grid; gap: 6px; }
input, select, button { min-height: 40px; font: inherit; }
button { border: 1px solid #aab2bc; background: #fff; padding: 6px 12px; }
button:disabled { opacity: .5; }
table { width: 100%; border-collapse: collapse; }
th, td { border-bottom: 1px solid #d7dce2; padding: 10px; text-align: left; }
th { color: #4c5968; font-size: 14px; }
@media (max-width: 640px) { .administration-page { padding: 16px; overflow-x: auto; } table { min-width: 680px; } }
</style>
