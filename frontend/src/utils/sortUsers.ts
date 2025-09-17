import type { User } from './api'

/**
 * 並び順: 初回コメント時間(昇順) → channelId(昇順) → displayName(昇順)
 * firstCommentedAt が欠落している場合は末尾に寄せる。
 */
export function sortUsersStable(input: readonly User[]): User[] {
  const users = [...input]
  return users.sort((a, b) => {
    const ta = Date.parse(a.firstCommentedAt || '')
    const tb = Date.parse(b.firstCommentedAt || '')
    const aHas = Number.isFinite(ta)
    const bHas = Number.isFinite(tb)
    if (aHas && bHas && ta !== tb) return ta - tb
    if (aHas && !bHas) return -1
    if (!aHas && bHas) return 1

    // tie: use channelId
    const idA = (a.channelId || '').toLowerCase()
    const idB = (b.channelId || '').toLowerCase()
    if (idA && idB && idA !== idB) return idA.localeCompare(idB, 'en')

    // fallback: displayName
    const na = (a.displayName || '').toLowerCase()
    const nb = (b.displayName || '').toLowerCase()
    return na.localeCompare(nb, 'en')
  })
}

