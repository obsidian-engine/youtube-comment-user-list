import { useState, useEffect } from 'react'

export function useCurrentTime() {
  const [currentTime, setCurrentTime] = useState<string>('')

  useEffect(() => {
    const updateTime = () => {
      const now = new Date()
      const pad = (n: number) => String(n).padStart(2, '0')
      const timeString = `${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`
      setCurrentTime(timeString)
    }

    // 初回実行
    updateTime()

    // 1秒ごとに更新
    const intervalId = setInterval(updateTime, 1000)

    return () => clearInterval(intervalId)
  }, [])

  return currentTime
}