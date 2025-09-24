import { describe, test, expect } from 'vitest'
import { sortUsersStable } from './sortUsers'
import type { User } from './api'

describe('ユーザー並び順', () => {
  test('初回コメント時間→チャンネルIDの優先で安定ソートされる', () => {
    const users: User[] = [
      {
        channelId: 'UC3',
        displayName: 'User3',
        joinedAt: '2024-01-01T12:00:00Z',
        firstCommentedAt: '2024-01-01T10:30:00Z'
      },
      {
        channelId: 'UC1',
        displayName: 'User1',
        joinedAt: '2024-01-01T12:00:00Z',
        firstCommentedAt: '2024-01-01T10:00:00Z'
      },
      {
        channelId: 'UC2',
        displayName: 'User2',
        joinedAt: '2024-01-01T12:00:00Z',
        firstCommentedAt: '2024-01-01T10:00:00Z'  // UC1と同時間
      }
    ]

    const sorted = sortUsersStable(users)

    expect(sorted[0].channelId).toBe('UC1') // 10:00:00 + channelId順
    expect(sorted[1].channelId).toBe('UC2') // 10:00:00 + channelId順
    expect(sorted[2].channelId).toBe('UC3') // 10:30:00
  })

  test('firstCommentedAtがない場合は末尾に寄せる', () => {
    const users: User[] = [
      {
        channelId: 'UC2',
        displayName: 'NoFirstComment',
        joinedAt: '2024-01-01T12:00:00Z'
        // firstCommentedAt なし
      },
      {
        channelId: 'UC1',
        displayName: 'WithFirstComment',
        joinedAt: '2024-01-01T12:00:00Z',
        firstCommentedAt: '2024-01-01T10:00:00Z'
      }
    ]

    const sorted = sortUsersStable(users)

    expect(sorted[0].channelId).toBe('UC1') // firstCommentedAtがあるので先頭
    expect(sorted[1].channelId).toBe('UC2') // firstCommentedAtがないので末尾
  })

  test('無効なfirstCommentedAt（不正な日付文字列）は末尾に寄せる', () => {
    const users: User[] = [
      {
        channelId: 'UC2',
        displayName: 'InvalidDate',
        joinedAt: '2024-01-01T12:00:00Z',
        firstCommentedAt: 'invalid-date'
      },
      {
        channelId: 'UC1',
        displayName: 'ValidDate',
        joinedAt: '2024-01-01T12:00:00Z',
        firstCommentedAt: '2024-01-01T10:00:00Z'
      }
    ]

    const sorted = sortUsersStable(users)

    expect(sorted[0].channelId).toBe('UC1') // 有効な日付なので先頭
    expect(sorted[1].channelId).toBe('UC2') // 無効な日付なので末尾
  })

  test('全てfirstCommentedAtがない場合はchannelId順', () => {
    const users: User[] = [
      {
        channelId: 'UC3',
        displayName: 'User3',
        joinedAt: '2024-01-01T12:00:00Z'
      },
      {
        channelId: 'UC1',
        displayName: 'User1',
        joinedAt: '2024-01-01T12:00:00Z'
      },
      {
        channelId: 'UC2',
        displayName: 'User2',
        joinedAt: '2024-01-01T12:00:00Z'
      }
    ]

    const sorted = sortUsersStable(users)

    expect(sorted[0].channelId).toBe('UC1')
    expect(sorted[1].channelId).toBe('UC2')
    expect(sorted[2].channelId).toBe('UC3')
  })
})