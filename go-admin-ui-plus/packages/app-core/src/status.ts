export const STATUS_NORMAL = '2'
export const statusToWire = (status: unknown): number => Number(status)
export const statusToForm = (status: unknown): string => String(status ?? STATUS_NORMAL)
