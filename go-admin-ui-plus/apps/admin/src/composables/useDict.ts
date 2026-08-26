import { configureDictionaryLoader } from '../../../../packages/app-core/src'
import { getDicts } from '@/api/admin/dict/data'

configureDictionaryLoader(async type => (await getDicts(type)).data ?? [])

export * from '../../../../packages/ui/src/composables/useDict'
