import type { HistorySnapshot } from '../../utils/api'
import { formatSnapshotSavedAt } from '../../hooks/useAppState'
import { HistoryUserTable } from './HistoryUserTable'
import { HistoryCommentSearch } from './HistoryCommentSearch'
import { HistoryVotes } from './HistoryVotes'

interface HistoryDetailProps {
  snapshot: HistorySnapshot
  onBack: () => void
}

export function HistoryDetail({ snapshot, onBack }: HistoryDetailProps) {
  return (
    <div className="space-y-4">
      {/* ヘッダー */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flexWrap: 'wrap' }}>
        <button
          aria-label="戻る"
          onClick={onBack}
          style={{
            fontFamily: 'var(--f-mono)',
            fontSize: '11px',
            letterSpacing: '0.12em',
            textTransform: 'uppercase',
            padding: '6px 12px',
            background: 'transparent',
            color: 'var(--c-ink)',
            border: '1px solid var(--c-line-strong)',
            cursor: 'pointer',
          }}
        >
          ← 戻る
        </button>
        <span
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            padding: '4px 10px',
            fontFamily: 'var(--f-mono)',
            fontSize: '11px',
            letterSpacing: '0.1em',
            textTransform: 'uppercase',
            background: 'rgba(181, 107, 0, 0.1)',
            color: '#b56b00',
            border: '1px solid rgba(181, 107, 0, 0.3)',
          }}
        >
          閲覧モード (read-only)
        </span>
        <span>
          {snapshot.videoTitle ? (
            <span>
              <span
                style={{
                  display: 'block',
                  fontSize: '15px',
                  fontWeight: 600,
                  color: 'var(--c-ink)',
                  lineHeight: 1.4,
                }}
              >
                {snapshot.videoTitle}
              </span>
              {snapshot.channelTitle && (
                <span
                  style={{
                    display: 'block',
                    fontSize: '12px',
                    color: 'var(--c-ink-dim)',
                    marginTop: '2px',
                  }}
                >
                  {snapshot.channelTitle}
                </span>
              )}
              <span
                style={{
                  display: 'block',
                  fontFamily: 'var(--f-mono)',
                  fontSize: '11px',
                  color: 'var(--c-ink-mute)',
                  marginTop: '2px',
                }}
              >
                {snapshot.videoId}
              </span>
            </span>
          ) : (
            <span
              style={{
                fontFamily: 'var(--f-mono)',
                fontSize: '12px',
                color: 'var(--c-ink-dim)',
              }}
            >
              {snapshot.videoId}
            </span>
          )}
        </span>
        {snapshot.savedAt && (
          <span
            style={{
              fontFamily: 'var(--f-mono)',
              fontSize: '11px',
              color: 'var(--c-ink-mute)',
            }}
          >
            保存日時: {formatSnapshotSavedAt(snapshot.savedAt)}
          </span>
        )}
      </div>

      {/* 視聴者一覧 */}
      <section>
        <h3
          style={{
            fontFamily: 'var(--f-mono)',
            fontSize: '11px',
            letterSpacing: '0.2em',
            textTransform: 'uppercase',
            color: 'var(--c-accent-2)',
            marginBottom: '8px',
          }}
        >
          視聴者一覧 ({snapshot.users.length} 人)
        </h3>
        <HistoryUserTable users={snapshot.users} />
      </section>

      {/* コメント検索 */}
      <HistoryCommentSearch comments={snapshot.comments} />

      {/* 投票集計 */}
      <section>
        <h3
          style={{
            fontFamily: 'var(--f-mono)',
            fontSize: '11px',
            letterSpacing: '0.2em',
            textTransform: 'uppercase',
            color: 'var(--c-accent-2)',
            marginBottom: '8px',
          }}
        >
          投票集計
        </h3>
        <HistoryVotes
          comments={snapshot.comments}
          videoId={snapshot.videoId}
          savedAt={snapshot.savedAt}
        />
      </section>
    </div>
  )
}
