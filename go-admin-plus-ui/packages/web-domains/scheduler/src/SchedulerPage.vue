<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import type { Definition, DefinitionInput, ExecutionStatus, ParameterField, TaskType } from '@go-admin/domain-scheduler'
import { settleSchedulerPageOperation, type SchedulerController } from './scheduler-controller'

const props = defineProps<{ controller: SchedulerController }>()
const emit = defineEmits<{ sessionRequired: [] }>()
const revision = ref(0)
const tab = ref<'definitions' | 'executions'>('definitions')
const search = ref('')
const executionFilters = reactive<{ definitionId: string; status: ExecutionStatus | '' }>({ definitionId: '', status: '' })
const edited = ref<Definition | null>(null)
const form = reactive({ name: '', taskType: '', minutes: '0', hours: '0', daysOfMonth: '', months: '1,2,3,4,5,6,7,8,9,10,11,12', weekdays: '', parameters: {} as Record<string, string | number | boolean> })
const definitions = computed(() => { void revision.value; return props.controller.definitions.snapshot() })
const executions = computed(() => { void revision.value; return props.controller.executions.snapshot() })
const taskTypes = computed(() => { void revision.value; return props.controller.taskTypes() })
const selectedTask = computed<TaskType | undefined>(() => taskTypes.value.find(value => value.key === form.taskType))
const blocked = computed(() => { void revision.value; return props.controller.busy || props.controller.hasPendingRepair() })
const can = (permission: string) => { void revision.value; return props.controller.can(permission) }
const failureMessage = computed(() => { void revision.value; const value = props.controller.failure(); if (value === 'relogin') return 'Your session must be renewed.'; if (value === 'forbidden') return 'You do not have permission for that action.'; if (value === 'validation') return 'Review the submitted values.'; if (value === 'not-found') return 'The task definition no longer exists.'; if (value === 'conflict') return 'The definition changed. Refresh before retrying.'; if (value === 'unavailable') return 'The scheduler service is unavailable.'; return '' })
const run = (operation: () => Promise<unknown>) => settleSchedulerPageOperation(operation, () => { if (props.controller.failure() === 'relogin') emit('sessionRequired'); revision.value += 1 })
const parseSet = (value: string) => value.trim() === '' ? [] : value.split(',').map(item => Number(item.trim()))
const request = (): DefinitionInput => ({ name: form.name.trim(), taskType: form.taskType, schedule: { minutes: parseSet(form.minutes), hours: parseSet(form.hours), daysOfMonth: parseSet(form.daysOfMonth), months: parseSet(form.months), weekdays: parseSet(form.weekdays) }, parameters: { ...form.parameters } })
const resetForm = () => { edited.value = null; form.name = ''; form.taskType = taskTypes.value[0]?.key ?? ''; form.minutes = '0'; form.hours = '0'; form.daysOfMonth = ''; form.months = '1,2,3,4,5,6,7,8,9,10,11,12'; form.weekdays = ''; form.parameters = {} }
const initializeParameters = (task: TaskType | undefined) => { const next: Record<string, string | number | boolean> = {}; for (const field of task?.fields ?? []) next[field.name] = field.kind === 'boolean' ? false : field.kind === 'integer' ? field.minimum ?? 0 : field.allowedValues?.[0] ?? ''; form.parameters = next }
watch(() => form.taskType, () => { if (!edited.value || edited.value.taskType !== form.taskType) initializeParameters(selectedTask.value) })
const edit = (value: Definition) => { edited.value = value; form.name = value.name; form.taskType = value.taskType; form.minutes = value.schedule.minutes.join(','); form.hours = value.schedule.hours.join(','); form.daysOfMonth = value.schedule.daysOfMonth.join(','); form.months = value.schedule.months.join(','); form.weekdays = value.schedule.weekdays.join(','); form.parameters = { ...value.parameters } }
const submit = () => run(async () => { const result = edited.value ? await props.controller.updateDefinition(edited.value, request()) : await props.controller.createDefinition(request()); if (result === 'completed') resetForm() })
const searchDefinitions = () => run(() => props.controller.definitions.search({ search: search.value }))
const searchExecutions = () => run(() => props.controller.executions.search({ ...executionFilters }))
const numberValue = (field: ParameterField) => ({ min: field.minimum, max: field.maximum })
onMounted(() => run(async () => { if (can('scheduler.definitions.read')) { await props.controller.refreshTaskTypes(); resetForm(); await props.controller.definitions.refresh() }; if (can('scheduler.executions.read')) await props.controller.executions.refresh(); if (!can('scheduler.definitions.read')) tab.value = 'executions' }))
</script>

<template>
  <main class="scheduler-page">
    <header><h1>Task schedules</h1></header>
    <p v-if="failureMessage" role="alert">{{ failureMessage }}</p>
    <button v-if="controller.hasPendingRepair()" type="button" data-testid="repair-scheduler" :disabled="controller.busy" @click="run(() => controller.repairProjection())">Refresh saved changes</button>
    <nav class="tabs" aria-label="Scheduler views">
      <button v-if="can('scheduler.definitions.read')" type="button" :aria-pressed="tab === 'definitions'" @click="tab = 'definitions'">Definitions</button>
      <button v-if="can('scheduler.executions.read')" type="button" :aria-pressed="tab === 'executions'" @click="tab = 'executions'">Executions</button>
    </nav>

    <section v-if="tab === 'definitions' && can('scheduler.definitions.read')">
      <form class="toolbar" @submit.prevent="searchDefinitions"><label>Search<input v-model.trim="search" maxlength="100"></label><button type="submit">Search</button><button type="button" @click="search = ''; run(() => controller.definitions.reset())">Reset</button></form>
      <table><thead><tr><th>Name</th><th>Task</th><th>Schedule</th><th>Status</th><th>Revision</th><th>Actions</th></tr></thead><tbody>
        <tr v-for="item in definitions.rows" :key="item.id" :data-row-key="item.id"><td>{{ item.name }}</td><td>{{ item.taskType }}</td><td>{{ item.schedule.hours.join(',') }}:{{ item.schedule.minutes.join(',') }} UTC</td><td>{{ item.enabled ? 'Enabled' : 'Stopped' }}</td><td>{{ item.revision }}</td><td><button v-if="can('scheduler.definitions.write')" type="button" data-action="edit" :disabled="item.enabled || blocked" @click="edit(item)">Edit</button><button v-if="can('scheduler.definitions.write')" type="button" data-action="toggle" :disabled="blocked" @click="run(() => item.enabled ? controller.stopDefinition(item) : controller.enableDefinition(item))">{{ item.enabled ? 'Stop' : 'Enable' }}</button><button v-if="can('scheduler.definitions.delete')" type="button" data-action="delete" :disabled="blocked" @click="run(() => controller.deleteDefinition(item))">Delete</button></td></tr>
      </tbody></table>
      <div class="pagination"><button type="button" :disabled="definitions.page <= 1" @click="run(() => controller.definitions.setPage(definitions.page - 1))">Previous</button><span>Page {{ definitions.page }}</span><button type="button" :disabled="definitions.page * definitions.pageSize >= definitions.total" @click="run(() => controller.definitions.setPage(definitions.page + 1))">Next</button></div>
      <form v-if="can('scheduler.definitions.write')" class="editor" data-testid="scheduler-definition-form" @submit.prevent="submit">
        <h2>{{ edited ? 'Edit definition' : 'Create definition' }}</h2>
        <label>Name<input name="name" v-model.trim="form.name" required maxlength="100"></label><label>Task type<select name="taskType" v-model="form.taskType" required :disabled="Boolean(edited)"><option disabled value="">Select</option><option v-for="task in taskTypes" :key="task.key" :value="task.key">{{ task.label }}</option></select></label>
        <fieldset><legend>UTC schedule</legend><label>Minutes<input name="minutes" v-model.trim="form.minutes" required pattern="[0-9, ]+"></label><label>Hours<input name="hours" v-model.trim="form.hours" required pattern="[0-9, ]+"></label><label>Days of month<input name="daysOfMonth" v-model.trim="form.daysOfMonth" pattern="[0-9, ]*"></label><label>Months<input name="months" v-model.trim="form.months" required pattern="[0-9, ]+"></label><label>Weekdays<input name="weekdays" v-model.trim="form.weekdays" pattern="[0-9, ]*"></label></fieldset>
        <fieldset v-if="selectedTask" data-testid="scheduler-parameters"><legend>Task parameters</legend><label v-for="field in selectedTask.fields" :key="field.name">{{ field.label }}<select v-if="field.allowedValues" :name="field.name" v-model="form.parameters[field.name]" :required="field.required"><option v-for="value in field.allowedValues" :key="value" :value="value">{{ value }}</option></select><input v-else-if="field.kind === 'integer'" :name="field.name" v-model.number="form.parameters[field.name]" type="number" :min="numberValue(field).min" :max="numberValue(field).max" :required="field.required"><input v-else-if="field.kind === 'boolean'" :name="field.name" v-model="form.parameters[field.name]" type="checkbox"><input v-else :name="field.name" v-model="form.parameters[field.name]" maxlength="256" :required="field.required"></label></fieldset>
        <div class="commands"><button type="submit" :disabled="blocked">{{ edited ? 'Save' : 'Create' }}</button><button v-if="edited" type="button" @click="resetForm">Cancel</button></div>
      </form>
    </section>

    <section v-else-if="tab === 'executions' && can('scheduler.executions.read')">
      <form class="toolbar" @submit.prevent="searchExecutions"><label>Definition ID<input v-model.trim="executionFilters.definitionId"></label><label>Status<select v-model="executionFilters.status"><option value="">All</option><option value="succeeded">Succeeded</option><option value="failed">Failed</option></select></label><button type="submit">Filter</button></form>
      <table><thead><tr><th>Started</th><th>Task</th><th>Status</th><th>Scheduled</th><th>Error</th><th>Executor</th></tr></thead><tbody><tr v-for="item in executions.rows" :key="item.id"><td>{{ item.startedAt }}</td><td>{{ item.taskType }}</td><td>{{ item.status }}</td><td>{{ item.scheduledFor }}</td><td>{{ item.errorCode ?? '' }}</td><td>{{ item.executorOwner }}</td></tr></tbody></table>
      <div class="pagination"><button type="button" :disabled="executions.page <= 1" @click="run(() => controller.executions.setPage(executions.page - 1))">Previous</button><span>Page {{ executions.page }}</span><button type="button" :disabled="executions.page * executions.pageSize >= executions.total" @click="run(() => controller.executions.setPage(executions.page + 1))">Next</button></div>
    </section>
    <section v-else><p>No scheduler view is available.</p></section>
  </main>
</template>

<style scoped>
.scheduler-page { display: grid; gap: 20px; max-width: 1180px; padding: 24px; } h1, h2 { margin: 0; letter-spacing: 0; } section, .editor, fieldset { display: grid; gap: 16px; } .tabs, .toolbar, .pagination, .commands { display: flex; gap: 8px; align-items: end; flex-wrap: wrap; } .tabs button[aria-pressed="true"] { border-bottom-color: #176b54; color: #176b54; } .editor { grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); border-top: 1px solid #d7dce2; padding-top: 16px; } .editor h2, .editor fieldset, .commands { grid-column: 1 / -1; } fieldset { grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); border: 1px solid #d7dce2; } label { display: grid; gap: 6px; } input, select, button { min-height: 40px; font: inherit; } input[type="checkbox"] { width: 22px; } button { border: 1px solid #aab2bc; background: #fff; padding: 6px 12px; } button:disabled { opacity: .5; } table { width: 100%; border-collapse: collapse; } th, td { border-bottom: 1px solid #d7dce2; padding: 10px; text-align: left; } th { color: #4c5968; font-size: 14px; } @media (max-width: 700px) { .scheduler-page { padding: 16px; overflow-x: auto; } table { min-width: 840px; } }
</style>
