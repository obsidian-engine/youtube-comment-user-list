type TabType = 'users' | 'comments'

interface TabsProps {
  activeTab: TabType
  onTabChange: (tab: TabType) => void
}

export function Tabs({ activeTab, onTabChange }: TabsProps) {
  const tabClass = (tab: TabType) =>
    activeTab === tab
      ? 'px-4 py-2 rounded-md bg-neutral-900 text-white dark:bg-white dark:text-neutral-900 text-[14px] font-medium transition'
      : 'px-4 py-2 rounded-md bg-transparent text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-white/10 text-[14px] transition'

  return (
    <div className="flex gap-1 p-1 rounded-lg bg-slate-100/80 dark:bg-white/5 backdrop-blur">
      <button onClick={() => onTabChange('users')} className={tabClass('users')}>
        名前読み上げ
      </button>
      <button onClick={() => onTabChange('comments')} className={tabClass('comments')}>
        コメント検索
      </button>
    </div>
  )
}
