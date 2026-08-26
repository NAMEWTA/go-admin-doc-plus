const methods = new Set(['get', 'put', 'post', 'delete', 'options', 'head', 'patch', 'trace'])
const errorCategories = [
  'validation',
  'authentication',
  'authorization',
  'not_found',
  'conflict',
  'internal'
]
const sensitiveDetail = /(?:\bselect\b.{0,100}\bfrom\b|\binsert\s+into\b|\bdelete\s+from\b|\bupdate\s+\w+\s+set\b|sqlstate|\bpq:|\bgorm\b|stack\s*trace|\/(?:etc|home|Users|var)\/|[A-Za-z]:\\|\b(?:secret|session|password)\s*[=:])/i

const visitStrings = (value, path, visit) => {
  if (typeof value === 'string') {
    visit(value, path)
    return
  }
  if (Array.isArray(value)) {
    value.forEach((item, index) => visitStrings(item, `${path}/${index}`, visit))
    return
  }
  if (value && typeof value === 'object') {
    for (const [key, item] of Object.entries(value)) visitStrings(item, `${path}/${key}`, visit)
  }
}

export const validatePolicy = (
  document,
  { canonical = false, operationIds = new Map(), source = 'contract' } = {}
) => {
  const problems = []
  if (typeof document?.openapi !== 'string' || !document.openapi.startsWith('3.1.')) {
    problems.push('/openapi must declare OpenAPI 3.1')
  }

  visitStrings(document, '#', (value, path) => {
    if (sensitiveDetail.test(value)) problems.push(`${path} exposes sensitive internal detail`)
  })

  for (const [path, pathItem] of Object.entries(document?.paths ?? {})) {
    for (const [method, operation] of Object.entries(pathItem ?? {})) {
      if (!methods.has(method) || !operation || typeof operation !== 'object') continue
      if (typeof operation.operationId === 'string') {
        const first = operationIds.get(operation.operationId)
        const current = `${source}:${path}/${method}`
        if (first) problems.push(`${current}/operationId duplicates ${first}`)
        else operationIds.set(operation.operationId, current)
      }
      for (const [status, response] of Object.entries(operation.responses ?? {})) {
        if (!/^[45](?:\d\d|XX)$/.test(status)) continue
        if (!response?.content?.['application/problem+json']) {
          problems.push(`${path}/${method}/responses/${status} must use application/problem+json`)
        }
      }
    }
  }

  if (canonical) {
    const actual = document?.components?.schemas?.Problem?.properties?.category?.enum
    if (JSON.stringify(actual) !== JSON.stringify(errorCategories)) {
      problems.push('/components/schemas/Problem category enum differs from the stable error categories')
    }
  }

  if (problems.length) throw new Error(`contract policy failed:\n- ${problems.join('\n- ')}`)
}
