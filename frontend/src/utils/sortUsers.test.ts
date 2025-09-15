import { describe, it, expect, beforeEach } from 'vitest'
import { __mock } from '../mocks/handlers'
import App from '../App'
import { render, screen, waitFor } from '@testing-library/react'
import { createElement } from 'react'

// TDD: 仕様
// 1) 参加時間 (joinedAt) 昇順
// 2) 参加時間が同一なら channelId 昇順（表示は従来どおり displayName）

function namesInTable(): string[] {
  const rows = document.querySelectorAll('tbody tr')
  return Array.from(rows).map((tr) => {
    const nameCell = tr.querySelectorAll('td')[1]
    return nameCell?.textContent?.trim() || ''
  })
}

describe('ユーザー並び順', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('参加時間→チャンネルIDの優先で安定ソートされる', async () => {
    // displayNameのアルファベット順とchannelId順が逆転するケースを含める
    __mock.users = [
      { channelId: 'UC5', displayName: 'Zoe', joinedAt: '2024-01-01T09:00:00.000Z' },
      { channelId: 'UC4', displayName: 'Amy', joinedAt: '2024-01-01T09:00:00.000Z' },
      { channelId: 'UC2', displayName: 'Bob', joinedAt: '2024-01-01T10:00:00.000Z' },
      { channelId: 'UC3', displayName: 'Charlie', joinedAt: '2024-01-01T10:00:00.000Z' },
      { channelId: 'UC1', displayName: 'Alice', joinedAt: '2024-01-01T10:00:00.000Z' },
    ]

    render(createElement(App))

    // テーブルが描画され、5名が表示されるまで待機
    await waitFor(() => expect(namesInTable().length).toBe(5))

    // 期待順: 09:00 グループは channelId 昇順（UC4→UC5 なので Amy→Zoe）
    //         10:00 グループは UC1→UC2→UC3（Alice→Bob→Charlie）
    expect(namesInTable()).toEqual(['Amy', 'Zoe', 'Alice', 'Bob', 'Charlie'])
  })
})
