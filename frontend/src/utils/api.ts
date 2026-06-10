export const BASE = import.meta.env.VITE_BACKEND_URL || ''

export class HttpError extends Error {
  constructor(
    public readonly status: number,
    message?: string,
  ) {
    super(message ?? `HTTP ${status}`)
    this.name = 'HttpError'
  }
}

export type LogDetail = {
  level: string
  source: string
  message: string
  timestamp?: string
}

export type ErrorResponse = {
  error: string
  code?: string
  message?: string
  httpCode: number
  logs?: LogDetail[]
}

export class BackendError extends Error {
  code?: string
  httpCode: number
  logs: LogDetail[]

  constructor(msg: string, opts: { code?: string; httpCode: number; logs?: LogDetail[] }) {
    super(msg)
    this.name = 'BackendError'
    this.code = opts.code
    this.httpCode = opts.httpCode
    this.logs = opts.logs ?? []
  }
}

async function parseErrorResponse(res: Response): Promise<never> {
  let errResp: ErrorResponse | null = null
  try {
    errResp = (await res.json()) as ErrorResponse
  } catch {
    // JSON parse 失敗時は fallback
  }
  if (errResp) {
    throw new BackendError(errResp.message ?? errResp.error ?? `HTTP ${res.status}`, {
      code: errResp.code,
      httpCode: errResp.httpCode ?? res.status,
      logs: errResp.logs,
    })
  }
  throw new HttpError(res.status)
}

async function json<T>(res: Response): Promise<T> {
  if (!res.ok) return parseErrorResponse(res)
  return res.json() as Promise<T>
}

async function throwIfError(res: Response): Promise<void> {
  if (!res.ok) return parseErrorResponse(res)
}

export type StatusResponse = {
  status?: 'WAITING' | 'ACTIVE'
  count?: number
  videoId?: string
  startedAt?: string
  endedAt?: string
  snapshotSavedAt?: string
}

export async function getStatus(signal?: AbortSignal): Promise<StatusResponse> {
  const res = await fetch(`${BASE}/status`, { signal })
  return json<StatusResponse>(res)
}

export type User = {
  channelId: string
  displayName: string
  joinedAt: string
  firstCommentedAt?: string
  commentCount?: number
  latestCommentedAt?: string
}

export async function getUsers(signal?: AbortSignal): Promise<User[] | null> {
  const res = await fetch(`${BASE}/users.json`, { signal })
  return json<User[] | null>(res)
}

export async function postSwitchVideo(videoId: string, signal?: AbortSignal): Promise<void> {
  const res = await fetch(`${BASE}/switch-video`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ videoId }),
    signal,
  })
  await throwIfError(res)
}

export type BackendLog = {
  level: 'info' | 'warn' | 'error'
  source: string
  message: string
}

export type PullResponse = {
  addedCount: number
  skippedCount: number
  autoReset: boolean
  pollingIntervalMillis: number
  logs: BackendLog[]
}

export async function postPull(signal?: AbortSignal): Promise<PullResponse> {
  const res = await fetch(`${BASE}/pull`, { method: 'POST', signal })
  return json<PullResponse>(res)
}

export async function postReset(signal?: AbortSignal): Promise<void> {
  const res = await fetch(`${BASE}/reset`, { method: 'POST', signal })
  await throwIfError(res)
}

export type Comment = {
  id: string
  channelId: string
  displayName: string
  message: string
  publishedAt: string
}

const MAX_RETRIES = 3
const INITIAL_DELAY = 1000

async function fetchWithRetry<T>(
  url: string,
  options: RequestInit = {},
  retries = MAX_RETRIES,
): Promise<T> {
  let lastError: Error | null = null

  for (let i = 0; i < retries; i++) {
    try {
      const res = await fetch(url, options)
      if (!res.ok) {
        return parseErrorResponse(res)
      }
      return (await res.json()) as T
    } catch (e) {
      lastError = e instanceof Error ? e : new Error(String(e))

      // AbortErrorはリトライしない
      if (lastError.name === 'AbortError') {
        throw lastError
      }

      // 最後のリトライでなければ待機
      if (i < retries - 1) {
        const delay = INITIAL_DELAY * Math.pow(2, i)
        await new Promise((resolve) => setTimeout(resolve, delay))
      }
    }
  }

  throw lastError
}

export async function searchComments(
  keywords: string[],
  signal?: AbortSignal,
): Promise<Comment[] | null> {
  const params = new URLSearchParams({ keywords: keywords.join(',') })
  return fetchWithRetry<Comment[] | null>(`${BASE}/comments?${params}`, { signal })
}

export interface HistorySummary {
  videoId: string
  savedAt: string
  userCount: number
  commentCount: number
}

export interface HistorySnapshot {
  videoId: string
  liveChatId?: string
  savedAt: string
  users: User[]
  comments: Comment[]
  processedMsgs?: string[]
  state?: { status?: string; startedAt?: string; endedAt?: string }
}

export async function getHistorySnapshots(signal?: AbortSignal): Promise<HistorySummary[]> {
  const res = await fetch(`${BASE}/history/snapshots`, { signal })
  if (!res.ok) return parseErrorResponse(res)
  const data = (await res.json()) as { items?: HistorySummary[] }
  return data.items ?? []
}

export async function getHistorySnapshot(
  videoId: string,
  signal?: AbortSignal,
): Promise<HistorySnapshot> {
  const res = await fetch(`${BASE}/history/snapshots/${encodeURIComponent(videoId)}`, { signal })
  if (!res.ok) return parseErrorResponse(res)
  return res.json() as Promise<HistorySnapshot>
}
