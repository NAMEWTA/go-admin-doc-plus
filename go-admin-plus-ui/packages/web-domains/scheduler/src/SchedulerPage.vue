<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import type { Definition, DefinitionInput, ExecutionStatus, ParameterField, TaskType } from '@go-admin/domain-scheduler'
import { settleSchedulerPageOperation, type SchedulerController } from './scheduler-controller'

const props = defineProps<{ controller: SchedulerController }>()
const emit = defineEmits<{ sessionRequired: [] }>()
const revision = ref(0)
const tab = ref<'definitions' | 'executions'>('definitions')
const search = ref('')
const formOpen = ref(false)
const executionFilters = reactive<{ definitionId: string; status: ExecutionStatus | '' }>({ definitionId: '', status: '' })
const edited = ref<Definition | null>(null)
const form = reactive({ name: '', taskType: '', minutes: '0', hours: '0', daysOfMonth: '', months: '1,2,3,4,5,6,7,8,9,10,11,12', weekdays: '', parameters: {} as Record<string, string | number | boolean> })
const definitions = computed(() => { void revision.value; return props.controller.definitions.snapshot() })
const executions = computed(() => { void revision.value; return props.controller.executions.snapshot() })
const taskTypes = computed(() => { void revision.value; return props.controller.taskTypes() })
const selectedTask = computed<TaskType | undefined>(() => taskTypes.value.find(value => value.key === form.taskType))
const blocked = computed(() => { void revision.value; return props.controller.busy || props.controller.hasPendingRepair() })
const can = (permission: string) => { void revision.value; return props.controller.can(permission) }
const failureMessage = computed(() => { void revision.value; const value = props.controller.failure(); if (value === 'relogin') return '会话已失效，请重新登录。'; if (value === 'forbidden') return '没有执行该操作的权限。'; if (value === 'validation') return '请检查提交内容。'; if (value === 'not-found') return '任务定义已不存在。'; if (value === 'conflict') return '任务定义已发生变化，请刷新后重试。'; if (value === 'unavailable') return '调度服务暂不可用。'; return '' })
const run = (operation: () => Promise<unknown>) => settleSchedulerPageOperation(operation, () => { if (props.controller.failure() === 'relogin') emit('sessionRequired'); revision.value += 1 })
const parseSet = (value: string) => value.trim() === '' ? [] : value.split(',').map(item => Number(item.trim()))
const request = (): DefinitionInput => ({ name: form.name.trim(), taskType: form.taskType, schedule: { minutes: parseSet(form.minutes), hours: parseSet(form.hours), daysOfMonth: parseSet(form.daysOfMonth), months: parseSet(form.months), weekdays: parseSet(form.weekdays) }, parameters: { ...form.parameters } })
const initializeParameters = (task: TaskType | undefined) => { const next: Record<string, string | number | boolean> = {}; for (const field of task?.fields ?? []) next[field.name] = field.kind === 'boolean' ? false : field.kind === 'integer' ? field.minimum ?? 0 : field.allowedValues?.[0] ?? ''; form.parameters = next }
const resetForm = () => { const task = props.controller.taskTypes()[0]; edited.value = null; form.name = ''; form.taskType = task?.key ?? ''; form.minutes = '0'; form.hours = '0'; form.daysOfMonth = ''; form.months = '1,2,3,4,5,6,7,8,9,10,11,12'; form.weekdays = ''; initializeParameters(task) }
const closeForm = () => { formOpen.value = false; resetForm() }
const create = () => { resetForm(); formOpen.value = true }
watch(() => form.taskType, () => { if (!edited.value || edited.value.taskType !== form.taskType) initializeParameters(selectedTask.value) })
const edit = (value: Definition) => { edited.value = value; form.name = value.name; form.taskType = value.taskType; form.minutes = value.schedule.minutes.join(','); form.hours = value.schedule.hours.join(','); form.daysOfMonth = value.schedule.daysOfMonth.join(','); form.months = value.schedule.months.join(','); form.weekdays = value.schedule.weekdays.join(','); form.parameters = { ...value.parameters }; formOpen.value = true }
const submit = () => run(async () => { const result = edited.value ? await props.controller.updateDefinition(edited.value, request()) : await props.controller.createDefinition(request()); if (result === 'completed') closeForm() })
const searchDefinitions = () => run(() => props.controller.definitions.search({ search: search.value }))
const searchExecutions = () => run(() => props.controller.executions.search({ ...executionFilters }))
const numberValue = (field: ParameterField) => ({ min: field.minimum, max: field.maximum })
const executionStatusLabel = (value: ExecutionStatus) => value === 'succeeded' ? '成功' : '失败'
onMounted(() => run(async () => { if (can('scheduler.definitions.read')) { await props.controller.refreshTaskTypes(); resetForm(); await props.controller.definitions.refresh() }; if (can('scheduler.executions.read')) await props.controller.executions.refresh(); if (!can('scheduler.definitions.read')) tab.value = 'executions' }))
</script>

<template>
  <main class="scheduler-page">
    <header><h1>任务调度</h1></header>
    <p v-if="failureMessage" role="alert">{{ failureMessage }}</p>
    <button v-if="controller.hasPendingRepair()" type="button" data-testid="repair-scheduler" :disabled="controller.busy" @click="run(() => controller.repairProjection())">刷新已保存的变更</button>
    <nav class="tabs" aria-label="任务调度视图">
      <button v-if="can('scheduler.definitions.read')" type="button" :aria-pressed="tab === 'definitions'" @click="tab = 'definitions'">任务管理</button>
      <button v-if="can('scheduler.executions.read')" type="button" :aria-pressed="tab === 'executions'" @click="tab = 'executions'">执行记录</button>
    </nav>

    <section v-if="tab === 'definitions' && can('scheduler.definitions.read')">
      <form class="toolbar" @submit.prevent="searchDefinitions"><label>任务名称<input v-model.trim="search" maxlength="100" placeholder="请输入任务名称"></label><button type="submit">搜索</button><button type="button" @click="search = ''; run(() => controller.definitions.reset())">重置</button><button v-if="can('scheduler.definitions.write')" type="button" data-testid="open-scheduler-definition-form" @click="create">新增</button></form>
      <table><thead><tr><th>任务名称</th><th>任务类型</th><th>执行计划</th><th>状态</th><th>版本</th><th>操作</th></tr></thead><tbody>
        <tr v-for="item in definitions.rows" :key="item.id" :data-row-key="item.id"><td>{{ item.name }}</td><td>{{ item.taskType }}</td><td>{{ item.schedule.hours.join(',') }}:{{ item.schedule.minutes.join(',') }} UTC</td><td>{{ item.enabled ? '运行中' : '已停止' }}</td><td>{{ item.revision }}</td><td><button v-if="can('scheduler.definitions.write')" type="button" data-action="edit" :disabled="item.enabled || blocked" @click="edit(item)">修改</button><button v-if="can('scheduler.definitions.write')" type="button" data-action="toggle" :disabled="blocked" @click="run(() => item.enabled ? controller.stopDefinition(item) : controller.enableDefinition(item))">{{ item.enabled ? '停止' : '启用' }}</button><button v-if="can('scheduler.definitions.delete')" type="button" data-action="delete" :disabled="blocked" @click="run(() => controller.deleteDefinition(item))">删除</button></td></tr>
      </tbody></table>
      <div class="pagination"><button type="button" :disabled="definitions.page <= 1" @click="run(() => controller.definitions.setPage(definitions.page - 1))">上一页</button><span>第 {{ definitions.page }} 页</span><button type="button" :disabled="definitions.page * definitions.pageSize >= definitions.total" @click="run(() => controller.definitions.setPage(definitions.page + 1))">下一页</button></div>
      <div v-if="formOpen && can('scheduler.definitions.write')" class="management-dialog-backdrop" @click.self="closeForm" @keydown.esc="closeForm"><form class="management-dialog management-dialog--wide scheduler-definition-form" data-testid="scheduler-definition-form" role="dialog" aria-modal="true" aria-labelledby="scheduler-definition-title" @submit.prevent="submit"><header class="management-dialog__header"><h2 id="scheduler-definition-title">{{ edited ? '修改任务' : '新增任务' }}</h2><button type="button" aria-label="关闭" @click="closeForm">×</button></header><div class="management-dialog__body"><label>任务名称<input name="name" v-model.trim="form.name" autofocus required maxlength="100"></label><label>任务类型<select name="taskType" v-model="form.taskType" required :disabled="Boolean(edited)"><option disabled value="">请选择</option><option v-for="task in taskTypes" :key="task.key" :value="task.key">{{ task.label }}</option></select></label><fieldset><legend>UTC 执行计划</legend><label>分钟<input name="minutes" v-model.trim="form.minutes" required pattern="[0-9, ]+"></label><label>小时<input name="hours" v-model.trim="form.hours" required pattern="[0-9, ]+"></label><label>每月日期<input name="daysOfMonth" v-model.trim="form.daysOfMonth" pattern="[0-9, ]*"></label><label>月份<input name="months" v-model.trim="form.months" required pattern="[0-9, ]+"></label><label>星期<input name="weekdays" v-model.trim="form.weekdays" pattern="[0-9, ]*"></label></fieldset><fieldset v-if="selectedTask" data-testid="scheduler-parameters"><legend>任务参数</legend><label v-for="field in selectedTask.fields" :key="field.name">{{ field.label }}<select v-if="field.allowedValues" :name="field.name" v-model="form.parameters[field.name]" :required="field.required"><option v-for="value in field.allowedValues" :key="value" :value="value">{{ value }}</option></select><input v-else-if="field.kind === 'integer'" :name="field.name" v-model.number="form.parameters[field.name]" type="number" :min="numberValue(field).min" :max="numberValue(field).max" :required="field.required"><input v-else-if="field.kind === 'boolean'" :name="field.name" v-model="form.parameters[field.name]" type="checkbox"><input v-else :name="field.name" v-model="form.parameters[field.name]" maxlength="256" :required="field.required"></label></fieldset></div><footer class="management-dialog__footer"><button type="button" @click="closeForm">取消</button><button type="submit" :disabled="blocked">保存</button></footer></form></div>
    </section>

    <section v-else-if="tab === 'executions' && can('scheduler.executions.read')">
      <form class="toolbar" @submit.prevent="searchExecutions"><label>任务 ID<input v-model.trim="executionFilters.definitionId"></label><label>状态<select v-model="executionFilters.status"><option value="">全部</option><option value="succeeded">成功</option><option value="failed">失败</option></select></label><button type="submit">筛选</button></form>
      <table><thead><tr><th>开始时间</th><th>任务类型</th><th>状态</th><th>计划时间</th><th>错误</th><th>执行器</th></tr></thead><tbody><tr v-for="item in executions.rows" :key="item.id"><td>{{ item.startedAt }}</td><td>{{ item.taskType }}</td><td>{{ executionStatusLabel(item.status) }}</td><td>{{ item.scheduledFor }}</td><td>{{ item.errorCode ?? '' }}</td><td>{{ item.executorOwner }}</td></tr></tbody></table>
      <div class="pagination"><button type="button" :disabled="executions.page <= 1" @click="run(() => controller.executions.setPage(executions.page - 1))">上一页</button><span>第 {{ executions.page }} 页</span><button type="button" :disabled="executions.page * executions.pageSize >= executions.total" @click="run(() => controller.executions.setPage(executions.page + 1))">下一页</button></div>
    </section>
    <section v-else><p>当前没有可访问的任务调度视图。</p></section>
  </main>
</template>

<style scoped>
.scheduler-page { display: grid; gap: 20px; } h1, h2 { margin: 0; letter-spacing: 0; } section, fieldset { display: grid; gap: 16px; } .commands { display: flex; gap: 8px; align-items: end; flex-wrap: wrap; } .scheduler-definition-form fieldset { grid-column: 1 / -1; } fieldset { grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); padding: 12px; border: 1px solid var(--ga-border-light); border-radius: var(--ga-radius); } label { display: grid; gap: 6px; } input[type="checkbox"] { width: 22px; } @media (max-width: 700px) { .scheduler-page { overflow-x: auto; } table { min-width: 840px; } }
</style>
