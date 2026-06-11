export type TabType = 'users' | 'comments' | 'votes' | 'logs' | 'history' | 'help'

interface TabsProps {
  activeTab: TabType
  onTabChange: (tab: TabType) => void
}

export function Tabs({ activeTab, onTabChange }: TabsProps) {
  const baseStyle: React.CSSProperties = {
    fontFamily: 'var(--f-mono)',
    fontSize: '12px',
    letterSpacing: '0.14em',
    background: 'transparent',
    border: 'none',
    borderBottom: '3px solid transparent',
    padding: '10px 16px 8px',
    cursor: 'pointer',
    transition: 'color 0.15s, border-color 0.15s',
  }

  const activeStyle: React.CSSProperties = {
    ...baseStyle,
    color: 'var(--c-ink)',
    fontWeight: 700,
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
      {tabs.map((t) => (
        <button
          key={t.id}
          role="tab"
          aria-selected={activeTab === t.id}
          onClick={() => onTabChange(t.id)}
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
