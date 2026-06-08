import type { Comment } from './api'

export type VoteCounts = Record<string, number>

// 1コメンター1票。channelId ごとに publishedAt 昇順で走査し、
// 最初に keywords のいずれかと完全一致するコメントを発見したらそのワードに投票確定。
// 同一 channelId の以降のコメントは集計対象外。
// マッチ判定: trim 後のメッセージ全体が keyword と完全一致（部分一致不可、前後空白のみ無視、
// 大文字小文字は厳密に区別する）。サーバ API は lowercase 部分一致で候補を広めに返すが、
// 本関数で厳密フィルタする方針。
export function countVotes(comments: Comment[], keywords: string[]): VoteCounts {
  const counts: VoteCounts = Object.fromEntries(keywords.map((k) => [k, 0]))
  if (keywords.length === 0) return counts

  const keywordSet = new Set(keywords)

  const sorted = [...comments].sort((a, b) => a.publishedAt.localeCompare(b.publishedAt))

  const voted = new Set<string>()
  for (const c of sorted) {
    if (voted.has(c.channelId)) continue
    const trimmed = c.message.trim()
    if (!keywordSet.has(trimmed)) continue
    voted.add(c.channelId)
    counts[trimmed] += 1
  }
  return counts
}
