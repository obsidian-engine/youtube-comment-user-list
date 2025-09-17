import { describe, it, expect, beforeEach } from 'vitest'
import { __mock } from '../mocks/handlers'
import App from '../App'
import { render, screen, waitFor } from '@testing-library/react'
import { createElement } from 'react'

// TDD: 仕様
// 1) 初回コメント時間 (firstCommentedAt) 昇順
// 2) 初回コメント時間が同一なら channelId 昇順（表示は従来どおり displayName）
// 3) firstCommentedAtがない場合は末尾に寄せる

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

  it('初回コメント時間→チャンネルIDの優先で安定ソートされる', async () => {
    // displayNameのアルファベット順とchannelId順が逆転するケース＋firstCommentedAtあり/なしの混在
    __mock.users = [
      { channelId: 'UC5', displayName: 'Zoe', joinedAt: '2024-01-01T09:00:00.000Z', firstCommentedAt: '2024-01-01T09:02:00.000Z' },
      { channelId: 'UC4', displayName: 'Amy', joinedAt: '2024-01-01T09:00:00.000Z', firstCommentedAt: '2024-01-01T09:02:00.000Z' },
      { channelId: 'UC2', displayName: 'Bob', joinedAt: '2024-01-01T10:00:00.000Z', firstCommentedAt: '2024-01-01T09:01:00.000Z' },
      { channelId: 'UC3', displayName: 'Charlie', joinedAt: '2024-01-01T10:00:00.000Z', firstCommentedAt: '' }, // コメントなし
      { channelId: 'UC1', displayName: 'Alice', joinedAt: '2024-01-01T10:00:00.000Z', firstCommentedAt: '2024-01-01T09:01:00.000Z' },
      { channelId: 'UC6', displayName: 'David', joinedAt: '2024-01-01T08:00:00.000Z' }, // firstCommentedAtフィールドなし
    ]

    render(createElement(App))

    // テーブルが描画され、6名が表示されるまで待機
    await waitFor(() => expect(namesInTable().length).toBe(6))

    // 期待順:
    // 1) 09:01 グループ: UC1→UC2（Alice→Bob）
    // 2) 09:02 グループ: UC4→UC5（Amy→Zoe）
    // 3) firstCommentedAtなし: Charlie→David
    expect(namesInTable()).toEqual(['Alice', 'Bob', 'Amy', 'Zoe', 'Charlie', 'David'])
  })
})
