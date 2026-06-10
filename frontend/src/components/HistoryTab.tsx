import { useEffect } from 'react'
import { useHistory } from '../hooks/useHistory'
import { HistoryList } from './HistoryTab/HistoryList'
import { HistoryDetail } from './HistoryTab/HistoryDetail'

export function HistoryTab() {
  const history = useHistory()

  useEffect(() => {
    void history.loadList()
    // loadList は useCallback で安定しているため依存配列から除外
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  if (history.selected) {
    return <HistoryDetail snapshot={history.selected} onBack={history.clearSelected} />
  }

  return (
    <HistoryList
      snapshots={history.snapshots}
      loading={history.loading}
      error={history.error}
      onSelect={history.select}
    />
  )
}
