/** An error that the App Shell transport has already presented to the user. */
export interface ReportedError extends Error {
  reported?: true
}

export const asReportedError = (error: unknown): ReportedError =>
  error instanceof Error ? error : new Error(String(error))
