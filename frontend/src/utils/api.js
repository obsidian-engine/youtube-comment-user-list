const BASE = import.meta.env.VITE_BACKEND_URL || ''

async function json(res) {
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}

export async function getStatus(signal) {
  const res = await fetch(`${BASE}/status`, { signal })
  return json(res)
}

export async function getUsers(signal) {
  const res = await fetch(`${BASE}/users.json`, { signal })
  return json(res)
}

export async function postSwitchVideo(videoId) {
  const res = await fetch(`${BASE}/switch-video`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ videoId }),
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
}

export async function postPull() {
  const res = await fetch(`${BASE}/pull`, { method: 'POST' })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
}

export async function postReset() {
  const res = await fetch(`${BASE}/reset`, { method: 'POST' })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
}

