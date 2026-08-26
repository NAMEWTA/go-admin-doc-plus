<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { validateProduct, type Product, type ProductInput } from '@go-admin/domain-demo'
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

const settle = async (operation: () => Promise<unknown>) => {
  try { await operation() } catch { /* controller keeps the stable failure */ }
  finally {
    revision.value += 1
    if (props.controller.failure() === 'relogin') emit('sessionRequired')
    if (props.controller.failure() === 'forbidden') emit('forbidden')
  }
}
const resetForm = () => { editing.value = null; formError.value = ''; Object.assign(form, props.controller.empty()); delete form.id; delete form.revision }
const edit = (product: Product) => { editing.value = product; formError.value = ''; Object.assign(form, { sku: product.sku, name: product.name, description: product.description, priceCents: product.priceCents, status: product.status, id: product.id, revision: product.revision }) }
const save = () => settle(async () => {
  if (Object.keys(validateProduct(form)).length > 0) { formError.value = 'Check the highlighted product fields.'; return }
  formError.value = ''
  const result = await props.controller.save({ ...form }); if (result === 'completed') resetForm()
})
const remove = (products: ReadonlyArray<Product>) => settle(() => props.controller.remove(products))
const selected = computed(() => snapshot.value.rows.filter(row => snapshot.value.selectedKeys.includes(row.id)))
const toggle = (product: Product, checked: boolean) => {
  const ids = new Set(snapshot.value.selectedKeys)
  if (checked) ids.add(product.id)
  else ids.delete(product.id)
  props.controller.list.select(snapshot.value.rows.filter(row => ids.has(row.id))); revision.value += 1
}
onMounted(() => settle(() => props.controller.list.refresh()))
</script>

<template>
  <section class="demo-products" aria-labelledby="demo-products-title">
    <header class="demo-products__header">
      <div><h1 id="demo-products-title">Products</h1><p>{{ snapshot.total }} records</p></div>
      <button v-if="controller.pendingRepair" type="button" :disabled="controller.busy" data-testid="repair" @click="settle(() => controller.repairProjection())">Refresh results</button>
    </header>
    <p v-if="failure" role="alert" :data-failure="failure">{{ failure }}</p>
    <form class="demo-products__search" @submit.prevent="settle(() => controller.list.search({ search }))">
      <label>Search <input v-model="search" name="search" maxlength="100"></label>
      <button type="submit" :disabled="blocked">Search</button>
      <button type="button" :disabled="blocked" @click="search = ''; settle(() => controller.list.reset())">Reset</button>
    </form>
    <div class="demo-products__grid">
      <div class="demo-products__table">
        <div class="demo-products__actions"><button type="button" :disabled="blocked || selected.length === 0" @click="remove(selected)">Delete selected</button></div>
        <table><thead><tr><th aria-label="Select"></th><th><button type="button" :disabled="blocked" @click="settle(() => controller.list.setSort({ key: 'sku', direction: 'ascending' }))">SKU</button></th><th><button type="button" :disabled="blocked" @click="settle(() => controller.list.setSort({ key: 'name', direction: 'ascending' }))">Name</button></th><th><button type="button" :disabled="blocked" @click="settle(() => controller.list.setSort({ key: 'priceCents', direction: 'ascending' }))">Price</button></th><th>Status</th><th>Actions</th></tr></thead>
          <tbody><tr v-for="product in snapshot.rows" :key="product.id">
            <td><input type="checkbox" :checked="snapshot.selectedKeys.includes(product.id)" :aria-label="`Select ${product.sku}`" @change="toggle(product, ($event.target as HTMLInputElement).checked)"></td>
            <td>{{ product.sku }}</td><td>{{ product.name }}</td><td>{{ product.priceCents }}</td><td>{{ product.status }}</td>
            <td><button type="button" :disabled="blocked" @click="edit(product)">Edit</button><button type="button" :disabled="blocked" @click="remove([product])">Delete</button></td>
          </tr></tbody>
        </table>
        <nav aria-label="Pagination"><button type="button" :disabled="blocked || snapshot.page <= 1" @click="settle(() => controller.list.setPage(snapshot.page - 1))">Previous</button><span>Page {{ snapshot.page }}</span><label>Rows <select :value="snapshot.pageSize" :disabled="blocked" @change="settle(() => controller.list.setPageSize(Number(($event.target as HTMLSelectElement).value)))"><option :value="10">10</option><option :value="20">20</option><option :value="50">50</option></select></label><button type="button" :disabled="blocked || snapshot.page * snapshot.pageSize >= snapshot.total" @click="settle(() => controller.list.setPage(snapshot.page + 1))">Next</button></nav>
      </div>
      <form class="demo-products__form" @submit.prevent="save">
        <h2>{{ editing ? 'Edit product' : 'Create product' }}</h2>
        <p v-if="formError" role="alert">{{ formError }}</p>
        <label>SKU <input v-model="form.sku" name="sku" maxlength="32" required></label>
        <label>Name <input v-model="form.name" name="name" maxlength="120" required></label>
        <label>Description <textarea v-model="form.description" name="description" maxlength="500"></textarea></label>
        <label>Price <input v-model.number="form.priceCents" name="priceCents" type="number" min="0" max="100000000" required></label>
        <label>Status <select v-model="form.status" name="status"><option value="active">Active</option><option value="inactive">Inactive</option></select></label>
        <div><button type="submit" :disabled="blocked">Save</button><button type="button" :disabled="controller.busy" @click="resetForm">Cancel</button></div>
      </form>
    </div>
  </section>
</template>

<style scoped>
.demo-products { display: grid; gap: 16px; color: #17202a; }
.demo-products__header, .demo-products__actions, nav { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
h1, h2, p { margin: 0; }
.demo-products__search { display: flex; align-items: end; gap: 8px; }
.demo-products__grid { display: grid; grid-template-columns: minmax(0, 1fr) minmax(260px, 340px); gap: 16px; }
.demo-products__table { min-width: 0; overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 8px; border-bottom: 1px solid #dfe6e9; text-align: left; }
.demo-products__form { display: grid; align-content: start; gap: 10px; padding-left: 16px; border-left: 1px solid #dfe6e9; }
label { display: grid; gap: 4px; }
input, textarea, select, button { font: inherit; }
button { min-height: 34px; }
[role="alert"] { padding: 8px; border-left: 3px solid #b42318; background: #fff1f0; }
@media (max-width: 720px) { .demo-products__grid { grid-template-columns: 1fr; } .demo-products__form { padding-left: 0; border-left: 0; border-top: 1px solid #dfe6e9; padding-top: 16px; } .demo-products__search { flex-wrap: wrap; } }
</style>
