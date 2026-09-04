<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { demoPermissions, validateProduct, type Product, type ProductInput, type ProductQuery } from '@go-admin-plus/domain-demo'
import { AppPage, EmptyState, FormDialog, FormGrid, Pagination, QueryBar, StatusTag, TableToolbar } from '@go-admin-plus/ui/components'
import type { DemoController } from './demo-controller'

const props = defineProps<{ controller: DemoController }>()
const emit = defineEmits<{ sessionRequired: []; forbidden: [] }>()
const revision = ref(0), search = ref(''), formOpen = ref(false), editing = ref<Product | null>(null), formError = ref('')
const form = reactive<ProductInput & { id?: string; revision?: number }>(props.controller.empty())
const snapshot = computed(() => { void revision.value; return props.controller.list.snapshot() })
const failure = computed(() => { void revision.value; return props.controller.failure() })
const failureReference = computed(() => { void revision.value; return props.controller.failureTraceId() })
const failureMessage = computed(() => {
  const current = failure.value
  return current ? ({ relogin: '会话已失效，请重新登录。', forbidden: '没有执行该产品操作的权限。', validation: '请检查产品表单。', conflict: '产品数据已发生变化，请刷新后重试。', 'not-found': '产品已不存在，请刷新列表。', unavailable: '产品服务暂不可用。' } as const)[current] : ''
})
const blocked = computed(() => { void revision.value; return props.controller.busy || props.controller.pendingRepair })
const projectionVisible = computed(() => { void revision.value; return props.controller.projectionVisible })
const canRead = computed(() => { void revision.value; return props.controller.can(demoPermissions.read) })
const canWrite = computed(() => { void revision.value; return props.controller.can(demoPermissions.write) })
const canDelete = computed(() => { void revision.value; return props.controller.can(demoPermissions.delete) })
const pageBusy = computed(() => snapshot.value.loading && !projectionVisible.value)
const settle = async (operation: () => Promise<unknown>) => {
  try { await operation() } catch {}
  finally { if (props.controller.takeCompletion() === 'save') { resetForm(); formOpen.value = false }; revision.value += 1; if (props.controller.failure() === 'relogin') emit('sessionRequired'); if (props.controller.failure() === 'forbidden') emit('forbidden') }
}
const refresh = () => settle(() => props.controller.pendingRepair ? props.controller.repairProjection() : props.controller.list.refresh())
const searchProducts = () => settle(() => props.controller.list.search({ search: search.value }))
const resetSearch = () => { search.value = ''; void settle(() => props.controller.list.reset()) }
const resetForm = () => { editing.value = null; formError.value = ''; Object.assign(form, props.controller.empty()); delete form.id; delete form.revision }
const closeForm = () => { formOpen.value = false; resetForm() }
const create = () => { resetForm(); formOpen.value = true }
const edit = (product: Product) => { editing.value = product; formError.value = ''; Object.assign(form, { sku: product.sku, name: product.name, description: product.description, priceCents: product.priceCents, status: product.status, id: product.id, revision: product.revision }); formOpen.value = true }
const save = () => settle(async () => { if (Object.keys(validateProduct(form)).length > 0) { formError.value = '请检查产品表单中的必填项和取值范围。'; return }; formError.value = ''; await props.controller.save({ ...form }) })
const remove = (products: ReadonlyArray<Product>) => settle(() => props.controller.remove(products))
const selected = computed(() => snapshot.value.rows.filter(row => snapshot.value.selectedKeys.includes(row.id)))
type ProductSortKey = ProductQuery['sort']
const currentSort = computed(() => snapshot.value.sort ?? { key: 'updatedAt', direction: 'descending' as const })
const sortDirection = (key: ProductSortKey) => currentSort.value.key === key ? currentSort.value.direction : 'none'
const sortMarker = (key: ProductSortKey) => currentSort.value.key === key ? currentSort.value.direction === 'ascending' ? '↑' : '↓' : ''
const sortBy = (key: ProductSortKey) => settle(() => props.controller.list.setSort({ key, direction: currentSort.value.key === key && currentSort.value.direction === 'ascending' ? 'descending' : 'ascending' }))
const toggle = (product: Product, checked: boolean) => { const ids = new Set(snapshot.value.selectedKeys); if (checked) ids.add(product.id); else ids.delete(product.id); props.controller.list.select(snapshot.value.rows.filter(row => ids.has(row.id))); revision.value += 1 }
onMounted(() => { void settle(() => props.controller.list.refresh()) })
</script>

<template>
  <AppPage title="产品示例" description="标准列表、表单、权限和并发冲突参考" :busy="pageBusy">
    <template #actions><StatusTag tone="info" :label="`${canRead ? snapshot.total : 0} 条记录`" /></template>
    <p v-if="failure" class="page-alert" role="alert" :data-failure="failure">{{ failureMessage }}<span v-if="failureReference"> 参考编号：{{ failureReference }}</span></p>
    <template v-if="projectionVisible && canRead">
      <QueryBar :busy="blocked" :reset-disabled="!search" @search="searchProducts" @reset="resetSearch"><label>产品搜索<input v-model="search" name="search" maxlength="100" placeholder="请输入 SKU 或名称"></label></QueryBar>
      <TableToolbar :selected-count="selected.length" :busy="blocked" @refresh="refresh"><button v-if="canWrite" type="button" data-testid="open-product-form" :disabled="blocked" @click="create">新增产品</button><button v-if="canDelete" type="button" data-testid="delete-selected-products" :disabled="blocked || selected.length === 0" @click="remove(selected)">批量删除</button></TableToolbar>
      <div class="table-scroll" role="region" aria-label="产品列表">
        <table v-if="snapshot.rows.length"><thead><tr><th v-if="canDelete" aria-label="选择"></th><th :aria-sort="sortDirection('sku')"><button type="button" :disabled="blocked" @click="sortBy('sku')">SKU {{ sortMarker('sku') }}</button></th><th :aria-sort="sortDirection('name')"><button type="button" :disabled="blocked" @click="sortBy('name')">名称 {{ sortMarker('name') }}</button></th><th :aria-sort="sortDirection('priceCents')"><button type="button" :disabled="blocked" @click="sortBy('priceCents')">价格（分） {{ sortMarker('priceCents') }}</button></th><th>状态</th><th>操作</th></tr></thead><tbody><tr v-for="product in snapshot.rows" :key="product.id"><td v-if="canDelete"><input type="checkbox" :checked="snapshot.selectedKeys.includes(product.id)" :aria-label="`选择 ${product.sku}`" :disabled="blocked" @change="toggle(product, ($event.target as HTMLInputElement).checked)"></td><td>{{ product.sku }}</td><td>{{ product.name }}</td><td>{{ product.priceCents }}</td><td><StatusTag :tone="product.status === 'active' ? 'success' : 'neutral'" :label="product.status === 'active' ? '启用' : '停用'" /></td><td class="row-actions"><button v-if="canWrite" type="button" data-action="edit" :disabled="blocked" @click="edit(product)">修改</button><button v-if="canDelete" type="button" data-action="delete" :aria-label="`删除 ${product.sku}`" :disabled="blocked" @click="remove([product])">删除</button></td></tr></tbody></table>
        <EmptyState v-else title="暂无产品" :action-label="canWrite ? '新增产品' : undefined" @action="create" />
      </div>
      <Pagination :page="snapshot.page" :page-size="snapshot.pageSize" :total="snapshot.total" :disabled="blocked" @update:page="settle(() => controller.list.setPage($event))" @update:page-size="settle(() => controller.list.setPageSize($event))" />
    </template>
    <EmptyState v-else-if="!pageBusy" :title="failureMessage || '暂无可用产品数据'" action-label="重试" @action="refresh" />
    <FormDialog :model-value="formOpen && canWrite" :title="editing ? '修改产品' : '新增产品'" :busy="controller.busy" @update:model-value="$event ? formOpen = true : closeForm()" @cancel="closeForm" @submit="save">
      <p v-if="formError" role="alert">{{ formError }}</p>
      <FormGrid :columns="2"><label>SKU<input v-model="form.sku" name="sku" autofocus required minlength="3" maxlength="32" pattern="[A-Za-z0-9][A-Za-z0-9_-]{2,31}"></label><label>名称<input v-model="form.name" name="name" required minlength="3" maxlength="120"></label><label class="wide">描述<textarea v-model="form.description" name="description" maxlength="500"></textarea></label><label>价格（分）<input v-model.number="form.priceCents" name="priceCents" type="number" min="0" max="100000000" required></label><label>状态<select v-model="form.status" name="status"><option value="active">启用</option><option value="inactive">停用</option></select></label></FormGrid>
    </FormDialog>
  </AppPage>
</template>

<style scoped>
.page-alert{margin:0}.table-scroll{min-width:0;overflow:auto}.row-actions{display:flex;gap:8px;white-space:nowrap}label{display:grid;gap:5px}.wide{grid-column:1/-1}@media(max-width:640px){.wide{grid-column:auto}}
</style>
