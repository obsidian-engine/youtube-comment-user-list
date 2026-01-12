export const BASE = import.meta.env.VITE_BACKEND_URL || ''

async function json<T>(res: Response): Promise<T> {
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json() as Promise<T>
}

export type StatusResponse = {
  status?: 'WAITING' | 'ACTIVE'
  Status?: 'WAITING' | 'ACTIVE'
  count?: number
  videoId?: string
  startedAt?: string
  endedAt?: string
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

export async function getUsers(signal?: AbortSignal): Promise<User[]> {
  const res = await fetch(`${BASE}/users.json`, { signal })
  return json<User[]>(res)
}

export async function postSwitchVideo(videoId: string, signal?: AbortSignal): Promise<void> {
  const res = await fetch(`${BASE}/switch-video`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ videoId }),
    signal,
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
}

export async function postPull(signal?: AbortSignal): Promise<void> {
  const res = await fetch(`${BASE}/pull`, { method: 'POST', signal })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
}

export async function postReset(signal?: AbortSignal): Promise<void> {
  const res = await fetch(`${BASE}/reset`, { method: 'POST', signal })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
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
        throw new Error(`HTTP ${res.status}`)
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

export async function searchComments(keywords: string[], signal?: AbortSignal): Promise<Comment[]> {
  const params = new URLSearchParams({ keywords: keywords.join(',') })
  return fetchWithRetry<Comment[]>(`${BASE}/comments?${params}`, { signal })
}
