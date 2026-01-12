import type { Comment } from '../types'

export const BASE = import.meta.env.VITE_BACKEND_URL || ''

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
