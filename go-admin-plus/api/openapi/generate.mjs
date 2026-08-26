import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const openapiRoot = dirname(fileURLToPath(import.meta.url))
const repositoryRoot = join(openapiRoot, '..', '..')
const sourcePath = join(repositoryRoot, 'docs', 'admin', 'admin_swagger.json')
const outputPath = join(openapiRoot, 'openapi.json')

const rewriteRefs = value => {
  if (Array.isArray(value)) return value.map(rewriteRefs)
  if (!value || typeof value !== 'object') return value
  return Object.fromEntries(Object.entries(value).map(([key, child]) => [
    key,
    key === '$ref' && typeof child === 'string'
      ? child.replace('#/definitions/', '#/components/schemas/')
      : rewriteRefs(child)
  ]))
}

const parameterSchema = parameter => {
  if (parameter.schema) return rewriteRefs(parameter.schema)
  const schema = {}
  for (const key of ['type', 'format', 'items', 'enum', 'default']) {
    if (parameter[key] !== undefined) schema[key] = rewriteRefs(parameter[key])
  }
  return Object.keys(schema).length ? schema : { type: 'string' }
}

const convertOperation = operation => {
  const converted = rewriteRefs(operation)
  delete converted.consumes
  delete converted.produces

  const parameters = []
  for (const parameter of converted.parameters ?? []) {
    if (parameter.in === 'body') {
      converted.requestBody = {
        required: Boolean(parameter.required),
        content: {
          'application/json': { schema: parameterSchema(parameter) }
        }
      }
      continue
    }
    const next = { ...parameter, schema: parameterSchema(parameter) }
    for (const key of ['type', 'format', 'items', 'enum', 'default']) delete next[key]
    parameters.push(next)
  }
  if (parameters.length) converted.parameters = parameters
  else delete converted.parameters

  converted.responses = Object.fromEntries(Object.entries(converted.responses ?? {}).map(([status, response]) => {
    const { schema, ...metadata } = response
    if (!schema) return [status, metadata]
    return [status, {
      ...metadata,
      content: {
        'application/json': { schema: rewriteRefs(schema) }
      }
    }]
  }))
  return converted
}

const sortDeep = value => {
  if (Array.isArray(value)) return value.map(sortDeep)
  if (!value || typeof value !== 'object') return value
  return Object.fromEntries(Object.keys(value).sort().map(key => [key, sortDeep(value[key])]))
}

const source = JSON.parse(readFileSync(sourcePath, 'utf8'))
const paths = {}
for (const [path, item] of Object.entries(source.paths ?? {})) {
  paths[path] = {}
  for (const [method, operation] of Object.entries(item)) {
    paths[path][method] = convertOperation(operation)
  }
}

const genericEnvelope = () => ({
  description: 'Standard go-admin response envelope',
  content: { 'application/json': { schema: { $ref: '#/components/schemas/response.Response' } } }
})
const moduleOperation = (method, operationId, path) => {
  const parameters = [...path.matchAll(/\{([^}]+)\}/g)].map(match => ({
    name: match[1],
    in: 'path',
    required: true,
    schema: { type: 'string' }
  }))
  const operation = {
    operationId,
    tags: ['ModularMonolith'],
    security: [{ Bearer: [] }],
    responses: { 200: genericEnvelope() }
  }
  if (parameters.length) operation.parameters = parameters
  if (method === 'post' || method === 'put' || method === 'patch') {
    operation.requestBody = {
      required: true,
      content: { 'application/json': { schema: { type: 'object', additionalProperties: true } } }
    }
  }
  return operation
}

const renameParameters = (operation, renames) => {
  for (const parameter of operation.parameters ?? []) {
    if (parameter.in === 'path' && renames[parameter.name]) parameter.name = renames[parameter.name]
  }
  return operation
}

const moveOperation = (from, method, to, options = {}) => {
  const operation = paths[from]?.[method]
  if (!operation) throw new Error(`cannot move missing ${method.toUpperCase()} ${from}`)
  delete paths[from][method]
  if (Object.keys(paths[from]).length === 0) delete paths[from]
  paths[to] ??= {}
  const next = renameParameters(operation, options.renameParameters ?? {})
  if (options.dropPathParameters) {
    next.parameters = (next.parameters ?? []).filter(parameter => parameter.in !== 'path')
    if (!next.parameters.length) delete next.parameters
  }
  paths[to][options.method ?? method] = next
}

const movePath = (from, to, renameParameters = {}) => {
  for (const method of Object.keys(paths[from] ?? {})) {
    moveOperation(from, method, to, { renameParameters })
  }
}

// The checked-in Swagger 2 file still supplies useful DTO schemas, but some
// paths predate the current modular router. Normalize those paths before the
// canonical OpenAPI artifact is generated. The Go contract test below compares
// this inventory in both directions with the routes registered by the modules.
movePath('/logout', '/api/v1/logout')
movePath('/api/v1/sys-config', '/api/v1/config')
movePath('/api/v1/sys-config/{id}', '/api/v1/config/{id}')
movePath('/api/v1/dept/{deptId}', '/api/v1/dept/{id}', { deptId: 'id' })
movePath('/api/v1/dict/type/{dictId}', '/api/v1/dict/type/{id}', { dictId: 'id' })
movePath('/api/v1/menuTreeselect/{roleId}', '/api/v1/roleMenuTreeselect/{roleId}')
moveOperation('/api/v1/post/{postId}', 'get', '/api/v1/post/{id}', {
  renameParameters: { postId: 'id' }
})
moveOperation('/api/v1/role-status/{id}', 'put', '/api/v1/role-status', { dropPathParameters: true })
moveOperation('/api/v1/sys-user/{userId}', 'get', '/api/v1/sys-user/{id}', {
  renameParameters: { userId: 'id' }
})
moveOperation('/api/v1/sys-user/{userId}', 'put', '/api/v1/sys-user', { dropPathParameters: true })
moveOperation('/api/v1/sys-user/{userId}', 'delete', '/api/v1/sys-user', { dropPathParameters: true })
renameParameters(paths['/api/v1/sys/tables/info/{tableId}'].get, { configKey: 'tableId' })
delete paths['/api/v1/sys-api'].delete

const moduleRoutes = [
  ['get', '/api/v1/demo-product', 'listDemoProducts'],
  ['get', '/api/v1/demo-product/{id}', 'getDemoProduct'],
  ['post', '/api/v1/demo-product', 'createDemoProduct'],
  ['put', '/api/v1/demo-product/{id}', 'updateDemoProduct'],
  ['delete', '/api/v1/demo-product', 'deleteDemoProducts'],
  ['get', '/api/v1/sysjob', 'listJobs'],
  ['get', '/api/v1/sysjob/{id}', 'getJob'],
  ['post', '/api/v1/sysjob', 'createJob'],
  ['put', '/api/v1/sysjob', 'updateJob'],
  ['delete', '/api/v1/sysjob', 'deleteJobs'],
  ['get', '/api/v1/job/start/{id}', 'startJob'],
  ['get', '/api/v1/job/remove/{id}', 'removeJob'],
  ['get', '/api/v1/metrics', 'getMetrics'],
  ['get', '/api/v1/health', 'getLegacyHealth'],
  ['get', '/api/v1/gen/preview/{tableId}', 'previewGeneratedCode'],
  ['get', '/api/v1/gen/toproject/{tableId}', 'generateCodeToProject'],
  ['get', '/api/v1/gen/apitofile/{tableId}', 'generateAPIFile'],
  ['get', '/api/v1/gen/todb/{tableId}', 'generateMenuAndAPI'],
  ['get', '/api/v1/gen/tabletree', 'getGeneratedTableTree'],
  ['get', '/api/v1/configKey/{configKey}', 'getConfigByKey'],
  ['get', '/api/v1/deptTree', 'getDepartmentTree'],
  ['get', '/api/v1/roleDeptTreeselect/{roleId}', 'getRoleDepartmentTree'],
  ['put', '/api/v1/roledatascope', 'updateRoleDataScope'],
  ['get', '/api/v1/sys/tables/info', 'getGeneratedTableInfo']
]
for (const [method, path, operationId] of moduleRoutes) {
  paths[path] ??= {}
  paths[path][method] = moduleOperation(method, operationId, path)
}

paths['/api/v1/metrics'].get = {
  operationId: 'getMetrics',
  tags: ['Operations'],
  responses: {
    200: {
      description: 'Prometheus metrics',
      content: { 'text/plain': { schema: { type: 'string' } } }
    }
  }
}
paths['/api/v1/health'].get = {
  operationId: 'getLegacyHealth',
  tags: ['Operations'],
  responses: { 200: { description: 'Legacy liveness response with an empty body' } }
}

const publicOperations = new Set([
  'GET /api/v1/app-config',
  'GET /api/v1/captcha',
  'GET /api/v1/health',
  'GET /api/v1/job/remove/{id}',
  'GET /api/v1/job/start/{id}',
  'GET /api/v1/metrics',
  'GET /api/v1/runtime/capabilities',
  'POST /api/v1/login'
])
for (const [path, item] of Object.entries(paths)) {
  if (!path.startsWith('/api/v1/')) continue
  for (const [method, operation] of Object.entries(item)) {
    const key = `${method.toUpperCase()} ${path}`
    if (publicOperations.has(key)) delete operation.security
    else operation.security = [{ Bearer: [] }]
  }
}

for (const [path, item] of Object.entries(paths)) {
  if (!path.startsWith('/api/v1/') || ['/api/v1/health', '/api/v1/metrics', '/api/v1/runtime/capabilities'].includes(path)) continue
  for (const operation of Object.values(item)) {
    const success = operation.responses?.['200']
    const schema = success?.content?.['application/json']?.schema
    if (!schema) continue
    success.content['application/json'].schema = {
      oneOf: [schema, { $ref: '#/components/schemas/ApiErrorEnvelope' }]
    }
  }
}

const operationResponse = schema => ({
  description: 'Successful operational response',
  content: { 'application/json': { schema: { $ref: `#/components/schemas/${schema}` } } }
})

paths['/health/live'] = {
  get: {
    operationId: 'getLiveness',
    tags: ['Operations'],
    responses: { 200: operationResponse('OperationalStatus') }
  }
}
paths['/health/ready'] = {
  get: {
    operationId: 'getReadiness',
    tags: ['Operations'],
    responses: {
      200: operationResponse('OperationalStatus'),
      503: {
        description: 'Application or a required dependency is not ready',
        content: { 'application/json': { schema: { $ref: '#/components/schemas/OperationalStatus' } } }
      }
    }
  }
}
paths['/api/v1/runtime/capabilities'] = {
  get: {
    operationId: 'getRuntimeCapabilities',
    tags: ['Operations'],
    responses: { 200: operationResponse('RuntimeCapabilities') }
  }
}

const schemas = rewriteRefs(source.definitions ?? {})
schemas.ApiErrorEnvelope = {
  type: 'object',
  required: ['code', 'msg'],
  properties: {
    code: { type: 'integer', format: 'int32' },
    data: { nullable: true },
    msg: { type: 'string' }
  }
}
schemas.OperationalStatus = {
  type: 'object',
  required: ['status'],
  properties: {
    status: { type: 'string', enum: ['live', 'ready', 'not_ready', 'method_not_allowed'] }
  }
}
schemas.RuntimeCapabilities = {
  type: 'object',
  required: ['hostProfile', 'version', 'desktop', 'offline', 'nativeDialogs'],
  properties: {
    hostProfile: { type: 'string' },
    version: { type: 'string' },
    desktop: { type: 'boolean' },
    offline: { type: 'boolean' },
    nativeDialogs: { type: 'boolean' }
  }
}

const document = sortDeep({
  openapi: '3.0.3',
  info: {
    title: source.info?.title ?? 'go-admin API',
    description: 'Authoritative Phase 1 HTTP wire contract for every go-admin host.',
    version: source.info?.version ?? '2.0.0'
  },
  servers: [{ url: '/', description: 'Same-origin Web or host-provided Desktop endpoint' }],
  paths,
  components: {
    securitySchemes: {
      Bearer: { type: 'http', scheme: 'bearer', bearerFormat: 'JWT' }
    },
    schemas
  }
})

const generated = `${JSON.stringify(document, null, 2)}\n`
if (process.argv.includes('--check')) {
  if (!existsSync(outputPath) || readFileSync(outputPath, 'utf8') !== generated) {
    console.error(`OpenAPI artifact is stale: run node ${fileURLToPath(import.meta.url)}`)
    process.exit(1)
  }
  console.log(`OpenAPI artifact is current: ${Object.keys(paths).length} paths`)
} else {
  writeFileSync(outputPath, generated)
  console.log(`Generated ${outputPath}: ${Object.keys(paths).length} paths`)
}
