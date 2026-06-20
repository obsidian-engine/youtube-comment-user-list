import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MatchModeDescription } from '../MatchModeDescription'

describe('MatchModeDescription', () => {
  describe('variant="poll"', () => {
    it('matchMode="exact" のとき完全一致の説明文を表示', () => {
      render(<MatchModeDescription matchMode="exact" variant="poll" />)
      expect(
        screen.getByText(
          'キーワードを 1 つずつ追加してください。コメントが完全一致した場合のみ 1 票としてカウントされます。',
        ),
      ).toBeInTheDocument()
    })

    it('matchMode="partial" のとき部分一致の説明文を表示', () => {
      render(<MatchModeDescription matchMode="partial" variant="poll" />)
      expect(
        screen.getByText(
          'キーワードを 1 つずつ追加してください。コメントにキーワードが含まれる場合に 1 票としてカウントされます。',
        ),
      ).toBeInTheDocument()
    })
  })

  describe('variant="history"', () => {
    it('matchMode="exact" のとき完全一致の説明文を表示', () => {
      render(<MatchModeDescription matchMode="exact" variant="history" />)
      expect(
        screen.getByText('コメントがキーワードと完全一致した場合に 1 票としてカウントします。'),
      ).toBeInTheDocument()
    })

    it('matchMode="partial" のとき部分一致の説明文を表示', () => {
      render(<MatchModeDescription matchMode="partial" variant="history" />)
      expect(
        screen.getByText('コメントにキーワードが含まれる場合に 1 票としてカウントします。'),
      ).toBeInTheDocument()
    })
  })
})
