#!/usr/bin/env node

import { existsSync, readFileSync, readdirSync } from 'node:fs'
import { basename, dirname, extname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const commandExtensions = new Set(['.cjs', '.js', '.json', '.md', '.mjs', '.ps1', '.sh', '.ts', '.yaml', '.yml'])
const ignoredCommandDirectories = new Set(['node_modules', 'dist', 'target'])

const commandFiles = directory => {
  if (!existsSync(directory)) return []
  return readdirSync(directory, { withFileTypes: true }).flatMap(entry => {
    if (ignoredCommandDirectories.has(entry.name)) return []
    const path = join(directory, entry.name)
    if (entry.isDirectory()) return commandFiles(path)
    return commandExtensions.has(extname(entry.name)) || /^(?:Containerfile|Taskfile)/.test(entry.name) ? [path] : []
  })
}

const workspacePackageNames = root => {
  const workspaceRoot = join(root, 'go-admin-plus-ui')
  const names = new Set()
  const rootManifest = join(workspaceRoot, 'package.json')
  if (existsSync(rootManifest)) names.add(JSON.parse(readFileSync(rootManifest, 'utf8')).name)
  for (const packageRoot of ['apps', 'packages', 'packages/adapters', 'packages/domains', 'packages/web-domains']) {
    const directory = join(workspaceRoot, packageRoot)
    if (!existsSync(directory)) continue
    for (const entry of readdirSync(directory, { withFileTypes: true })) {
      const manifest = join(directory, entry.name, 'package.json')
      if (entry.isDirectory() && existsSync(manifest)) names.add(JSON.parse(readFileSync(manifest, 'utf8')).name)
    }
  }
  return names
}

const workflowJob = (source, id) => {
  const lines = source.split('\n')
  const start = lines.findIndex(line => line === `  ${id}:`)
  if (start === -1) return ''
  const nextJob = lines.slice(start + 1).findIndex(line => /^  [A-Za-z0-9_-]+:\s*$/.test(line))
  const end = nextJob === -1 ? lines.length : start + nextJob + 1
  return lines.slice(start, end).join('\n')
}

export const checkArchitecture = root => {
  const failures = []
  const canonicalGoModule = 'github.com/NAMEWTA/go-admin-plus/go-admin-plus'
  const canonicalWorkspaceName = '@go-admin-plus/workspace'
  const canonicalVerification = {
    test: 'task test',
    typecheck: 'pnpm --dir go-admin-plus-ui typecheck',
    lint: 'task lint',
    build: 'task build TARGET=all PROFILE=server-sqlite'
  }
  const required = [
    'Taskfile.yml', '.github/workflows/ci.yml', 'go-admin-plus/go.mod',
    'go-admin-plus/cmd/go-admin-plus/main.go', 'go-admin-plus/cmd/desktop-sidecar/main.go',
    'go-admin-plus/cmd/config-check/main.go', 'go-admin-plus/cmd/migrate/main.go',
    'go-admin-plus/internal/app/product/registry.go', 'go-admin-plus/internal/modules',
    'go-admin-plus-ui/package.json', 'go-admin-plus-ui/pnpm-workspace.yaml',
    'go-admin-plus-ui/tests/shell/vitest.config.ts',
    'go-admin-plus-ui/tests/shell/node-tests.mjs',
    'go-admin-plus-ui/apps/admin-web/package.json',
    'go-admin-plus-ui/apps/admin-desktop/src-tauri/tauri.conf.json',
    'scripts/go-admin-plus-ui/build.sh',
    'scripts/go-admin-plus-ui/package.sh',
    'speculo/.speculo/specdev/config.json'
  ]
  const forbidden = ['go-admin-ui-plus', 'go-admin-plus/app', 'go-admin-plus/common', 'go-admin-plus/api', 'go-admin-plus/cmd/go-admin-desktop']
  for (const path of required) if (!existsSync(join(root, path))) failures.push(`missing canonical path: ${path}`)
  for (const path of forbidden) if (existsSync(join(root, path))) failures.push(`removed path still exists: ${path}`)

  const goModulePath = join(root, 'go-admin-plus/go.mod')
  if (existsSync(goModulePath)) {
    const declaration = readFileSync(goModulePath, 'utf8').match(/^module\s+(\S+)$/m)?.[1]
    if (declaration !== canonicalGoModule) failures.push(`Go module path must be ${canonicalGoModule}`)
  }

  const frontendManifestPath = join(root, 'go-admin-plus-ui/package.json')
  let frontendManifest
  if (existsSync(frontendManifestPath)) {
    frontendManifest = JSON.parse(readFileSync(frontendManifestPath, 'utf8'))
    if (frontendManifest.name !== canonicalWorkspaceName) failures.push(`frontend workspace name must be ${canonicalWorkspaceName}`)
    const firstTypecheckCommand = frontendManifest.scripts?.typecheck?.split('&&', 1)[0]?.trim()
    if (firstTypecheckCommand !== 'pnpm --recursive --if-present typecheck') {
      failures.push('frontend root typecheck must recursively run every workspace package typecheck script')
    }
    if (!frontendManifest.scripts?.test?.includes('node tests/shell/node-tests.mjs')) {
      failures.push('frontend root test must run Node unit test discovery')
    }
  }

  const specDevConfigPath = join(root, 'speculo/.speculo/specdev/config.json')
  if (existsSync(specDevConfigPath)) {
    const verification = JSON.parse(readFileSync(specDevConfigPath, 'utf8')).verification ?? {}
    for (const [name, command] of Object.entries(canonicalVerification)) {
      if (verification[name] !== command) failures.push(`SpecDev verification.${name} must be ${command}`)
    }
  }

  const packageNames = workspacePackageNames(root)
  const commandRoots = ['.github', 'scripts', 'release', 'deploy']
  const surfaces = [join(root, 'Taskfile.yml'), ...commandRoots.flatMap(path => commandFiles(join(root, path)))]
  for (const file of surfaces.filter(existsSync)) {
    const source = readFileSync(file, 'utf8')
    for (const match of source.matchAll(/@go-admin-plus\/[A-Za-z0-9._-]+/g)) {
      if (!packageNames.has(match[0])) failures.push(`${relative(root, file)} references unknown workspace package ${match[0]}`)
    }
  }

  const packageScriptPath = join(root, 'scripts/go-admin-plus-ui/package.sh')
  if (existsSync(packageScriptPath)) {
    const packageScript = readFileSync(packageScriptPath, 'utf8')
    const requiredPackageContracts = [
      [/GO_ADMIN_BUILD_DIR="\$web_dist" run_pnpm --filter @go-admin-plus\/admin-web build/, 'local package script must build Web into the artifacts staging directory'],
      [/tar -C "\$web_stage" -czf "\$package_tmp" dist/, 'local package script must archive the staged Web dist'],
      [/case \$\(go env GOHOSTOS\) in/, 'local package script must select Desktop bundles from GOHOSTOS'],
      [/darwin\) desktop_bundle=app/, 'local package script must support the macOS app bundle'],
      [/windows\) desktop_bundle=nsis/, 'local package script must support the Windows NSIS bundle'],
      [/tauri build \\\n\s+--features custom-protocol --bundles "\$desktop_bundle"/, 'local package script must enable the Tauri production protocol']
    ]
    for (const [pattern, message] of requiredPackageContracts) {
      if (!pattern.test(packageScript)) failures.push(message)
    }
    if (/pnpm build:prod/.test(packageScript)) failures.push('local package script must not invoke the aggregate frontend build')
  }

  const buildScriptPath = join(root, 'scripts/go-admin-plus-ui/build.sh')
  if (existsSync(buildScriptPath)) {
    const buildScript = readFileSync(buildScriptPath, 'utf8')
    if (!/node "\$repo_root\/release\/shared\/sidecar\/build\.mjs" --host/.test(buildScript)) {
      failures.push('Desktop build must stage the host Go sidecar')
    }
    if (!/run_pnpm --filter @go-admin-plus\/admin-desktop tauri build[\s\\]+--features custom-protocol --no-bundle/.test(buildScript)) {
      failures.push('Desktop build must compile the Tauri host without bundling')
    }
    if (!/all\)\s*\n\s*build_web\s*\n\s*build_desktop/.test(buildScript)) {
      failures.push('aggregate product build must include native Desktop')
    }
    if (!/desktop\)\s*\n\s*build_desktop/.test(buildScript)) {
      failures.push('Desktop target must use the native build')
    }
    if (/pnpm build:prod/.test(buildScript)) failures.push('product build must not stop at aggregate WebView assets')
  }

  const frontendTaskScriptRoot = join(root, 'scripts/go-admin-plus-ui')
  if (existsSync(frontendTaskScriptRoot)) {
    for (const entry of readdirSync(frontendTaskScriptRoot, { withFileTypes: true })) {
      if (!entry.isFile() || extname(entry.name) !== '.sh' || entry.name === 'common.sh') continue
      const path = join(frontendTaskScriptRoot, entry.name)
      if (/\b(?:exec\s+)?pnpm\s/.test(readFileSync(path, 'utf8'))) {
        failures.push('frontend task script must use managed pnpm resolution: ' + relative(root, path).replaceAll('\\', '/'))
      }
    }
  }

  const ciWorkflowPath = join(root, '.github/workflows/ci.yml')
  if (existsSync(ciWorkflowPath)) {
    const desktopJob = workflowJob(readFileSync(ciWorkflowPath, 'utf8'), 'desktop-rust')
    if (!/pnpm --dir go-admin-plus-ui install --frozen-lockfile/.test(desktopJob)) {
      failures.push('Desktop CI must install the frozen frontend workspace')
    }
    if (!/node release\/shared\/sidecar\/build\.mjs --host/.test(desktopJob)) {
      failures.push('Desktop CI must stage the host Go sidecar')
    }
    if (!/pnpm --dir go-admin-plus-ui --filter @go-admin-plus\/admin-desktop tauri build \\\n+\s+--features custom-protocol --no-bundle/.test(desktopJob)) {
      failures.push('Desktop CI must link the Tauri host without bundling')
    }
  }

  const internalRoot = join(root, 'go-admin-plus/internal')
  if (existsSync(internalRoot)) {
    const allowed = new Set(['app', 'application', 'contracts', 'host', 'modules', 'platform'])
    for (const entry of readdirSync(internalRoot, { withFileTypes: true })) {
      if (entry.isDirectory() && !allowed.has(entry.name)) failures.push(`backend layer is outside the canonical architecture: internal/${entry.name}`)
    }
  }
  const commandRoot = join(root, 'go-admin-plus/cmd')
  if (existsSync(commandRoot)) {
    const allowed = new Set(['config-check', 'desktop-sidecar', 'go-admin-plus', 'migrate'])
    for (const entry of readdirSync(commandRoot, { withFileTypes: true })) {
      if (entry.isDirectory() && !allowed.has(entry.name)) failures.push(`backend command is outside the canonical command plane: cmd/${entry.name}`)
    }
  }
  const workspacePath = join(root, 'go-admin-plus-ui/pnpm-workspace.yaml')
  if (existsSync(workspacePath)) {
    const workspace = readFileSync(workspacePath, 'utf8')
    for (const pattern of ['apps/*', 'packages/*', 'packages/adapters/*', 'packages/domains/*', 'packages/web-domains/*']) {
      if (!workspace.includes(pattern)) failures.push(`workspace does not declare ${pattern}`)
    }
  }
  const frontendTestConfigPath = join(root, 'go-admin-plus-ui/tests/shell/vitest.config.ts')
  if (existsSync(frontendTestConfigPath)) {
    const config = readFileSync(frontendTestConfigPath, 'utf8')
    if (!/['"]packages\/\*\*\/\*\.spec\.ts['"]/.test(config)) {
      failures.push('frontend test discovery must include every workspace package spec')
    }
    if (!/['"]tests\/e2e\/\*\*\/\*\.spec\.ts['"]/.test(config)) {
      failures.push('frontend test discovery must include E2E harness unit specs')
    }
  }
  if (frontendManifest) {
    const frontendRoot = join(root, 'go-admin-plus-ui')
    const typecheck = frontendManifest.scripts?.typecheck ?? ''
    const testProjects = commandFiles(join(frontendRoot, 'tests')).filter(path => basename(path) === 'tsconfig.json')
    for (const project of testProjects) {
      const projectPath = relative(frontendRoot, project).replaceAll('\\', '/')
      if (!typecheck.includes(projectPath)) failures.push(`frontend root typecheck omits test project ${projectPath}`)
    }
  }
  return failures
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const root = resolve(process.argv[2] ?? dirname(fileURLToPath(import.meta.url)), process.argv[2] ? '.' : '../..')
  const failures = checkArchitecture(root)
  if (failures.length) {
    console.error(`ARCHITECTURE_CHECK_FAIL\n${failures.join('\n')}`)
    process.exit(1)
  }
  console.log('ARCHITECTURE_CHECK_PASS')
}
