import path from 'node:path'
import { fileURLToPath } from 'node:url'

export const configRoot = path.dirname(fileURLToPath(import.meta.url))
export const repoRoot = path.resolve(configRoot, '../..')
export const adminRoot = path.join(repoRoot, 'apps/admin')
export const adminSrc = path.join(adminRoot, 'src')
export const rootDist = path.join(repoRoot, 'dist')
