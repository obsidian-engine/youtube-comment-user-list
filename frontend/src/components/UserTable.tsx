import { useState, useMemo, useEffect, useCallback } from 'react'
import { Tooltip } from './Tooltip'
import { isJapaneseTextTooLong, truncateJapaneseText } from '../utils/textUtils'
import { SELECT_CLASS } from '../utils/styles'

interface User {
  channelId?: string
  displayName?: string
  joinedAt?: string
  commentCount?: number
  firstCommentedAt?: string
  latestCommentedAt?: string
}

type UserData = User | string

interface UserTableProps {
  users: UserData[]
  intervalSec?: number
  setIntervalSec?: (value: number) => void
  isRefreshing?: boolean
  showCommentTime?: boolean
  onToggleCommentTime?: () => void
}

type SortField = 'commentCount' | 'firstCommentedAt'
type SortOrder = 'asc' | 'desc'

interface SortState {
  field: SortField | null
  order: SortOrder
}

const getUserDisplayName = (user: UserData): string => {
  if (typeof user === 'string') return user
  return user.displayName || user.channelId || 'Unknown User'
}

const getUserKey = (user: UserData, index: number): string => {
  if (typeof user === 'string') return `${user}-${index}`
  return `${user.channelId || user.displayName}-${index}`
}

const getUserCommentCount = (user: UserData): number => {
  if (typeof user === 'string') return 0
  return user.commentCount ?? 0
}

const getUserFirstComment = (user: UserData): string => {
  if (typeof user === 'string') return '--:--'
  if (user.firstCommentedAt && user.firstCommentedAt !== '') {
    return new Date(user.firstCommentedAt).toLocaleTimeString('ja-JP', {
      hour: '2-digit',
      minute: '2-digit',
      timeZone: 'Asia/Tokyo',
    })
  }
  return '--:--'
}

const getUserLatestComment = (user: UserData): string => {
  if (typeof user === 'string') return '--:--'
  if (user.latestCommentedAt && user.latestCommentedAt !== '') {
    return new Date(user.latestCommentedAt).toLocaleTimeString('ja-JP', {
      hour: '2-digit',
      minute: '2-digit',
      timeZone: 'Asia/Tokyo',
    })
  }
  return '--:--'
}

const getUserChannelUrl = (user: UserData): string | null => {
  if (typeof user === 'string') return null
  if (!user.channelId) return null
  return `https://www.youtube.com/channel/${user.channelId}`
}

const getUserFirstCommentTime = (user: UserData): Date | null => {
  if (typeof user === 'string') return null
  if (user.firstCommentedAt && user.firstCommentedAt !== '') {
    return new Date(user.firstCommentedAt)
  }
  return null
}

function CopyLinkButton({ url, displayName }: { url: string; displayName: string }) {
  const [copied, setCopied] = useState(false)
  const copyText = `${displayName}さん ${url}`

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(copyText)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      try {
        const textarea = document.createElement('textarea')
        textarea.value = copyText
        textarea.style.position = 'fixed'
        textarea.style.opacity = '0'
        document.body.appendChild(textarea)
        textarea.select()
        document.execCommand('copy')
        document.body.removeChild(textarea)
        setCopied(true)
        setTimeout(() => setCopied(false), 1500)
      } catch {
        // コピー失敗は無視
      }
    }
  }, [copyText])

  return (
    <button
      onClick={handleCopy}
      title="チャンネルURLをコピー"
      aria-label="チャンネルURLをコピー"
      style={{
        flexShrink: 0,
        background: 'none',
        border: 'none',
        cursor: 'pointer',
        color: copied ? 'var(--c-success)' : 'var(--c-ink-mute)',
        transition: 'color 0.2s',
        padding: '0 2px',
      }}
    >
      {copied ? (
        <svg
          className="w-4 h-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
        </svg>
      ) : (
        <svg
          className="w-4 h-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
          />
        </svg>
      )}
    </button>
  )
}

interface SortButtonProps {
  field: SortField
  currentSort: SortState
  onSort: (field: SortField) => void
  children: React.ReactNode
}

function SortButton({ field, currentSort, onSort, children }: SortButtonProps) {
  const isActive = currentSort.field === field
  const isDesc = isActive && currentSort.order === 'desc'
  const isAsc = isActive && currentSort.order === 'asc'
  const ariaLabel = field === 'commentCount' ? '発言数でソート' : '初回コメントでソート'

  return (
    <button
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: '4px',
        background: 'none',
        border: 'none',
        color: 'inherit',
        cursor: 'pointer',
        fontFamily: 'inherit',
        fontWeight: 'inherit',
        fontSize: 'inherit',
        letterSpacing: 'inherit',
        textTransform: 'inherit',
      }}
      onClick={() => onSort(field)}
      aria-label={ariaLabel}
    >
      {children}
      <span style={{ display: 'flex', flexDirection: 'column', width: '12px', height: '12px' }}>
        <svg
          style={{ width: '12px', height: '6px', color: isAsc ? '#fff' : 'rgba(255,255,255,0.3)' }}
          fill="currentColor"
          viewBox="0 0 12 6"
        >
          <path d="M6 0L12 6H0z" />
        </svg>
        <svg
          style={{ width: '12px', height: '6px', color: isDesc ? '#fff' : 'rgba(255,255,255,0.3)' }}
          fill="currentColor"
          viewBox="0 0 12 6"
        >
          <path d="M6 6L0 0h12z" />
        </svg>
      </span>
    </button>
  )
}

const thStyle: React.CSSProperties = {
  padding: '12px 16px',
  fontFamily: 'var(--f-mono)',
  fontWeight: 700,
  fontSize: '11px',
  letterSpacing: '0.18em',
  textTransform: 'uppercase',
  color: '#fff',
  textAlign: 'center',
  background: 'var(--c-ink)',
}

export function UserTable({
  users,
  intervalSec = 0,
  setIntervalSec,
  isRefreshing = false,
  showCommentTime = true,
  onToggleCommentTime,
}: UserTableProps) {
  const [sortState, setSortState] = useState<SortState>({ field: null, order: 'asc' })

  useEffect(() => {
    setSortState({ field: null, order: 'asc' })
  }, [users])

  const handleSort = (field: SortField) => {
    setSortState((prevState) => {
      if (prevState.field === field) {
        return { field, order: prevState.order === 'asc' ? 'desc' : 'asc' }
      } else {
        return { field, order: field === 'commentCount' ? 'desc' : 'asc' }
      }
    })
  }

  const handleReset = () => {
    setSortState({ field: null, order: 'asc' })
  }

  const isSorted = sortState.field !== null

  const sortedUsers = useMemo(() => {
    if (!sortState.field) return users

    return [...users].sort((a, b) => {
      if (sortState.field === 'commentCount') {
        const countA = getUserCommentCount(a)
        const countB = getUserCommentCount(b)
        return sortState.order === 'asc' ? countA - countB : countB - countA
      } else if (sortState.field === 'firstCommentedAt') {
        const timeA = getUserFirstCommentTime(a)
        const timeB = getUserFirstCommentTime(b)
        if (!timeA && !timeB) return 0
        if (!timeA) return 1
        if (!timeB) return -1
        const diff = timeA.getTime() - timeB.getTime()
        return sortState.order === 'asc' ? diff : -diff
      }
      return 0
    })
  }, [users, sortState])

  return (
    <section
      style={{
        overflow: 'hidden',
        border: '1px solid var(--c-line-strong)',
        background: 'var(--c-bg-2)',
      }}
    >
      {/* コントロールヘッダー */}
      <div
        style={{
          padding: '10px 16px',
          borderBottom: '1px solid var(--c-line)',
          background: 'var(--c-bg)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: '12px',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <button
            onClick={handleReset}
            disabled={!isSorted || isRefreshing}
            aria-label="ソートリセット"
            style={{
              fontFamily: 'var(--f-mono)',
              fontSize: '11px',
              letterSpacing: '0.14em',
              padding: '5px 10px',
              background: isSorted ? 'var(--c-ink)' : 'transparent',
              color: isSorted ? '#fff' : 'var(--c-ink-mute)',
              border: `1px solid ${isSorted ? 'var(--c-ink)' : 'var(--c-line-strong)'}`,
              cursor: isSorted ? 'pointer' : 'not-allowed',
              transition: 'all 0.2s',
            }}
          >
            ↻ 参加早い人が上
          </button>
          {setIntervalSec && (
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <label
                htmlFor="interval-select"
                style={{
                  fontFamily: 'var(--f-mono)',
                  fontSize: '10px',
                  letterSpacing: '0.14em',
                  textTransform: 'uppercase',
                  color: 'var(--c-ink-mute)',
                }}
              >
                更新間隔
              </label>
              <select
                id="interval-select"
                aria-label="更新間隔"
                value={intervalSec}
                onChange={(e) => setIntervalSec(Number(e.target.value))}
                disabled={isRefreshing}
                className={SELECT_CLASS}
              >
                <option value="0">停止</option>
                <option value="60">60s</option>
                <option value="90">90s</option>
                <option value="120">120s</option>
              </select>
            </div>
          )}
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          {onToggleCommentTime && (
            <button
              onClick={onToggleCommentTime}
              disabled={isRefreshing}
              aria-label="コメント時間表示切り替え"
              style={{
                fontFamily: 'var(--f-mono)',
                fontSize: '11px',
                letterSpacing: '0.14em',
                padding: '5px 10px',
                background: 'transparent',
                color: 'var(--c-ink-dim)',
                border: '1px solid var(--c-line-strong)',
                cursor: 'pointer',
              }}
            >
              {showCommentTime ? '時間非表示' : '時間表示'}
            </button>
          )}
          {isRefreshing && (
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <div
                data-testid="loading-spinner"
                style={{
                  width: '20px',
                  height: '20px',
                  border: '2px solid var(--c-line-strong)',
                  borderTopColor: 'var(--c-ink)',
                  borderRadius: '50%',
                  animation: 'spin 0.7s linear infinite',
                }}
              />
              <span
                style={{
                  fontFamily: 'var(--f-mono)',
                  fontSize: '11px',
                  color: 'var(--c-ink-dim)',
                }}
              >
                更新中...
              </span>
            </div>
          )}
        </div>
      </div>

      <table className="w-full table-fixed" style={{ fontSize: '14px', lineHeight: '1.7' }}>
        <thead>
          <tr>
            <th style={{ ...thStyle, width: '80px' }}>#</th>
            <th style={{ ...thStyle, width: '350px', maxWidth: '350px' }}>名前</th>
            <th style={{ ...thStyle, width: '120px' }}>
              <SortButton field="commentCount" currentSort={sortState} onSort={handleSort}>
                発言数
              </SortButton>
            </th>
            {showCommentTime && (
              <th style={{ ...thStyle, width: '150px' }} className="hidden md:table-cell">
                <SortButton field="firstCommentedAt" currentSort={sortState} onSort={handleSort}>
                  初回コメント
                </SortButton>
              </th>
            )}
            {showCommentTime && (
              <th style={{ ...thStyle, width: '150px' }} className="hidden md:table-cell">
                最新コメント
              </th>
            )}
          </tr>
        </thead>
        <tbody>
          {sortedUsers.map((user, i) => {
            const channelUrl = getUserChannelUrl(user)
            const rowBg = i % 2 === 0 ? 'var(--c-bg)' : 'var(--c-bg-2)'
            return (
              <tr
                key={getUserKey(user, i)}
                style={{
                  background: rowBg,
                  borderBottom: '1px solid var(--c-line)',
                  transition: 'background 0.15s',
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLTableRowElement).style.background =
                    'rgba(0,108,138,0.06)'
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLTableRowElement).style.background = rowBg
                }}
              >
                <td
                  style={{
                    padding: '10px 16px',
                    fontFamily: 'var(--f-mono)',
                    fontSize: '12px',
                    color: 'var(--c-ink-mute)',
                    textAlign: 'center',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {String(i + 1).padStart(2, '0')}
                </td>
                <td style={{ padding: '10px 16px', color: 'var(--c-ink)', fontWeight: 500 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                    <Tooltip
                      content={getUserDisplayName(user)}
                      disabled={!isJapaneseTextTooLong(getUserDisplayName(user), 30)}
                      className="block min-w-0 flex-1 max-w-[280px]"
                    >
                      <span className="block truncate">
                        {isJapaneseTextTooLong(getUserDisplayName(user), 30)
                          ? truncateJapaneseText(getUserDisplayName(user), 30)
                          : getUserDisplayName(user)}
                      </span>
                    </Tooltip>
                    {channelUrl && (
                      <CopyLinkButton url={channelUrl} displayName={getUserDisplayName(user)} />
                    )}
                  </div>
                </td>
                <td
                  style={{
                    padding: '10px 16px',
                    fontFamily: 'var(--f-mono)',
                    fontSize: '13px',
                    color: 'var(--c-ink-dim)',
                    textAlign: 'center',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                  data-testid={`comment-count-${i}`}
                >
                  {getUserCommentCount(user)}
                </td>
                {showCommentTime && (
                  <td
                    className="hidden md:table-cell"
                    style={{
                      padding: '10px 16px',
                      fontFamily: 'var(--f-mono)',
                      fontSize: '13px',
                      color: 'var(--c-ink-dim)',
                      textAlign: 'center',
                    }}
                    data-testid={`first-comment-${i}`}
                  >
                    {getUserFirstComment(user)}
                  </td>
                )}
                {showCommentTime && (
                  <td
                    className="hidden md:table-cell"
                    style={{
                      padding: '10px 16px',
                      fontFamily: 'var(--f-mono)',
                      fontSize: '13px',
                      color: 'var(--c-ink-dim)',
                      textAlign: 'center',
                    }}
                    data-testid={`latest-comment-${i}`}
                  >
                    {getUserLatestComment(user)}
                  </td>
                )}
              </tr>
            )
          })}
        </tbody>
      </table>
      {sortedUsers.length === 0 && (
        <p
          style={{
            padding: '20px 16px',
            fontFamily: 'var(--f-mono)',
            fontSize: '12px',
            color: 'var(--c-ink-mute)',
          }}
        >
          ユーザーがいません。
        </p>
      )}
    </section>
  )
}
