#!/usr/bin/env node

import { dirname, join, resolve } from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

import { hostTriple, outputName } from '../../../../release/shared/sidecar/build.mjs'
import { verifyDesktopProductionAssets, verifyDesktopProductionFiles } from './verify-production.mjs'

export const desktopProductionArtifactPaths = (repository, platform = process.platform, architecture = process.arch) => {
  const root = resolve(repository)
  const triple = hostTriple(platform, architecture)
  const appRoot = join(root, 'go-admin-plus-ui/apps/admin-desktop')
  return {
    webview: join(appRoot, 'dist'),
    sidecar: join(appRoot, 'src-tauri/binaries', outputName(triple)),
    host: join(appRoot, 'src-tauri/target/release', platform === 'win32' ? 'go-admin-plus-desktop.exe' : 'go-admin-plus-desktop')
  }
}

export const verifyDesktopProductionBuild = async (repository, platform = process.platform, architecture = process.arch) => {
  const artifacts = desktopProductionArtifactPaths(repository, platform, architecture)
  await verifyDesktopProductionAssets(artifacts.webview)
  await verifyDesktopProductionFiles([artifacts.sidecar, artifacts.host])
}

const invoked = process.argv[1] && pathToFileURL(resolve(process.argv[1])).href === import.meta.url
if (invoked) {
  const repository = resolve(dirname(fileURLToPath(import.meta.url)), '../../../..')
  try {
    await verifyDesktopProductionBuild(repository)
    process.stdout.write('DESKTOP_PRODUCTION_BUILD_PASS\n')
  } catch (error) {
    process.stderr.write(`${error instanceof Error ? error.message : 'desktop production build verification failed'}\n`)
    process.exitCode = 1
  }
}
