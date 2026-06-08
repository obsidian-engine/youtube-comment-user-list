import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { PollControls } from '../PollTab/PollControls'

interface LoadingButtonProps {
  children: React.ReactNode
  onClick: () => void
  disabled?: boolean
  isLoading?: boolean
  [key: string]: unknown
}

vi.mock('../LoadingButton', () => ({
  LoadingButton: ({ children, onClick, disabled, isLoading, ...props }: LoadingButtonProps) => (
    <button onClick={onClick} disabled={(disabled as boolean) || (isLoading as boolean)} {...props}>
      {children}
    </button>
  ),
}))

describe('PollControls', () => {
  const onAddKeyword = vi.fn()
  const onRemoveKeyword = vi.fn()
  const onClear = vi.fn()
  const onRecount = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  const renderControls = (props: Partial<Parameters<typeof PollControls>[0]> = {}) => {
    return render(
      <PollControls
        keywords={['hoge', 'fuga']}
        onAddKeyword={onAddKeyword}
        onRemoveKeyword={onRemoveKeyword}
        onClear={onClear}
        onRecount={onRecount}
        isLoading={false}
        lastUpdated="12:34:56"
        {...props}
      />,
    )
  }

  describe('表示要素', () => {
    it('見出しと説明文を表示', () => {
      renderControls()
      expect(screen.getByText('投票キーワード（完全一致でカウント）')).toBeInTheDocument()
      expect(screen.getByText(/キーワードを 1 つずつ追加/)).toBeInTheDocument()
    })

    it('キーワード chip を全件表示', () => {
      renderControls({ keywords: ['a', 'b', 'c'] })
      expect(screen.getByText('a')).toBeInTheDocument()
      expect(screen.getByText('b')).toBeInTheDocument()
      expect(screen.getByText('c')).toBeInTheDocument()
    })

    it('keywords 空時は説明文を表示', () => {
      renderControls({ keywords: [] })
      expect(screen.getByText(/キーワード未設定。追加すると一覧表示されます/)).toBeInTheDocument()
    })

    it('最終更新時刻を表示', () => {
      renderControls({ lastUpdated: '23:45:12' })
      expect(screen.getByText(/最終更新: 23:45:12/)).toBeInTheDocument()
    })
  })

  describe('キーワード追加', () => {
    it('input に入力 → 追加ボタンクリックで onAddKeyword 呼び出し + input クリア', async () => {
      const user = userEvent.setup()
      renderControls({ keywords: [] })

      const input = screen.getByPlaceholderText('投票キーワードを入力') as HTMLInputElement
      await user.type(input, 'newword')
      const addBtn = screen.getByRole('button', { name: '追加' })
      await user.click(addBtn)

      expect(onAddKeyword).toHaveBeenCalledWith('newword')
      expect(input.value).toBe('')
    })

    it('空文字 / 空白のみは追加ボタン disabled', async () => {
      const user = userEvent.setup()
      renderControls({ keywords: [] })

      const addBtn = screen.getByRole('button', { name: '追加' })
      expect(addBtn).toBeDisabled()

      const input = screen.getByPlaceholderText('投票キーワードを入力')
      await user.type(input, '   ')
      expect(addBtn).toBeDisabled()
    })
  })

  describe('キーワード削除', () => {
    it('× ボタンで onRemoveKeyword 呼び出し', async () => {
      const user = userEvent.setup()
      renderControls({ keywords: ['hoge'] })

      const removeBtn = screen.getByLabelText('hogeを削除')
      await user.click(removeBtn)

      expect(onRemoveKeyword).toHaveBeenCalledWith('hoge')
    })
  })

  describe('ボタン動作', () => {
    it('クリアボタン: keywords あれば表示 → onClear 呼び出し', async () => {
      const user = userEvent.setup()
      renderControls({ keywords: ['hoge'] })

      const clearBtn = screen.getByRole('button', { name: 'クリア' })
      await user.click(clearBtn)

      expect(onClear).toHaveBeenCalled()
    })

    it('クリアボタン: keywords 空なら非表示', () => {
      renderControls({ keywords: [] })
      expect(screen.queryByRole('button', { name: 'クリア' })).toBeNull()
    })

    it('集計ボタン → onRecount 呼び出し', async () => {
      const user = userEvent.setup()
      renderControls()

      const button = screen.getByRole('button', { name: '今すぐ集計' })
      await user.click(button)

      expect(onRecount).toHaveBeenCalled()
    })

    it('集計ボタン: keywords 空なら disabled', () => {
      renderControls({ keywords: [] })
      const button = screen.getByRole('button', { name: '今すぐ集計' })
      expect(button).toBeDisabled()
    })
  })

  describe('isLoading 状態', () => {
    it('isLoading=true で 追加 input / クリアが disabled', () => {
      renderControls({ isLoading: true })
      expect(screen.getByPlaceholderText('投票キーワードを入力')).toBeDisabled()
      expect(screen.getByRole('button', { name: 'クリア' })).toBeDisabled()
    })

    it('isLoading=true で 集計ボタンは isLoading 経由で disabled', () => {
      renderControls({ isLoading: true })
      const button = screen.getByRole('button', { name: '今すぐ集計' })
      expect(button).toBeDisabled()
    })
  })
})
