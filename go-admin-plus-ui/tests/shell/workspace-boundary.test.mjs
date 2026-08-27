import assert from 'node:assert/strict'
import { readFile, readdir } from 'node:fs/promises'
import { dirname, join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const workspaceRoot = join(dirname(fileURLToPath(import.meta.url)), '../..')
const readJson = async path => JSON.parse(await readFile(path, 'utf8'))
const browserAdapterWorkspaceDependencies = ['@go-admin/domain-files', '@go-admin/platform']
const assertBrowserAdapterDependencies = dependencies => {
  for (const dependency of browserAdapterWorkspaceDependencies) {
    assert.equal(dependencies[dependency], 'workspace:*')
  }
  assert.deepEqual(
    Object.keys(dependencies).filter(name => name.startsWith('@go-admin/')).sort(),
    browserAdapterWorkspaceDependencies,
    'browser adapter may only depend on its platform and Files domain ports'
  )
}
const requiredPackageNames = [
  '@go-admin/adapter-browser',
  '@go-admin/adapter-desktop',
  '@go-admin/admin-desktop',
  '@go-admin/admin-web',
  '@go-admin/app-shell',
  '@go-admin/domain-audit',
  '@go-admin/domain-demo',
  '@go-admin/domain-files',
  '@go-admin/domain-generator',
  '@go-admin/domain-iam',
  '@go-admin/domain-organization',
  '@go-admin/domain-scheduler',
  '@go-admin/domain-settings',
  '@go-admin/platform',
  '@go-admin/ui',
  '@go-admin/web-domain-audit',
  '@go-admin/web-domain-demo',
  '@go-admin/web-domain-files',
  '@go-admin/web-domain-generator',
  '@go-admin/web-domain-iam',
  '@go-admin/web-domain-organization',
  '@go-admin/web-domain-scheduler',
  '@go-admin/web-domain-settings'
]

const sourceFiles = async root => {
  const files = []
  for (const entry of await readdir(root, { withFileTypes: true }).catch(() => [])) {
    if (entry.name === 'node_modules' || entry.name === 'dist') continue
    const path = join(root, entry.name)
    if (entry.isDirectory()) files.push(...await sourceFiles(path))
    else if (/\.(?:[cm]?[jt]s|vue)$/.test(entry.name)) files.push(path)
  }
  return files
}

const packageDirectories = async () => {
  const roots = [
    ['apps'],
    ['packages'],
    ['packages', 'adapters'],
    ['packages', 'domains'],
    ['packages', 'web-domains']
  ]
  const directories = []
  for (const segments of roots) {
    const root = join(workspaceRoot, ...segments)
    for (const entry of await readdir(root, { withFileTypes: true }).catch(() => [])) {
      if (!entry.isDirectory()) continue
      const directory = join(root, entry.name)
      try {
        await readFile(join(directory, 'package.json'))
        directories.push(directory)
      } catch {}
    }
  }
  return directories
}

test('all planned packages are private and expose only public entry points', async () => {
  const directories = await packageDirectories()
  const manifests = await Promise.all(directories.map(async directory => ({
    directory,
    manifest: await readJson(join(directory, 'package.json'))
  })))
  const names = new Set(manifests.map(({ manifest }) => manifest.name))

  assert.equal(names.size, manifests.length, 'workspace package names must be unique')
  for (const name of requiredPackageNames) {
    assert.ok(names.has(name), `required workspace package ${name} is missing`)
  }
  for (const { directory, manifest } of manifests) {
    assert.equal(manifest.private, true, relative(workspaceRoot, directory))
    assert.equal(manifest.type, 'module', relative(workspaceRoot, directory))
    assert.ok(manifest.exports, `${relative(workspaceRoot, directory)} has no public exports`)
    for (const [dependency, version] of Object.entries({
      ...manifest.dependencies,
      ...manifest.devDependencies
    })) {
      if (dependency.startsWith('@go-admin/')) {
        assert.equal(version, 'workspace:*', `${manifest.name} must link ${dependency} through the workspace`)
      }
    }
  }
})

test('workspace dependencies are acyclic', async () => {
  const manifests = await Promise.all((await packageDirectories()).map(async directory =>
    readJson(join(directory, 'package.json'))
  ))
  const graph = new Map(manifests.map(manifest => [
    manifest.name,
    Object.keys(manifest.dependencies ?? {}).filter(name => name.startsWith('@go-admin/'))
  ]))
  for (const [name, dependencies] of graph) {
    for (const dependency of dependencies) {
      assert.ok(graph.has(dependency), `${name} references unknown workspace package ${dependency}`)
    }
  }
  const visited = new Set()
  const active = new Set()

  const visit = name => {
    if (active.has(name)) throw new Error(`workspace dependency cycle at ${name}`)
    if (visited.has(name)) return
    active.add(name)
    for (const dependency of graph.get(name) ?? []) visit(dependency)
    active.delete(name)
    visited.add(name)
  }
  for (const name of graph.keys()) visit(name)
})

test('workspace consumers use package exports rather than source paths', async () => {
  const rootManifest = await readJson(join(workspaceRoot, 'package.json'))
  for (const [dependency, version] of Object.entries({
    ...rootManifest.dependencies,
    ...rootManifest.devDependencies
  })) {
    if (dependency.startsWith('@go-admin/')) assert.equal(version, 'workspace:*')
  }

  for (const file of await sourceFiles(workspaceRoot)) {
    const source = await readFile(file, 'utf8')
    assert.doesNotMatch(
      source,
      /from\s+['"]@go-admin\/[^'"]+\/src(?:\/|['"])/,
      relative(workspaceRoot, file)
    )
  }

  const activeExports = [
    'packages/adapters/browser/src/index.ts',
    'packages/app-shell/src/core/index.ts',
    'packages/app-shell/src/product/index.ts',
    'packages/platform/src/index.ts',
    'packages/ui/src/index.ts',
    'apps/admin-web/src/main.ts'
  ]
  for (const target of activeExports) assert.ok(await readFile(join(workspaceRoot, target), 'utf8'))
})

test('apps select adapters without owning runtime transport', async () => {
  const adminWeb = await readJson(join(workspaceRoot, 'apps/admin-web/package.json'))
  assert.equal(adminWeb.dependencies['@go-admin/adapter-browser'], 'workspace:*')
  assert.equal(adminWeb.dependencies['@go-admin/platform'], undefined)
  assert.deepEqual(
    Object.keys(adminWeb.dependencies).filter(name => name.startsWith('@go-admin/')).sort(),
    ['@go-admin/adapter-browser', '@go-admin/app-shell']
  )

  const browserAdapter = await readJson(join(workspaceRoot, 'packages/adapters/browser/package.json'))
  assertBrowserAdapterDependencies(browserAdapter.dependencies)

  const adminDesktop = await readJson(join(workspaceRoot, 'apps/admin-desktop/package.json'))
  assert.equal(adminDesktop.dependencies['@go-admin/adapter-desktop'], 'workspace:*')
  assert.deepEqual(
    Object.keys(adminDesktop.dependencies).filter(name => name.startsWith('@go-admin/')).sort(),
    ['@go-admin/adapter-desktop', '@go-admin/app-shell']
  )

  const productShell = await readFile(join(workspaceRoot, 'packages/app-shell/src/product/ProductWorkspace.vue'), 'utf8')
  for (const module of ['iam', 'audit', 'organization', 'settings', 'generator', 'scheduler', 'demo', 'files']) {
    assert.match(productShell, new RegExp(`@go-admin/web-domain-${module}`), `product shell must compose ${module}`)
  }
  assert.match(productShell, /productRoutesFor\(props\.host\)/)

  const appSource = (await sourceFiles(join(workspaceRoot, 'apps/admin-web/src')))
  for (const file of appSource) {
    const source = await readFile(file, 'utf8')
    assert.doesNotMatch(source, /\bfetch\s*\(|implements\s+ShellRuntimePort/)
  }
})

test('browser adapter dependency allowlist rejects additional workspace packages', () => {
  assert.throws(
    () => assertBrowserAdapterDependencies({
      '@go-admin/domain-files': 'workspace:*',
      '@go-admin/platform': 'workspace:*',
      '@go-admin/app-shell': 'workspace:*'
    }),
    /browser adapter may only depend/
  )
})

test('headless packages do not depend on Vue, DOM globals, deep imports, or credential values', async () => {
  const headlessRoots = [
    join(workspaceRoot, 'packages/app-shell/src/core'),
    join(workspaceRoot, 'packages/platform/src')
  ]
  for (const root of headlessRoots) {
    for (const entry of await readdir(root, { withFileTypes: true })) {
      if (!entry.isFile()) continue
      const source = await readFile(join(root, entry.name), 'utf8')
      assert.doesNotMatch(source, /from\s+['"](?:vue|vue-router)['"]|\b(?:window|document|localStorage)\b/)
      assert.doesNotMatch(source, /@go-admin\/[^'"]+\/src\//)
    }
  }
  const platformSource = await readFile(join(workspaceRoot, 'packages/platform/src/index.ts'), 'utf8')
  assert.doesNotMatch(platformSource, /\b(?:secret|password|sessionToken|authorization)\b/i)
})

test('mobile shell keeps navigation compact and assigns remaining height to content', async () => {
  const styles = await readFile(join(workspaceRoot, 'apps/admin-web/src/styles.css'), 'utf8')
  const mobileBreakpoint = styles.match(/@media \(max-width: 640px\) \{([\s\S]*)\}\s*$/)?.[1]

  assert.ok(mobileBreakpoint, 'mobile shell breakpoint is missing')
  assert.match(
    mobileBreakpoint,
    /\.shell__workspace\s*\{[^}]*grid-template-columns:\s*1fr;[^}]*grid-template-rows:\s*auto 1fr;[^}]*\}/
  )
})
