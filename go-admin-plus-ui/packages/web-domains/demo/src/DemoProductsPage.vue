<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { demoPermissions, validateProduct, type Product, type ProductInput } from '@go-admin/domain-demo'
import type { DemoController } from './demo-controller'

const props = defineProps<{ controller: DemoController }>()
const emit = defineEmits<{ sessionRequired: []; forbidden: [] }>()
const revision = ref(0)
const search = ref('')
const editing = ref<Product | null>(null)
const formError = ref('')
const form = reactive<ProductInput & { id?: string; revision?: number }>(props.controller.empty())
const snapshot = computed(() => { void revision.value; return props.controller.list.snapshot() })
const failure = computed(() => { void revision.value; return props.controller.failure() })
const blocked = computed(() => { void revision.value; return props.controller.busy || props.controller.pendingRepair })
const projectionVisible = computed(() => { void revision.value; return props.controller.projectionVisible })
const canRead = computed(() => { void revision.value; return props.controller.can(demoPermissions.read) })
const canWrite = computed(() => { void revision.value; return props.controller.can(demoPermissions.write) })
const canDelete = computed(() => { void revision.value; return props.controller.can(demoPermissions.delete) })

const settle = async (operation: () => Promise<unknown>) => {
  try { await operation() } catch { /* controller keeps the stable failure */ }
  finally {
    if (props.controller.takeCompletion() === 'save') resetForm()
    revision.value += 1
    if (props.controller.failure() === 'relogin') emit('sessionRequired')
    if (props.controller.failure() === 'forbidden') emit('forbidden')
  }
}
const resetForm = () => { editing.value = null; formError.value = ''; Object.assign(form, props.controller.empty()); delete form.id; delete form.revision }
const edit = (product: Product) => { editing.value = product; formError.value = ''; Object.assign(form, { sku: product.sku, name: product.name, description: product.description, priceCents: product.priceCents, status: product.status, id: product.id, revision: product.revision }) }
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
    <p v-if="failure" role="alert" :data-failure="failure">{{ failure }}</p>
    <form v-if="projectionVisible && canRead" class="demo-products__search" @submit.prevent="settle(() => controller.list.search({ search }))">
      <label>产品搜索<input v-model="search" name="search" placeholder="请输入 SKU 或名称"></label>
      <button type="submit" :disabled="blocked">搜索</button>
      <button type="button" :disabled="blocked" @click="search = ''; settle(() => controller.list.reset())">重置</button>
    </form>
    <div v-if="projectionVisible && canRead" class="demo-products__grid">
      <div class="demo-products__table">
        <div v-if="canDelete" class="demo-products__actions"><button type="button" :disabled="blocked || selected.length === 0" @click="remove(selected)">批量删除</button></div>
        <table><thead><tr><th aria-label="选择"></th><th><button type="button" :disabled="blocked" @click="settle(() => controller.list.setSort({ key: 'sku', direction: 'ascending' }))">SKU</button></th><th><button type="button" :disabled="blocked" @click="settle(() => controller.list.setSort({ key: 'name', direction: 'ascending' }))">名称</button></th><th><button type="button" :disabled="blocked" @click="settle(() => controller.list.setSort({ key: 'priceCents', direction: 'ascending' }))">价格</button></th><th>状态</th><th>操作</th></tr></thead>
          <tbody><tr v-for="product in snapshot.rows" :key="product.id">
            <td><input v-if="canDelete" type="checkbox" :checked="snapshot.selectedKeys.includes(product.id)" :aria-label="`Select ${product.sku}`" @change="toggle(product, ($event.target as HTMLInputElement).checked)"></td>
            <td>{{ product.sku }}</td><td>{{ product.name }}</td><td>{{ product.priceCents }}</td><td>{{ product.status }}</td>
            <td><button v-if="canWrite" type="button" :disabled="blocked" @click="edit(product)">修改</button><button v-if="canDelete" type="button" :aria-label="`删除 ${product.sku}`" :disabled="blocked" @click="remove([product])">删除</button></td>
          </tr></tbody>
        </table>
        <nav aria-label="分页"><button type="button" :disabled="blocked || snapshot.page <= 1" @click="settle(() => controller.list.setPage(snapshot.page - 1))">上一页</button><span>第 {{ snapshot.page }} 页</span><label>每页<select :value="snapshot.pageSize" :disabled="blocked" @change="settle(() => controller.list.setPageSize(Number(($event.target as HTMLSelectElement).value)))"><option :value="10">10</option><option :value="20">20</option><option :value="50">50</option></select></label><button type="button" :disabled="blocked || snapshot.page * snapshot.pageSize >= snapshot.total" @click="settle(() => controller.list.setPage(snapshot.page + 1))">下一页</button></nav>
      </div>
      <form v-if="canWrite" class="demo-products__form" @submit.prevent="save">
        <h2>{{ editing ? '修改产品' : '新增产品' }}</h2>
        <p v-if="formError" role="alert">{{ formError }}</p>
        <label>SKU<input v-model="form.sku" name="sku" maxlength="32" required></label>
        <label>名称<input v-model="form.name" name="name" required></label>
        <label>描述<textarea v-model="form.description" name="description"></textarea></label>
        <label>价格（分）<input v-model.number="form.priceCents" name="priceCents" type="number" min="0" max="100000000" required></label>
        <label>状态<select v-model="form.status" name="status"><option value="active">启用</option><option value="inactive">停用</option></select></label>
        <div><button type="submit" :disabled="blocked">保存</button><button type="button" :disabled="controller.busy" @click="resetForm">取消</button></div>
      </form>
    </div>
  </section>
</template>

<style scoped>
.demo-products { display: grid; gap: 16px; color: var(--ga-text-1); }
.demo-products__header, .demo-products__actions, nav { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
h1, h2, p { margin: 0; }
.demo-products__search { display: flex; align-items: end; gap: 8px; }
.demo-products__grid { display: grid; grid-template-columns: minmax(0, 1fr) minmax(260px, 340px); gap: 16px; }
.demo-products__table { min-width: 0; overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 8px; border-bottom: 1px solid var(--ga-border-light); text-align: left; }
.demo-products__form { display: grid; align-content: start; gap: 10px; padding-left: 16px; border-left: 1px solid var(--ga-border-light); }
label { display: grid; gap: 4px; }
input, textarea, select, button { font: inherit; }
button { min-height: 34px; }
[role="alert"] { padding: 8px; border-left: 3px solid var(--ga-danger); background: #fff1f0; }
@media (max-width: 720px) { .demo-products__grid { grid-template-columns: 1fr; } .demo-products__form { padding-left: 0; border-left: 0; border-top: 1px solid var(--ga-border-light); padding-top: 16px; } .demo-products__search { flex-wrap: wrap; } }
</style>
