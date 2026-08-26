import { existsSync, readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const contractsRoot = dirname(fileURLToPath(import.meta.url))
const uiRoot = join(contractsRoot, '..', '..')

export const generatedContractsPath = join(contractsRoot, 'src', 'generated.ts')

export const resolveOpenAPIPath = () => {
  const backendRoot = process.env.GO_ADMIN_PATH ?? join(uiRoot, '..', 'go-admin-plus')
  return join(backendRoot, 'api', 'openapi', 'openapi.json')
}

const propertyName = name => JSON.stringify(name)
const schemaReference = reference => {
  const prefix = '#/components/schemas/'
  if (!reference.startsWith(prefix)) return 'unknown'
  return `components['schemas'][${JSON.stringify(reference.slice(prefix.length))}]`
}

const schemaType = schema => {
  if (!schema || typeof schema !== 'object') return 'unknown'
  let value
  if (schema.$ref) value = schemaReference(schema.$ref)
  else if (schema.const !== undefined) value = JSON.stringify(schema.const)
  else if (Array.isArray(schema.enum) && schema.enum.length) {
    value = schema.enum.map(item => JSON.stringify(item)).join(' | ')
  } else if (Array.isArray(schema.allOf)) {
    value = schema.allOf.map(schemaType).join(' & ') || 'unknown'
  } else if (Array.isArray(schema.oneOf)) {
    value = schema.oneOf.map(schemaType).join(' | ') || 'unknown'
  } else if (Array.isArray(schema.anyOf)) {
    value = schema.anyOf.map(schemaType).join(' | ') || 'unknown'
  } else if (schema.type === 'array') {
    value = `ReadonlyArray<${schemaType(schema.items)}>`
  } else if (schema.type === 'object' || schema.properties || schema.additionalProperties) {
    const required = new Set(schema.required ?? [])
    const fields = Object.entries(schema.properties ?? {}).map(([name, child]) =>
      `readonly ${propertyName(name)}${required.has(name) ? '' : '?'}: ${schemaType(child)}`)
    if (schema.additionalProperties) {
      fields.push(`readonly [key: string]: ${schema.additionalProperties === true ? 'unknown' : schemaType(schema.additionalProperties)}`)
    }
    value = fields.length ? `{ ${fields.join('; ')} }` : 'Readonly<Record<string, unknown>>'
  } else {
    value = ({ integer: 'number', number: 'number', boolean: 'boolean', string: 'string' })[schema.type] ?? 'unknown'
  }
  return schema.nullable ? `${value} | null` : value
}

const responseType = operation => {
  const responses = operation.responses ?? {}
  const response = responses['200'] ?? responses['201'] ?? responses.default
  return schemaType(response?.content?.['application/json']?.schema)
}

const requestType = operation => schemaType(operation.requestBody?.content?.['application/json']?.schema)

const parametersType = (operation, location) => {
  const parameters = (operation.parameters ?? []).filter(parameter => parameter.in === location)
  if (!parameters.length) return 'Readonly<Record<string, never>>'
  return `{ ${parameters.map(parameter =>
    `readonly ${propertyName(parameter.name)}${parameter.required ? '' : '?'}: ${schemaType(parameter.schema)}`
  ).join('; ')} }`
}

const operationName = (method, path, operation) => {
  if (operation.operationId) return operation.operationId
  const words = `${method} ${path.replaceAll(/[{}]/g, '')}`.split(/[^a-zA-Z0-9]+/).filter(Boolean)
  return words.map((word, index) => index === 0
    ? word.toLowerCase()
    : `${word[0].toUpperCase()}${word.slice(1)}`).join('')
}

export const generateContracts = document => {
  if (document.openapi !== '3.0.3' || !document.paths || !document.components?.schemas) {
    throw new Error('unsupported or incomplete OpenAPI document')
  }

  const operationEntries = []
  const pathEntries = []
  for (const path of Object.keys(document.paths).sort()) {
    const methods = []
    for (const method of Object.keys(document.paths[path]).sort()) {
      const operation = document.paths[path][method]
      const name = operationName(method, path, operation)
      operationEntries.push(
        `    readonly ${propertyName(name)}: { ` +
        `readonly method: ${JSON.stringify(method.toUpperCase())}; ` +
        `readonly path: ${JSON.stringify(path)}; ` +
        `readonly pathParameters: ${parametersType(operation, 'path')}; ` +
        `readonly query: ${parametersType(operation, 'query')}; ` +
        `readonly requestBody: ${requestType(operation)}; ` +
        `readonly response: ${responseType(operation)} }`
      )
      methods.push(`readonly ${propertyName(method)}: operations[${propertyName(name)}]`)
    }
    pathEntries.push(`    readonly ${propertyName(path)}: { ${methods.join('; ')} }`)
  }

  const schemaEntries = Object.keys(document.components.schemas).sort().map(name =>
    `    readonly ${propertyName(name)}: ${schemaType(document.components.schemas[name])}`)

  return [
    '// Code generated from go-admin-plus/api/openapi/openapi.json. DO NOT EDIT.',
    '',
    'export interface components {',
    '  readonly schemas: {',
    schemaEntries.join('\n'),
    '  }',
    '}',
    '',
    'export interface operations {',
    operationEntries.join('\n'),
    '}',
    '',
    'export interface paths {',
    pathEntries.join('\n'),
    '}',
    ''
  ].join('\n')
}

export const readGeneratedContracts = () => {
  const openapiPath = resolveOpenAPIPath()
  if (!existsSync(openapiPath)) return { openapiPath, generated: null, document: null }
  const document = JSON.parse(readFileSync(openapiPath, 'utf8'))
  return { openapiPath, generated: generateContracts(document), document }
}
