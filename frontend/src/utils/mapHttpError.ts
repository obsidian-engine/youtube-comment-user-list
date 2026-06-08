import { HttpError } from './api'

export type HttpErrorCode =
  | 'SERVER_UNREACHABLE'
  | 'SERVER_ERROR'
  | 'NETWORK'
  | 'TIMEOUT'
  | 'GENERIC'

export function mapHttpError(e: unknown): HttpErrorCode {
  if (e instanceof HttpError) {
    if (e.status === 404) return 'SERVER_UNREACHABLE'
    if (e.status >= 500) return 'SERVER_ERROR'
    return 'GENERIC'
  }
  if (e instanceof TypeError && e.message.includes('Failed to fetch')) return 'NETWORK'
  if (e instanceof Error && e.name === 'TimeoutError') return 'TIMEOUT'
  if (e instanceof Error && e.name === 'AbortError') throw e // re-throw (caller で無視)
  return 'GENERIC'
}
