import { useState, useCallback, useEffect } from 'react'

const CHECKED_KEY = 'comment-search-checked'

const loadChecked = (): Record<string, boolean> => {
  const data = localStorage.getItem(CHECKED_KEY)
  return data ? JSON.parse(data) : {}
}

const saveChecked = (checked: Record<string, boolean>): void => {
  localStorage.setItem(CHECKED_KEY, JSON.stringify(checked))
}

const clearChecked = (): void => {
  localStorage.removeItem(CHECKED_KEY)
}

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
