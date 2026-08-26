<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { AuditRequestError, type AuditFact, type AuditFailure, type AuditFilters } from '@go-admin/domain-audit'
import { consumeCleanupFailure, type AuditController } from './audit-controller'

const props = defineProps<{ controller: AuditController }>()
const emit = defineEmits<{ relogin: [] }>()
const version = ref(0)
const filters = reactive({ kind: '', action: '', outcome: '', source: '', from: '', to: '' })
const cleanupBefore = ref('')
const selected = ref<AuditFact | null>(null)
const failure = ref<AuditFailure | null>(null)
const busy = ref(false)
const cleanupStatus = ref<'completed' | 'refresh-failed' | 'repair-required' | null>(null)
const snapshot = computed(() => { void version.value; return props.controller.list.snapshot() })
const run = async (action: () => Promise<unknown>) => {
	if (busy.value) return
	busy.value = true
	failure.value = null
	try {
		await action()
	} catch (error) {
		failure.value = error instanceof AuditRequestError ? error.category : 'unavailable'
		if (failure.value === 'relogin') emit('relogin')
	} finally {
		busy.value = false
		version.value += 1
	}
}
const asFilters = (): AuditFilters => ({
	...(filters.kind ? { kind: filters.kind as AuditFilters['kind'] } : {}),
	...(filters.action ? { action: filters.action as AuditFilters['action'] } : {}),
	...(filters.outcome ? { outcome: filters.outcome as AuditFilters['outcome'] } : {}),
	...(filters.source ? { source: filters.source as AuditFilters['source'] } : {}),
	...(filters.from ? { from: new Date(filters.from).toISOString() } : {}),
	...(filters.to ? { to: new Date(filters.to).toISOString() } : {}),
})
const search = () => run(() => props.controller.list.search(asFilters()))
const reset = async () => {
	Object.assign(filters, { kind: '', action: '', outcome: '', source: '', from: '', to: '' })
	await run(() => props.controller.list.reset())
}
const detail = (id: string) => run(async () => { selected.value = await props.controller.detail(id) })
const consumeRefreshFailure = () => {
	failure.value = consumeCleanupFailure(props.controller, () => emit('relogin'))
}
const cleanup = () => run(async () => {
	cleanupStatus.value = null
	const result = await props.controller.cleanup(new Date(`${cleanupBefore.value}T00:00:00Z`).toISOString())
	if (result === 'completed' || result === 'refresh-failed' || result === 'repair-required') cleanupStatus.value = result
	if (result === 'failed') failure.value = consumeCleanupFailure(props.controller, () => emit('relogin'))
	if (result === 'refresh-failed') consumeRefreshFailure()
})
const repairCleanup = () => run(async () => {
	const result = await props.controller.repairCleanup()
	if (result === 'completed' || result === 'refresh-failed') cleanupStatus.value = result
	if (result === 'refresh-failed') consumeRefreshFailure()
})
const formatDate = (value: string) => new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'medium' }).format(new Date(value))
onMounted(() => { void run(() => props.controller.list.refresh()) })
</script>

<template>
  <main class="audit-page">
    <header><h1>Audit</h1></header>
    <form class="filters" aria-label="Audit filters" @submit.prevent="search">
			<label>Kind<select v-model="filters.kind"><option value="">All</option><option value="login">Login</option><option value="operation">Operation</option></select></label>
			<label>Action<select v-model="filters.action" data-testid="audit-action"><option value="">All</option><option value="login">Login</option><option value="create">Create</option><option value="update">Update</option><option value="delete">Delete</option></select></label>
			<label>Outcome<select v-model="filters.outcome"><option value="">All</option><option value="succeeded">Succeeded</option><option value="failed">Failed</option></select></label>
			<label>Source<select v-model="filters.source" data-testid="audit-source"><option value="">All</option><option value="web">Web</option><option value="desktop">Desktop</option><option value="server">Server</option></select></label>
		<label>From<input v-model="filters.from" type="datetime-local"></label>
		<label>To<input v-model="filters.to" type="datetime-local"></label>
			<div class="commands"><button type="submit" data-testid="audit-search" :disabled="busy">Search</button><button type="button" :disabled="busy" @click="reset">Reset</button></div>
    </form>
		<p v-if="failure === 'relogin'" role="alert">Your session has expired. Sign in again.</p>
		<p v-else-if="failure === 'forbidden'" role="alert">You do not have permission to use this audit operation.</p>
		<p v-else-if="failure" role="alert">The audit service is temporarily unavailable.</p>
    <div class="table-wrap">
      <table>
        <thead><tr><th>Time</th><th>Kind</th><th>Action</th><th>Outcome</th><th>Subject</th><th>Source</th><th></th></tr></thead>
        <tbody>
					<tr v-for="fact in snapshot.rows" :key="fact.id" data-testid="audit-row">
						<td>{{ formatDate(fact.occurredAt) }}</td><td>{{ fact.kind }}</td><td>{{ fact.action }}</td><td>{{ fact.outcome }}</td><td class="long-text">{{ fact.subject }}</td><td>{{ fact.source }}</td>
							<td><button type="button" data-testid="audit-view" @click="detail(fact.id)">View</button></td>
          </tr>
        </tbody>
      </table>
    </div>
		<footer class="paging"><span>{{ snapshot.total }} records</span><button type="button" :disabled="busy || snapshot.page <= 1" @click="run(() => controller.list.setPage(snapshot.page - 1))">Previous</button><span>Page {{ snapshot.page }}</span><button type="button" :disabled="busy || snapshot.page * snapshot.pageSize >= snapshot.total" @click="run(() => controller.list.setPage(snapshot.page + 1))">Next</button></footer>
    <section class="cleanup" aria-labelledby="cleanup-title">
      <h2 id="cleanup-title">Retention cleanup</h2>
			<label>Delete records before<input v-model="cleanupBefore" data-testid="audit-cleanup-before" type="date" required></label>
			<button type="button" data-testid="audit-cleanup" :disabled="busy || !cleanupBefore" @click="cleanup">Delete eligible records</button>
    </section>
		<p v-if="cleanupStatus === 'completed'" role="status" data-testid="audit-cleanup-status">Deleted {{ controller.lastCleanup()?.deleted ?? 0 }} records.<span v-if="controller.lastCleanup()?.moreEligible"> More eligible records remain.</span></p>
		<p v-else-if="cleanupStatus === 'refresh-failed' || cleanupStatus === 'repair-required'" role="status">Cleanup completed, but the list must be refreshed before another cleanup.<button type="button" data-testid="audit-cleanup-repair" :disabled="busy" @click="repairCleanup">Retry refresh</button></p>
		<dialog :open="selected !== null"><template v-if="selected"><h2>Audit detail</h2><dl><dt>Subject</dt><dd class="long-text">{{ selected.subject }}</dd><dt>Actor</dt><dd class="long-text">{{ selected.actorRef ?? selected.actorType }}</dd><dt>Outcome</dt><dd>{{ selected.outcome }}</dd><dt>Occurred</dt><dd>{{ formatDate(selected.occurredAt) }}</dd></dl><button type="button" @click="selected = null">Close</button></template></dialog>
  </main>
</template>

<style scoped>
.audit-page { display: grid; gap: 20px; max-width: 1200px; padding: 24px; }
h1, h2 { margin: 0; letter-spacing: 0; }
.filters { display: grid; grid-template-columns: repeat(3, minmax(140px, 1fr)); gap: 12px; align-items: end; }
label { display: grid; gap: 6px; }
select, input, button { min-height: 36px; font: inherit; }
.commands, .paging, .cleanup { display: flex; gap: 10px; align-items: end; }
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 10px 8px; border-bottom: 1px solid #d7dce2; text-align: left; white-space: nowrap; }
.long-text { max-width: 360px; overflow-wrap: anywhere; white-space: normal; }
th { color: #4a5561; font-size: 13px; }
.cleanup { border-top: 1px solid #d7dce2; padding-top: 20px; }
dialog { width: min(520px, calc(100% - 32px)); border: 1px solid #b7c0ca; border-radius: 6px; }
dl { display: grid; grid-template-columns: 100px 1fr; gap: 10px; }
@media (max-width: 760px) { .filters { grid-template-columns: 1fr 1fr; } .commands { grid-column: 1 / -1; } .cleanup { align-items: stretch; flex-direction: column; } }
</style>
