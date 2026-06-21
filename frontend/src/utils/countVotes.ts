import type { Comment } from './api'

export type VoteCounts = Record<string, number>
export type Voter = { channelId: string; displayName: string; handle?: string; message: string }
export type VoteVoters = Record<string, Voter[]>
export type MatchMode = 'exact' | 'partial'

export function initCounts(keywords: string[]): VoteCounts {
  return Object.fromEntries(keywords.map((k) => [k, 0])) as VoteCounts
}

export function initVoters(keywords: string[]): VoteVoters {
  return Object.fromEntries(keywords.map((k) => [k, [] as Voter[]])) as VoteVoters
}

export function countVotes(
  comments: Comment[],
  keywords: string[],
  matchMode: MatchMode = 'exact',
): { counts: VoteCounts; voters: VoteVoters } {
  const counts: VoteCounts = initCounts(keywords)
  const voters: VoteVoters = initVoters(keywords)
  if (keywords.length === 0) return { counts, voters }

  const sorted = [...comments].sort((a, b) => a.publishedAt.localeCompare(b.publishedAt))

  // partial では包含関係にあるキーワード (例: 'ho' と 'hoge') を同時登録したとき、
  // 配列順ではなく最長一致を優先する。より具体的なキーワードに票を寄せるため。
  const partialKeywords =
    matchMode === 'partial' ? [...keywords].sort((a, b) => b.length - a.length) : keywords

  const voted = new Set<string>()
  for (const c of sorted) {
    if (voted.has(c.channelId)) continue
    const trimmed = c.message.trim()
    const matched =
      matchMode === 'exact'
        ? keywords.find((k) => trimmed === k)
        : partialKeywords.find((k) => trimmed.includes(k))
    if (matched === undefined) continue
    voted.add(c.channelId)
    counts[matched] += 1
    voters[matched].push({
      channelId: c.channelId,
      displayName: c.displayName,
      handle: c.handle,
      message: trimmed,
    })
  }
  return { counts, voters }
}
