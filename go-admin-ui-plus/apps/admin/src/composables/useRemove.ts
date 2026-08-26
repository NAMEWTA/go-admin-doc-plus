import { configureSuccessNotifier } from '../../../../packages/app-core/src'
import { msgSuccess } from '@/utils/message'

configureSuccessNotifier(msgSuccess)

export * from '../../../../packages/ui/src/composables/useRemove'
