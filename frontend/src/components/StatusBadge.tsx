import React from 'react'

export type LiveStatus = 'WAITING' | 'ACTIVE'

type Props = {
  status: LiveStatus
}

export function StatusBadge({ status }: Props) {
  const isActive = status === 'ACTIVE'
  const cls = `px-2 py-1 rounded font-medium ${
    isActive ? 'bg-green-100 text-green-700' : 'bg-gray-200 text-gray-700'
  }`
  return <span className={cls}>{status}</span>
}
