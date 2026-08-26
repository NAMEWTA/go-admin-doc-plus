export function parseTime(time?: string | number | Date | null, pattern = '{y}-{m}-{d} {h}:{i}:{s}'): string | null {
  if (!time) return null
  if (typeof time === 'string' && time.includes('01-01-01')) return '-'
  const normalized = typeof time === 'string' && /^[0-9]+$/.test(time) ? Number(time) : time
  const timestamp = typeof normalized === 'number' && String(normalized).length === 10 ? normalized * 1000 : normalized
  const date = timestamp instanceof Date ? timestamp : new Date(timestamp)
  const values: Record<string, number> = {
    y: date.getFullYear(), m: date.getMonth() + 1, d: date.getDate(),
    h: date.getHours(), i: date.getMinutes(), s: date.getSeconds(), a: date.getDay()
  }
  return pattern.replace(/{(y|m|d|h|i|s|a)+}/g, (result, key: string) => {
    const value = values[key]
    if (key === 'a') return ['日', '一', '二', '三', '四', '五', '六'][value]
    return result.length > 0 && value < 10 ? `0${value}` : String(value || 0)
  })
}
