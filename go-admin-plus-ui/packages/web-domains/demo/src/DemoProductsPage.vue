<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { demoPermissions, validateProduct, type Product, type ProductInput } from '@go-admin/domain-demo'
import type { DemoController } from './demo-controller'

const props = defineProps<{ controller: DemoController }>()
const emit = defineEmits<{ sessionRequired: []; forbidden: [] }>()
const revision = ref(0)
const search = ref('')
const formOpen = ref(false)
const editing = ref<Product | null>(null)
const formError = ref('')
const form = reactive<ProductInput & { id?: string; revision?: number }>(props.controller.empty())
const snapshot = computed(() => { void revision.value; return props.controller.list.snapshot() })
const failure = computed(() => { void revision.value; return props.controller.failure() })
const failureMessage = computed(() => {
  if (failure.value === 'relogin') return '会话已失效，请重新登录。'
  if (failure.value === 'forbidden') return '没有执行该产品操作的权限。'
  if (failure.value === 'validation') return '请检查产品表单。'
  if (failure.value === 'conflict') return '产品数据已发生变化，请刷新后重试。'
  if (failure.value === 'unavailable') return '产品服务暂不可用。'
  return ''
})
const blocked = computed(() => { void revision.value; return props.controller.busy || props.controller.pendingRepair })
const projectionVisible = computed(() => { void revision.value; return props.controller.projectionVisible })
const canRead = computed(() => { void revision.value; return props.controller.can(demoPermissions.read) })
const canWrite = computed(() => { void revision.value; return props.controller.can(demoPermissions.write) })
const canDelete = computed(() => { void revision.value; return props.controller.can(demoPermissions.delete) })

const settle = async (operation: () => Promise<unknown>) => {
  try { await operation() } catch { /* controller keeps the stable failure */ }
  finally {
    if (props.controller.takeCompletion() === 'save') { resetForm(); formOpen.value = false }
    revision.value += 1
    if (props.controller.failure() === 'relogin') emit('sessionRequired')
    if (props.controller.failure() === 'forbidden') emit('forbidden')
  }
}
const resetForm = () => { editing.value = null; formError.value = ''; Object.assign(form, props.controller.empty()); delete form.id; delete form.revision }
const closeForm = () => { formOpen.value = false; resetForm() }
const create = () => { resetForm(); formOpen.value = true }
const edit = (product: Product) => { editing.value = product; formError.value = ''; Object.assign(form, { sku: product.sku, name: product.name, description: product.description, priceCents: product.priceCents, status: product.status, id: product.id, revision: product.revision }); formOpen.value = true }
const save = () => settle(async () => {
  if (Object.keys(validateProduct(form)).length > 0) { formError.value = '请检查产品表单中的必填项和取值范围。'; return }
  formError.value = ''
  await props.controller.save({ ...form })
})
const remove = (products: ReadonlyArray<Product>) => settle(() => props.controller.remove(products))
const selected = computed(() => snapshot.value.rows.filter(row => snapshot.value.selectedKeys.includes(row.id)))
const toggle = (product: Product, checked: boolean) => {
  const ids = new Set(snapshot.value.selectedKeys)
  if (checked) ids.add(product.id)
  else ids.delete(product.id)
  props.controller.list.select(snapshot.value.rows.filter(row => ids.has(row.id))); revision.value += 1
}
onMounted(() => { void settle(() => props.controller.list.refresh()) })
</script>

<template>
  <section class="demo-products" aria-labelledby="demo-products-title">
    <header class="demo-products__header">
      <div><h1 id="demo-products-title">产品示例</h1><p>共 {{ canRead ? snapshot.total : 0 }} 条</p></div>
      <button v-if="controller.pendingRepair" type="button" :disabled="controller.busy" data-testid="repair" @click="settle(() => controller.repairProjection())">刷新结果</button>
      <button v-else-if="failure === 'unavailable'" type="button" :disabled="controller.busy" data-testid="retry" @click="settle(() => controller.list.refresh())">重试</button>
    </header>
    <p v-if="failure" role="alert" :data-failure="failure">{{ failureMessage }}</p>
    <form v-if="projectionVisible && canRead" class="demo-products__search" @submit.prevent="settle(() => controller.list.search({ search }))">
      <label>产品搜索<input v-model="search" name="search" placeholder="请输入 SKU 或名称"></label>
      <button type="submit" :disabled="blocked">搜索</button>
      <button type="button" :disabled="blocked" @click="search = ''; settle(() => controller.list.reset())">重置</button>
    </form>
    <div v-if="projectionVisible && canRead" class="demo-products__grid">
      <div class="demo-products__table">
        <div v-if="canWrite || canDelete" class="demo-products__actions"><button v-if="canWrite" type="button" data-testid="open-product-form" @click="create">新增</button><button v-if="canDelete" type="button" data-testid="delete-selected-products" :disabled="blocked || selected.length === 0" @click="remove(selected)">批量删除</button></div>
        <table><thead><tr><th aria-label="选择"></th><th><button type="button" :disabled="blocked" @click="settle(() => controller.list.setSort({ key: 'sku', direction: 'ascending' }))">SKU</button></th><th><button type="button" :disabled="blocked" @click="settle(() => controller.list.setSort({ key: 'name', direction: 'ascending' }))">名称</button></th><th><button type="button" :disabled="blocked" @click="settle(() => controller.list.setSort({ key: 'priceCents', direction: 'ascending' }))">价格</button></th><th>状态</th><th>操作</th></tr></thead>
          <tbody><tr v-for="product in snapshot.rows" :key="product.id">
            <td><input v-if="canDelete" type="checkbox" :checked="snapshot.selectedKeys.includes(product.id)" :aria-label="`选择 ${product.sku}`" @change="toggle(product, ($event.target as HTMLInputElement).checked)"></td>
            <td>{{ product.sku }}</td><td>{{ product.name }}</td><td>{{ product.priceCents }}</td><td>{{ product.status === 'active' ? '启用' : '停用' }}</td>
            <td><button v-if="canWrite" type="button" data-action="edit" :disabled="blocked" @click="edit(product)">修改</button><button v-if="canDelete" type="button" data-action="delete" :aria-label="`删除 ${product.sku}`" :disabled="blocked" @click="remove([product])">删除</button></td>
          </tr></tbody>
        </table>
        <nav aria-label="分页"><button type="button" :disabled="blocked || snapshot.page <= 1" @click="settle(() => controller.list.setPage(snapshot.page - 1))">上一页</button><span>第 {{ snapshot.page }} 页</span><label>每页<select :value="snapshot.pageSize" :disabled="blocked" @change="settle(() => controller.list.setPageSize(Number(($event.target as HTMLSelectElement).value)))"><option :value="10">10</option><option :value="20">20</option><option :value="50">50</option></select></label><button type="button" :disabled="blocked || snapshot.page * snapshot.pageSize >= snapshot.total" @click="settle(() => controller.list.setPage(snapshot.page + 1))">下一页</button></nav>
      </div>
    </div>
    <div v-if="formOpen && canWrite" class="management-dialog-backdrop" @click.self="closeForm" @keydown.esc="closeForm"><form class="management-dialog demo-products__form" role="dialog" aria-modal="true" :aria-labelledby="editing ? 'edit-product-title' : 'create-product-title'" @submit.prevent="save"><header class="management-dialog__header"><h2 :id="editing ? 'edit-product-title' : 'create-product-title'">{{ editing ? '修改产品' : '新增产品' }}</h2><button type="button" aria-label="关闭" :disabled="controller.busy" @click="closeForm">×</button></header><div class="management-dialog__body"><p v-if="formError" role="alert">{{ formError }}</p><label>SKU<input v-model="form.sku" name="sku" autofocus maxlength="32" required></label><label>名称<input v-model="form.name" name="name" required></label><label>描述<textarea v-model="form.description" name="description"></textarea></label><label>价格（分）<input v-model.number="form.priceCents" name="priceCents" type="number" min="0" max="100000000" required></label><label>状态<select v-model="form.status" name="status"><option value="active">启用</option><option value="inactive">停用</option></select></label></div><footer class="management-dialog__footer"><button type="button" :disabled="controller.busy" @click="closeForm">取消</button><button type="submit" :disabled="blocked">保存</button></footer></form></div>
  </section>
</template>

<style scoped>
.demo-products { display: grid; gap: 16px; color: var(--ga-text-1); }
.demo-products__header, .demo-products__actions, nav { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
h1, h2, p { margin: 0; }
.demo-products__search { display: flex; align-items: end; gap: 8px; }
.demo-products__grid { display: grid; grid-template-columns: minmax(0, 1fr); gap: 16px; }
.demo-products__table { min-width: 0; overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 8px; border-bottom: 1px solid var(--ga-border-light); text-align: left; }
.demo-products__form { display: grid; align-content: start; }
label { display: grid; gap: 4px; }
input, textarea, select, button { font: inherit; }
button { min-height: 34px; }
[role="alert"] { padding: 8px; border-left: 3px solid var(--ga-danger); background: #fff1f0; }
@media (max-width: 720px) { .demo-products__search { flex-wrap: wrap; } }
</style>
