import { Fragment, useState } from 'react'
import type { VoteCounts, VoteVoters } from '../../utils/countVotes'

interface PollResultsProps {
  keywords: string[]
  counts: VoteCounts
  voters: VoteVoters
  totalVotes: number
  isLoading: boolean
}

function voterListToTsv(
  voters: Array<{ displayName: string; channelId: string; handle?: string }>,
): string {
  return voters.map((v) => `${v.displayName}\t${v.handle || v.channelId}`).join('\n')
}

async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    return false
  }
}

const thStyle: React.CSSProperties = {
  padding: '12px 16px',
  fontFamily: 'var(--f-mono)',
  fontWeight: 700,
  fontSize: '11px',
  letterSpacing: '0.18em',
  textTransform: 'uppercase',
  color: '#fff',
  background: 'var(--c-ink)',
}

export function PollResults({ keywords, counts, voters, totalVotes, isLoading }: PollResultsProps) {
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  const [copiedKeyword, setCopiedKeyword] = useState<string | null>(null)

  if (keywords.length === 0) return null

  const toggleExpand = (word: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(word)) next.delete(word)
      else next.add(word)
      return next
    })
  }

  const handleCopy = async (word: string) => {
    const list = voters[word] ?? []
    if (list.length === 0) return
    const ok = await copyToClipboard(voterListToTsv(list))
    if (ok) {
      setCopiedKeyword(word)
      setTimeout(() => setCopiedKeyword((cur) => (cur === word ? null : cur)), 1500)
    }
  }

  return (
    <section className="card-editorial">
      <div className="eyebrow">
        TALLY · RESULTS
        <div className="eyebrow__rule" />
      </div>

      <table className="w-full" style={{ fontSize: '14px', marginTop: '8px' }}>
        <thead>
          <tr>
            <th style={{ ...thStyle, textAlign: 'left' }}>キーワード</th>
            <th style={{ ...thStyle, textAlign: 'right', width: '120px' }}>票数</th>
          </tr>
        </thead>
        <tbody>
          {keywords.map((word) => {
            const list = voters[word] ?? []
            const isOpen = expanded.has(word)
            return (
              <Fragment key={word}>
                <tr
                  style={{
                    borderBottom: '1px solid var(--c-line)',
                    cursor: 'pointer',
                    transition: 'background 0.15s',
                  }}
                  onClick={() => toggleExpand(word)}
                  onMouseEnter={(e) => {
                    (e.currentTarget as HTMLTableRowElement).style.background =
                      'rgba(0,95,120,0.06)'
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLTableRowElement).style.background = ''
                  }}
                >
                  <td style={{ padding: '12px 16px', color: 'var(--c-ink)' }}>
                    <span
                      style={{
                        display: 'inline-block',
                        width: '16px',
                        color: 'var(--c-accent-2)',
                        fontFamily: 'var(--f-mono)',
                      }}
                    >
                      {isOpen ? '▼' : '▶'}
                    </span>
                    {word}
                  </td>
                  <td
                    style={{
                      padding: '12px 16px',
                      textAlign: 'right',
                      fontFamily: 'var(--f-mono)',
                      fontWeight: 700,
                      color: 'var(--c-ink)',
                      fontVariantNumeric: 'tabular-nums',
                    }}
                  >
                    {counts[word] ?? 0}
                  </td>
                </tr>
                {isOpen && (
                  <tr
                    style={{ background: 'var(--c-bg)', borderBottom: '1px solid var(--c-line)' }}
                  >
                    <td colSpan={2} style={{ padding: '12px 16px' }}>
                      {list.length === 0 ? (
                        <div
                          style={{
                            fontFamily: 'var(--f-mono)',
                            fontSize: '12px',
                            color: 'var(--c-ink-mute)',
                          }}
                        >
                          投票したユーザーはいません
                        </div>
                      ) : (
                        <div className="space-y-2">
                          <div
                            style={{
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'space-between',
                            }}
                          >
                            <span
                              style={{
                                fontFamily: 'var(--f-mono)',
                                fontSize: '11px',
                                color: 'var(--c-ink-dim)',
                              }}
                            >
                              投票ユーザー ({list.length}人)
                            </span>
                            <button
                              onClick={(e) => {
                                e.stopPropagation()
                                void handleCopy(word)
                              }}
                              aria-label={
                                copiedKeyword === word ? 'コピー済' : 'クリップボードにコピー'
                              }
                              style={{
                                fontFamily: 'var(--f-mono)',
                                fontSize: '11px',
                                letterSpacing: '0.1em',
                                padding: '4px 10px',
                                background: 'transparent',
                                color: 'var(--c-ink)',
                                border: '1px solid var(--c-line-strong)',
                                cursor: 'pointer',
                              }}
                            >
                              {copiedKeyword === word ? 'コピー済' : '名前+ハンドルをコピー'}
                            </button>
                          </div>
                          <ul className="space-y-1">
                            {list.map((v) => (
                              <li
                                key={v.channelId}
                                style={{
                                  display: 'flex',
                                  alignItems: 'center',
                                  gap: '8px',
                                  fontSize: '13px',
                                  color: 'var(--c-ink)',
                                }}
                              >
                                <span>{v.displayName}</span>
                                <span
                                  style={{
                                    fontFamily: 'var(--f-mono)',
                                    fontSize: '11px',
                                    color: 'var(--c-ink-mute)',
                                  }}
                                >
                                  {v.handle || v.channelId}
                                </span>
                              </li>
                            ))}
                          </ul>
                        </div>
                      )}
                    </td>
                  </tr>
                )}
              </Fragment>
            )
          })}
        </tbody>
        <tfoot>
          <tr style={{ background: 'var(--c-bg)', borderTop: '1px solid var(--c-line-strong)' }}>
            <td
              style={{
                padding: '12px 16px',
                fontFamily: 'var(--f-mono)',
                fontSize: '12px',
                letterSpacing: '0.1em',
                textTransform: 'uppercase',
                color: 'var(--c-ink-dim)',
              }}
            >
              合計
            </td>
            <td
              style={{
                padding: '12px 16px',
                textAlign: 'right',
                fontFamily: 'var(--f-mono)',
                fontWeight: 700,
                color: 'var(--c-ink)',
                fontVariantNumeric: 'tabular-nums',
              }}
            >
              {totalVotes}
            </td>
          </tr>
        </tfoot>
      </table>

      {isLoading && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '16px',
          }}
        >
          <div
            data-testid="poll-loading-spinner"
            className="animate-spin"
            style={{
              width: '22px',
              height: '22px',
              border: '2px solid var(--c-line-strong)',
              borderTopColor: 'var(--c-ink)',
              borderRadius: '50%',
              animation: 'spin 0.7s linear infinite',
            }}
          />
        </div>
      )}
    </section>
  )
}
