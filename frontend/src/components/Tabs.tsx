import { useRef } from 'react'

export type TabType = 'users' | 'comments' | 'votes' | 'logs' | 'history' | 'help'

interface TabsProps {
  activeTab: TabType
  onTabChange: (tab: TabType) => void
}

const TABS: TabType[] = ['users', 'comments', 'votes', 'logs', 'history', 'help']

export function Tabs({ activeTab, onTabChange }: TabsProps) {
  const buttonRefs = useRef<(HTMLButtonElement | null)[]>([])

  const baseStyle: React.CSSProperties = {
    fontFamily: 'var(--f-mono)',
    fontSize: '11px',
    letterSpacing: '0.08em',
    background: 'transparent',
    border: 'none',
    borderBottomWidth: '3px',
    borderBottomStyle: 'solid',
    borderBottomColor: 'transparent',
    padding: '10px 16px 8px',
    cursor: 'pointer',
    transition: 'color 0.15s, border-color 0.15s',
  }

  const activeStyle: React.CSSProperties = {
    ...baseStyle,
    color: 'var(--c-ink)',
    fontWeight: 600,
    borderBottomColor: 'var(--c-accent)',
  }

  const inactiveStyle: React.CSSProperties = {
    ...baseStyle,
    color: 'var(--c-ink-dim)',
    fontWeight: 500,
  }

  const tabStyle = (tab: TabType) => (activeTab === tab ? activeStyle : inactiveStyle)

  const tabs: { id: TabType; label: string }[] = [
    { id: 'users', label: '名前読み上げ' },
    { id: 'comments', label: 'コメント検索' },
    { id: 'votes', label: '投票' },
    { id: 'logs', label: 'ログ' },
    { id: 'history', label: '履歴' },
    { id: 'help', label: 'ヘルプ' },
  ]

  const handleKeyDown = (e: React.KeyboardEvent, index: number) => {
    let nextIndex = index
    if (e.key === 'ArrowRight') nextIndex = (index + 1) % TABS.length
    else if (e.key === 'ArrowLeft') nextIndex = (index - 1 + TABS.length) % TABS.length
    else if (e.key === 'Home') nextIndex = 0
    else if (e.key === 'End') nextIndex = TABS.length - 1
    else return
    e.preventDefault()
    onTabChange(TABS[nextIndex])
    buttonRefs.current[nextIndex]?.focus()
  }

  return (
    <div
      role="tablist"
      style={{
        display: 'flex',
        flexWrap: 'wrap',
        gap: '0',
        borderBottom: '1px solid var(--c-line-strong)',
        width: '100%',
      }}
    >
      {tabs.map((t, i) => (
        <button
          key={t.id}
          ref={(el) => {
            buttonRefs.current[i] = el
          }}
          role="tab"
          aria-selected={activeTab === t.id}
          tabIndex={activeTab === t.id ? 0 : -1}
          onClick={() => onTabChange(t.id)}
          onKeyDown={(e) => handleKeyDown(e, i)}
          style={tabStyle(t.id)}
          onMouseEnter={(e) => {
            if (activeTab !== t.id) {
              (e.currentTarget as HTMLButtonElement).style.borderBottomColor =
                'rgba(10, 10, 15, 0.25)'
              ;(e.currentTarget as HTMLButtonElement).style.color = 'var(--c-ink)'
            }
          }}
          onMouseLeave={(e) => {
            if (activeTab !== t.id) {
              (e.currentTarget as HTMLButtonElement).style.borderBottomColor = 'transparent'
              ;(e.currentTarget as HTMLButtonElement).style.color = 'var(--c-ink-dim)'
            }
          }}
        >
          {t.label}
        </button>
      ))}
    </div>
  )
}
