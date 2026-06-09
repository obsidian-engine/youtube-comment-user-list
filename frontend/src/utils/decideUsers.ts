import { sortUsersStable } from './sortUsers'
import type { User } from './api'

/**
 * server から取得した users と既存 users から表示用 list を決定する。
 * - clearUsers=true: 強制置換 (切替・リセット時)
 * - clearUsers=false: server 空かつ既存ありなら既存保持 (配信終了 WAITING で視聴者一覧を残す)
 */
export function decideUsers(
  prev: User[],
  fetched: User[],
  options: { clearUsers: boolean },
): User[] {
  if (options.clearUsers) return sortUsersStable(fetched)
  if (fetched.length === 0 && prev.length > 0) return prev
  return sortUsersStable(fetched)
}
