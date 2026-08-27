import type { components } from './generated/client'

export type FileMetadata = components['schemas']['FileMetadata']
export type FilePage = components['schemas']['FilePage']
export type DeleteFileTarget = components['schemas']['DeleteFileTarget']
export type FilesFailure = 'relogin' | 'forbidden' | 'validation' | 'not-found' | 'conflict' | 'unavailable'
export type FilesPermissionCode = 'files.objects.read' | 'files.objects.write' | 'files.objects.delete'

export const filesPermissions = {
  read: 'files.objects.read',
  write: 'files.objects.write',
  delete: 'files.objects.delete',
} as const satisfies Readonly<Record<string, FilesPermissionCode>>

export interface FileQuery {
  readonly search: string
  readonly page: number
  readonly pageSize: number
  readonly sort: 'name' | 'sizeBytes' | 'createdAt'
  readonly direction: 'ascending' | 'descending'
}

export interface UploadCandidate {
  readonly name: string
  readonly type: string
  readonly size: number
  readonly body: Blob
}

export interface FilesClient {
  list(query: FileQuery): Promise<FilePage>
  upload(candidate: UploadCandidate): Promise<FileMetadata>
  download(id: string): Promise<Blob>
  delete(targets: ReadonlyArray<DeleteFileTarget>): Promise<void>
}

export class FilesRequestError extends Error {
  readonly category: FilesFailure
  constructor(category: FilesFailure) { super(category); this.category = category }
}

export const fileMediaTypes = ['application/pdf', 'image/jpeg', 'image/png', 'text/plain'] as const
export const maximumFileBytes = 10 * 1024 * 1024
export const codePointLength = (value: string): number => Array.from(value).length
export const validFileSearch = (value: string): boolean => codePointLength(value.trim()) <= 100
export const validFileName = (value: string): boolean => {
  const name = value.trim()
  return codePointLength(name) >= 1 && codePointLength(name) <= 255 && name !== '.' && name !== '..'
    && !name.includes('/') && !name.includes('\\') && !Array.from(name).some(character => {
      const codePoint = character.codePointAt(0) ?? 0
      return codePoint < 32 || codePoint === 127
    })
}
export const validUploadCandidate = (candidate: UploadCandidate): boolean => {
  return validFileName(candidate.name) && fileMediaTypes.includes(candidate.type as typeof fileMediaTypes[number])
    && Number.isSafeInteger(candidate.size) && candidate.size >= 0 && candidate.size <= maximumFileBytes
    && candidate.body.size === candidate.size && candidate.body.type === candidate.type
}
