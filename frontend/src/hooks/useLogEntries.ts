import { useState, useCallback, useRef } from 'react'

export type LogLevel = 'info' | 'warn' | 'error'

export type LogEntry = {
  id: number
  timestamp: Date
  level: LogLevel
  message: string
  addedCount?: number
  skippedCount?: number
}

export function useLogEntries() {
  const [entries, setEntries] = useState<LogEntry[]>([])
  const nextId = useRef(1)

  const addEntry = useCallback(
    (level: LogLevel, message: string, data?: { addedCount?: number; skippedCount?: number }) => {
      const entry: LogEntry = {
        id: nextId.current++,
        timestamp: new Date(),
        level,
        message,
        ...data,
      }
      setEntries((prev) => [entry, ...prev])
    },
    [],
  )

  const clear = useCallback(() => {
    setEntries([])
  }, [])

  return { entries, addEntry, clear }
}
