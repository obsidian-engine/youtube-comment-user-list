import { describe, test, expect } from 'vitest'
import { decideUsers } from '../decideUsers'
import { sortUsersStable } from '../sortUsers'
import type { User } from '../api'

const mkUser = (
  channelId: string,
  displayName: string,
  joinedAt: string,
  firstCommentedAt?: string,
  commentCount = 0,
): User => ({
  channelId,
  displayName,
  joinedAt,
  firstCommentedAt,
  commentCount,
})

describe('decideUsers', () => {
  test('clearUsers=true + fetched 非空 → sortUsersStable(fetched) と等価', () => {
    const prev: User[] = [mkUser('ch-old', 'Old User', '2024-01-01')]
    const fetched: User[] = [
      mkUser('ch-b', 'User B', '2024-01-02', '2024-02-02'),
      mkUser('ch-a', 'User A', '2024-01-01', '2024-02-01'),
    ]
    const result = decideUsers(prev, fetched, { clearUsers: true })
    expect(result).toEqual(sortUsersStable(fetched))
  })

  test('clearUsers=true + fetched 空 → []', () => {
    const prev: User[] = []
    const fetched: User[] = []
    const result = decideUsers(prev, fetched, { clearUsers: true })
    expect(result).toEqual([])
  })

  test('clearUsers=true + fetched 空 + prev 非空 → [] (強制 clear)', () => {
    const prev: User[] = [mkUser('ch-old', 'Old User', '2024-01-01')]
    const fetched: User[] = []
    const result = decideUsers(prev, fetched, { clearUsers: true })
    expect(result).toEqual([])
  })

  test('clearUsers=false + fetched 非空 → sortUsersStable(fetched) と等価', () => {
    const prev: User[] = [mkUser('ch-old', 'Old User', '2024-01-01')]
    const fetched: User[] = [
      mkUser('ch-b', 'User B', '2024-01-02', '2024-02-02'),
      mkUser('ch-a', 'User A', '2024-01-01', '2024-02-01'),
    ]
    const result = decideUsers(prev, fetched, { clearUsers: false })
    expect(result).toEqual(sortUsersStable(fetched))
  })

  test('clearUsers=false + fetched 空 + prev 非空 → prev (保持、同 array reference)', () => {
    const prev: User[] = [mkUser('ch-old', 'Old User', '2024-01-01')]
    const fetched: User[] = []
    const result = decideUsers(prev, fetched, { clearUsers: false })
    expect(result).toBe(prev)
  })

  test('clearUsers=false + fetched 空 + prev 空 → []', () => {
    const prev: User[] = []
    const fetched: User[] = []
    const result = decideUsers(prev, fetched, { clearUsers: false })
    expect(result).toEqual([])
  })

  test('clearUsers=false + fetched 非空 + prev 非空 → sortUsersStable(fetched) (server 優先)', () => {
    const prev: User[] = [mkUser('ch-old', 'Old User', '2024-01-01')]
    const fetched: User[] = [mkUser('ch-new', 'New User', '2024-02-01', '2024-02-10')]
    const result = decideUsers(prev, fetched, { clearUsers: false })
    expect(result).toEqual(sortUsersStable(fetched))
    expect(result).not.toBe(prev)
  })

  test('sortUsersStable 適用確認: 同 joinedAt の users で sortUsersStable と同じ順序になる', () => {
    const prev: User[] = []
    // firstCommentedAt で順序が決まる
    const fetched: User[] = [
      mkUser('ch-c', 'User C', '2024-01-01', '2024-03-01'),
      mkUser('ch-a', 'User A', '2024-01-01', '2024-01-01'),
      mkUser('ch-b', 'User B', '2024-01-01', '2024-02-01'),
    ]
    const result = decideUsers(prev, fetched, { clearUsers: false })
    const expected = sortUsersStable(fetched)
    expect(result).toEqual(expected)
    // ch-a が最初 (firstCommentedAt が最古)
    expect(result[0].channelId).toBe('ch-a')
    expect(result[1].channelId).toBe('ch-b')
    expect(result[2].channelId).toBe('ch-c')
  })
})
