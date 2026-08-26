import { readdirSync, realpathSync, statSync } from 'node:fs'
import { isAbsolute, join, relative, resolve, sep } from 'node:path'

const moduleIdPattern = /^[a-z][a-z0-9-]{1,63}$/
const goPackagePattern = /^[a-z][a-z0-9_]*$/

const assertRelativeOutput = (repositoryRoot, value, ownerRoot, label, source) => {
  if (typeof value !== 'string' || value.length === 0 || isAbsolute(value) || value.includes('\\')) {
    throw new Error(`${source}: ${label} output must be a repository-relative POSIX path`)
  }

  const repository = resolve(repositoryRoot)
  const output = resolve(repository, value)
  const owner = resolve(repository, ownerRoot)
  const ownerRelative = relative(owner, output)
  if (ownerRelative === '..' || ownerRelative.startsWith(`..${sep}`) || isAbsolute(ownerRelative)) {
    throw new Error(`${source}: ${label} output must stay inside ${ownerRoot}`)
  }
  return output
}

export const resolveModuleMetadata = (repositoryRoot, document, source = 'module contract') => {
  const id = document?.['x-go-admin-module']
  if (typeof id !== 'string' || !moduleIdPattern.test(id)) {
    throw new Error(`${source}: module id must match ${moduleIdPattern}`)
  }

  const codegen = document?.['x-go-admin-codegen']
  if (!codegen || typeof codegen !== 'object' || Array.isArray(codegen)) {
    throw new Error(`${source}: module codegen metadata is required`)
  }
  const codegenKeys = Object.keys(codegen).sort()
  const expectedCodegenKeys = ['goOutput', 'goPackage', 'typescriptOutput']
  if (JSON.stringify(codegenKeys) !== JSON.stringify(expectedCodegenKeys)) {
    throw new Error(`${source}: module codegen metadata must contain only ${expectedCodegenKeys.join(', ')}`)
  }
  if (typeof codegen.goPackage !== 'string' || !goPackagePattern.test(codegen.goPackage)) {
    throw new Error(`${source}: Go package must match ${goPackagePattern}`)
  }

  const goOutput = assertRelativeOutput(
    repositoryRoot,
    codegen.goOutput,
    join('go-admin-plus', 'internal', 'modules', id),
    'Go',
    source
  )
  if (!goOutput.endsWith(`${sep}transport${sep}openapi.gen.go`)) {
    throw new Error(`${source}: Go output must end with transport/openapi.gen.go`)
  }

  const typescriptOutput = assertRelativeOutput(
    repositoryRoot,
    codegen.typescriptOutput,
    join('go-admin-ui-plus', 'packages', 'domains', id),
    'TypeScript',
    source
  )
  if (!typescriptOutput.endsWith(`${sep}src${sep}generated`)) {
    throw new Error(`${source}: TypeScript output must end with src/generated`)
  }

  return { id, goPackage: codegen.goPackage, goOutput, typescriptOutput, source }
}

export const discoverModuleContracts = repositoryRoot => {
  const directory = join(repositoryRoot, 'contracts', 'openapi', 'modules')
  const realDirectory = realpathSync(directory)

  return readdirSync(directory, { withFileTypes: true })
    .filter(entry => !entry.name.startsWith('_') && /\.(?:json|ya?ml)$/i.test(entry.name))
    .map(entry => {
      if (!entry.isFile()) throw new Error(`${entry.name}: module contract must be a regular file`)
      const path = join(directory, entry.name)
      const realPath = realpathSync(path)
      const location = relative(realDirectory, realPath)
      if (location === '..' || location.startsWith(`..${sep}`) || isAbsolute(location)) {
        throw new Error(`${entry.name}: module contract resolves outside the modules directory`)
      }
      if (!statSync(realPath).isFile()) throw new Error(`${entry.name}: module contract must be a regular file`)
      return path
    })
    .sort((left, right) => left.localeCompare(right))
}
