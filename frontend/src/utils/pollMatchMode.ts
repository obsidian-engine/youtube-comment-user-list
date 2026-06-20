import type { MatchMode } from './countVotes'

export const POLL_MATCH_MODE_STORAGE_KEY = 'pollMatchMode'

export function loadStoredMatchMode(): MatchMode {
  try {
    const raw = localStorage.getItem(POLL_MATCH_MODE_STORAGE_KEY)
    if (raw === 'exact' || raw === 'partial') return raw
  } catch {
    // ignore
  }
  return 'exact'
}

export function saveStoredMatchMode(mode: MatchMode): void {
  try {
    localStorage.setItem(POLL_MATCH_MODE_STORAGE_KEY, mode)
  } catch {
    // ignore
  }
}
