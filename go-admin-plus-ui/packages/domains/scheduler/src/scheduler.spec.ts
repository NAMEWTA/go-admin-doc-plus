import { describe, expect, it } from 'vitest'
import { validDefinitionInput, validSchedule } from './scheduler'

const taskTypes = [{ key: 'reports.daily', label: 'Daily report', fields: [
  { name: 'name', label: 'Name', kind: 'string' as const, required: true, allowedValues: ['sales'] },
  { name: 'limit', label: 'Limit', kind: 'integer' as const, required: true, minimum: 1, maximum: 10 },
] }]

describe('scheduler model validation', () => {
  it('accepts structured UTC schedules and rejects ambiguous day selectors', () => {
    expect(validSchedule({ minutes: [0], hours: [1], daysOfMonth: [], months: [1], weekdays: [1] })).toBe(true)
    expect(validSchedule({ minutes: [0], hours: [1], daysOfMonth: [1], months: [1], weekdays: [1] })).toBe(false)
    expect(validSchedule({ minutes: [0, 0], hours: [1], daysOfMonth: [], months: [1], weekdays: [] })).toBe(false)
  })

  it('validates the selected typed registry descriptor without extra parameters', () => {
    const base = { name: 'Sales', taskType: 'reports.daily', schedule: { minutes: [0], hours: [1], daysOfMonth: [], months: [1], weekdays: [] }, parameters: { name: 'sales', limit: 5 } }
    expect(validDefinitionInput(base, taskTypes)).toBe(true)
    expect(validDefinitionInput({ ...base, parameters: { ...base.parameters, secret: 'x' } }, taskTypes)).toBe(false)
    expect(validDefinitionInput({ ...base, parameters: { name: 'sales', limit: 11 } }, taskTypes)).toBe(false)
    expect(validDefinitionInput({ ...base, parameters: { name: 'sales', limit: Number.MAX_SAFE_INTEGER + 1 } }, taskTypes)).toBe(false)
  })
})
