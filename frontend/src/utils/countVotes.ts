import type { Comment } from './api'

export type VoteCounts = Record<string, number>
export type Voter = { channelId: string; displayName: string }
export type VoteVoters = Record<string, Voter[]>

// 1コメンター1票。channelId ごとに publishedAt 昇順で走査し、
// 最初に keywords のいずれかと完全一致するコメントを発見したらそのワードに投票確定。
// 同一 channelId の以降のコメントは集計対象外。
// マッチ判定: trim 後のメッセージ全体が keyword と完全一致（部分一致不可、前後空白のみ無視、
// 大文字小文字は厳密に区別する）。サーバ API は lowercase 部分一致で候補を広めに返すが、
// 本関数で厳密フィルタする方針。
export function initCounts(keywords: string[]): VoteCounts {
  return Object.fromEntries(keywords.map((k) => [k, 0])) as VoteCounts
}

export function initVoters(keywords: string[]): VoteVoters {
  return Object.fromEntries(keywords.map((k) => [k, [] as Voter[]])) as VoteVoters
}

export function countVotes(
  comments: Comment[],
  keywords: string[],
): { counts: VoteCounts; voters: VoteVoters } {
  const counts: VoteCounts = initCounts(keywords)
  const voters: VoteVoters = initVoters(keywords)
  if (keywords.length === 0) return { counts, voters }

  const keywordSet = new Set(keywords)

  const sorted = [...comments].sort((a, b) => a.publishedAt.localeCompare(b.publishedAt))

  const voted = new Set<string>()
  for (const c of sorted) {
    if (voted.has(c.channelId)) continue
    const trimmed = c.message.trim()
    if (!keywordSet.has(trimmed)) continue
    voted.add(c.channelId)
    counts[trimmed] += 1
    voters[trimmed].push({ channelId: c.channelId, displayName: c.displayName })
  }
  return { counts, voters }
}
