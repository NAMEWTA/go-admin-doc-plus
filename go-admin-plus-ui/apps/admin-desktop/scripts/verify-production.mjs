#!/usr/bin/env node

import { lstat, readFile, readdir } from 'node:fs/promises'
import { dirname, resolve } from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

export const desktopNativeControlMarkers = Object.freeze([
  '/__desktop/test-control',
  'native-e2e',
  'VITE_GO_ADMIN_NATIVE_E2E',
  'GO_ADMIN_DESKTOP_NATIVE_E2E',
  'GO_ADMIN_DESKTOP_E2E_',
  'desktop_native_e2e',
  'E2E authenticated boundary verified',
  'E2E unauthenticated boundary verified',
  'E2E boundary blocked:',
  'E2E self scope enforced',
  'E2E all scope restored',
  'E2E authorization denied',
  'E2E control failed:',
  'E2E scope self',
  'E2E scope all',
  'E2E permissions off',
  'E2E permissions on',
  'E2E revoke session',
  'E2E reset theme',
  'E2E theme storage cleared',
  'E2E-FOREIGN',
  'E2E-001',
  'native E2E credential identity'
])

export const desktopProductionPermissions = Object.freeze([
  'allow-desktop-request',
  'allow-desktop-identity',
  'allow-desktop-first-setup-state',
  'allow-desktop-first-setup-submit',
  'allow-desktop-navigation',
  'allow-desktop-login',
  'allow-desktop-logout',
  'allow-desktop-session-heartbeat',
  'allow-desktop-session-renew',
  'allow-desktop-pick-file',
  'allow-desktop-save-file',
  'allow-desktop-notify',
  'allow-desktop-write-clipboard'
])

const exactArray = (actual, expected) => Array.isArray(actual) && actual.length === expected.length && actual.every((value, index) => value === expected[index])

export const validateDesktopProductionConfiguration = (config, capability) => {
  const windows = config?.app?.windows
  const mainWindow = Array.isArray(windows) && windows.length === 1 ? windows[0] : undefined
  if (
    config?.identifier !== 'com.goadmin.plus' ||
    config?.build?.frontendDist !== '../dist' ||
    !exactArray(config?.app?.security?.capabilities, ['main-window']) ||
    mainWindow?.label !== 'main' || mainWindow?.visible !== false ||
    mainWindow?.minWidth !== 960 || mainWindow?.minHeight !== 640 ||
    !exactArray(config?.bundle?.externalBin, ['binaries/go-admin-sidecar']) ||
    capability?.identifier !== 'main-window' ||
    !exactArray(capability?.windows, ['main']) ||
    !exactArray(capability?.permissions, desktopProductionPermissions)
  ) {
    throw new Error('desktop production capability configuration is invalid')
  }
}

const readJson = async (path, name) => {
  try {
    return JSON.parse(await readFile(path, 'utf8'))
  } catch {
    throw new Error(`desktop production ${name} is invalid`)
  }
}

export const verifyDesktopProductionConfiguration = async appRoot => {
  const root = resolve(appRoot)
  const [config, capability] = await Promise.all([
    readJson(resolve(root, 'src-tauri/tauri.conf.json'), 'Tauri configuration'),
    readJson(resolve(root, 'src-tauri/capabilities/main.json'), 'capability manifest')
  ])
  validateDesktopProductionConfiguration(config, capability)
}

const filesBelow = async directory => {
  const files = []
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const path = resolve(directory, entry.name)
    if (entry.isSymbolicLink()) throw new Error('desktop production assets contain a symbolic link')
    if (entry.isDirectory()) files.push(...await filesBelow(path))
    else if (entry.isFile()) files.push(path)
    else throw new Error('desktop production assets contain an unsupported entry')
  }
  return files
}

export const verifyDesktopProductionAssets = async directory => {
  const root = resolve(directory)
  const info = await lstat(root)
  if (!info.isDirectory() || info.isSymbolicLink()) throw new Error('desktop production assets directory is invalid')
  const files = await filesBelow(root)
  if (files.length === 0) throw new Error('desktop production assets are empty')
  await verifyDesktopProductionFiles(files)
}

export const verifyDesktopProductionFiles = async paths => {
  if (!Array.isArray(paths) || paths.length === 0) throw new Error('desktop production file list is empty')
  for (const unresolvedPath of paths) {
    const path = resolve(unresolvedPath)
    const info = await lstat(path)
    if (!info.isFile() || info.isSymbolicLink()) throw new Error('desktop production artifact is invalid')
    const content = await readFile(path)
    for (const marker of desktopNativeControlMarkers) {
      if (content.includes(Buffer.from(marker))) {
        throw new Error(`desktop production assets contain native test control: ${marker}`)
      }
    }
  }
}

const invoked = process.argv[1] && pathToFileURL(resolve(process.argv[1])).href === import.meta.url
if (invoked) {
  const appRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..')
  try {
    const args = process.argv.slice(2)
    if (args.length === 0) {
      await verifyDesktopProductionConfiguration(appRoot)
      await verifyDesktopProductionAssets(resolve(appRoot, 'dist'))
    }
    else if (args[0] === '--files' && args.length > 1) await verifyDesktopProductionFiles(args.slice(1))
    else throw new Error('desktop production verification arguments are invalid')
    process.stdout.write('DESKTOP_PRODUCTION_ASSETS_PASS\n')
  } catch (error) {
    process.stderr.write(`${error instanceof Error ? error.message : 'desktop production assets verification failed'}\n`)
    process.exitCode = 1
  }
}
