import { readdir, readFile } from 'node:fs/promises'
import { extname, join, relative, resolve } from 'node:path'

const root = resolve(import.meta.dirname, '../../..')
const domainsRoot = join(root, 'domains')
const sourceExtensions = new Set(['.ts', '.vue', '.js'])
const errors = []

const walk = async directory => {
  const files = []
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) files.push(...await walk(path))
    else if (sourceExtensions.has(extname(entry.name))) files.push(path)
  }
  return files
}

for (const file of await walk(domainsRoot)) {
  const source = await readFile(file, 'utf8')
  const label = relative(root, file)
  if (source.includes("@/")) errors.push(`${label}: imports the Admin App Shell`)
  if (/from\s+['"][^'"]*(?:system|jobs|demo|tools|monitor)\/src\/(?:api|pages)\//.test(source)) {
    errors.push(`${label}: deep-imports another Domain`)
  }
}

for (const file of await walk(join(root, 'packages/ui/src'))) {
  const source = await readFile(file, 'utf8')
  const label = relative(root, file)
  if (/requestDomain|['"]\/api\//.test(source)) errors.push(`${label}: UI package accesses a business API`)
}

const retiredViewRoots = ['admin', 'schedule', 'demo', 'dev-tools', 'sys-tools']
for (const name of retiredViewRoots) {
  try {
    const files = await walk(join(root, 'apps/admin/src/views', name))
    if (files.length) errors.push(`apps/admin/src/views/${name}: still owns business pages`)
  } catch(error) {
    if (error?.code !== 'ENOENT') throw error
  }
}

if (errors.length) {
  for (const error of errors) console.error(`ERROR: ${error}`)
  process.exitCode = 1
} else {
  console.log('Domain boundaries: OK')
}
