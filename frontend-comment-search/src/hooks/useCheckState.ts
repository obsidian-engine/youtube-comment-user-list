import { useState, useCallback, useEffect } from 'react'
import { loadChecked, saveChecked, clearChecked } from '../utils/storage'

export function useCheckState() {
  const [checked, setChecked] = useState<Record<string, boolean>>(() => loadChecked())

  // LocalStorageへの保存
  useEffect(() => {
    saveChecked(checked)
  }, [checked])

  const toggle = useCallback((commentId: string) => {
    setChecked((prev) => ({
      ...prev,
      [commentId]: !prev[commentId],
    }))
  }, [])

  const isChecked = useCallback(
    (commentId: string) => {
      return !!checked[commentId]
    },
    [checked],
  )

  const clear = useCallback(() => {
    setChecked({})
    clearChecked()
  }, [])

  const checkedCount = Object.values(checked).filter(Boolean).length

  return { checked, toggle, isChecked, clear, checkedCount }
}
