export type TabType = 'users' | 'comments' | 'votes' | 'logs' | 'history' | 'help'

interface TabsProps {
  activeTab: TabType
  onTabChange: (tab: TabType) => void
}

export function Tabs({ activeTab, onTabChange }: TabsProps) {
  const activeStyle: React.CSSProperties = {
    background: 'var(--c-ink)',
    color: '#fff',
    fontFamily: 'var(--f-mono)',
    fontSize: '12px',
    letterSpacing: '0.14em',
    textTransform: 'uppercase',
    border: '1px solid var(--c-ink)',
    padding: '8px 16px',
    cursor: 'pointer',
    transition: 'background 0.2s, color 0.2s, border-color 0.2s',
  }
  const inactiveStyle: React.CSSProperties = {
    background: 'transparent',
    color: 'var(--c-ink-dim)',
    fontFamily: 'var(--f-mono)',
    fontSize: '12px',
    letterSpacing: '0.14em',
    textTransform: 'uppercase',
    border: '1px solid var(--c-line-strong)',
    padding: '8px 16px',
    cursor: 'pointer',
    transition: 'background 0.2s, color 0.2s, border-color 0.2s',
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
    <div className="flex flex-wrap gap-1">
      {tabs.map((t) => (
        <button key={t.id} onClick={() => onTabChange(t.id)} style={tabStyle(t.id)}>
          {t.label}
        </button>
      ))}
    </div>
  )
}
