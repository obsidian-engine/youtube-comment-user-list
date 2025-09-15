const BASE = import.meta.env.VITE_BACKEND_URL || ''

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
}

export async function getUsers(signal?: AbortSignal): Promise<User[]> {
  const res = await fetch(`${BASE}/users.json`, { signal })
  return json<User[]>(res)
}

export async function postSwitchVideo(videoId: string): Promise<void> {
  const res = await fetch(`${BASE}/switch-video`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ videoId }),
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
}

export async function postPull(): Promise<void> {
  const res = await fetch(`${BASE}/pull`, { method: 'POST' })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
}

export async function postReset(): Promise<void> {
  const res = await fetch(`${BASE}/reset`, { method: 'POST' })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
}

