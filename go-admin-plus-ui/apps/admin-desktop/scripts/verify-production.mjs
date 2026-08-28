#!/usr/bin/env node

import { lstat, readFile, readdir } from 'node:fs/promises'
import { dirname, resolve } from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

const forbiddenMarkers = [
  '/__desktop/test-control',
  'native-e2e',
  'E2E scope self',
  'E2E permissions off',
  'E2E revoke session'
]

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
  for (const path of files) {
    const content = await readFile(path)
    for (const marker of forbiddenMarkers) {
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
    await verifyDesktopProductionAssets(resolve(appRoot, 'dist'))
    process.stdout.write('DESKTOP_PRODUCTION_ASSETS_PASS\n')
  } catch (error) {
    process.stderr.write(`${error instanceof Error ? error.message : 'desktop production assets verification failed'}\n`)
    process.exitCode = 1
  }
}
