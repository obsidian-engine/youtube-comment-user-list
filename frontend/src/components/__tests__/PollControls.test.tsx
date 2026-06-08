import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
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
  const onLoadFile = vi.fn()
  const onClear = vi.fn()
  const onRecount = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  const renderControls = (props: Partial<Parameters<typeof PollControls>[0]> = {}) => {
    return render(
      <PollControls
        keywords={['hoge', 'fuga']}
        onLoadFile={onLoadFile}
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
      expect(screen.getByText('投票キーワード（txt から読み込み）')).toBeInTheDocument()
      expect(screen.getByText(/メモ帳で 1 行 1 ワードの txt/)).toBeInTheDocument()
    })

    it('キーワード chip を全件表示', () => {
      renderControls({ keywords: ['a', 'b', 'c'] })
      expect(screen.getByText('a')).toBeInTheDocument()
      expect(screen.getByText('b')).toBeInTheDocument()
      expect(screen.getByText('c')).toBeInTheDocument()
    })

    it('keywords 空時は説明文を表示', () => {
      renderControls({ keywords: [] })
      expect(
        screen.getByText(/キーワード未設定。txt を読み込むと一覧表示されます/),
      ).toBeInTheDocument()
    })

    it('最終更新時刻を表示', () => {
      renderControls({ lastUpdated: '23:45:12' })
      expect(screen.getByText(/最終更新: 23:45:12/)).toBeInTheDocument()
    })
  })

  describe('ボタン動作', () => {
    it('読み込みボタン → file input click トリガー', async () => {
      const user = userEvent.setup()
      const { container } = renderControls()
      const input = container.querySelector('input[type="file"]') as HTMLInputElement
      const clickSpy = vi.spyOn(input, 'click')

      const button = screen.getByRole('button', {
        name: 'キーワードtxtを読み込み',
      })
      await user.click(button)

      expect(clickSpy).toHaveBeenCalled()
    })

    it('ファイル選択 → onLoadFile 呼び出し', async () => {
      const { container } = renderControls()
      const input = container.querySelector('input[type="file"]') as HTMLInputElement
      const file = new File(['hoge\nfuga'], 'k.txt', { type: 'text/plain' })

      await userEvent.upload(input, file)

      await waitFor(() => {
        expect(onLoadFile).toHaveBeenCalledTimes(1)
        expect(onLoadFile.mock.calls[0][0]).toBeInstanceOf(File)
      })
    })

    it('ファイル選択後 input.value はリセット（同一ファイル再選択可能）', async () => {
      const { container } = renderControls()
      const input = container.querySelector('input[type="file"]') as HTMLInputElement
      const file = new File(['hoge'], 'k.txt', { type: 'text/plain' })

      await userEvent.upload(input, file)

      await waitFor(() => {
        expect(input.value).toBe('')
      })
    })

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

    it('サンプルtxtダウンロードボタン → click でエラーなく実行', async () => {
      const user = userEvent.setup()
      // Blob URL 生成系をモック（jsdom 未対応の API）
      // vi.stubGlobal で復元を vi.unstubAllGlobals に委ね、他テストへの汚染を防ぐ
      const createObjectURL = vi.fn().mockReturnValue('blob:test')
      const revokeObjectURL = vi.fn()
      vi.stubGlobal('URL', {
        ...URL,
        createObjectURL,
        revokeObjectURL,
      })

      renderControls()
      const button = screen.getByRole('button', {
        name: 'サンプルtxtをダウンロード',
      })
      await user.click(button)

      expect(createObjectURL).toHaveBeenCalled()
      expect(revokeObjectURL).toHaveBeenCalledWith('blob:test')
    })
  })

  describe('isLoading 状態', () => {
    it('isLoading=true で 読み込み/サンプル/クリア が disabled', () => {
      renderControls({ isLoading: true })
      expect(screen.getByRole('button', { name: 'キーワードtxtを読み込み' })).toBeDisabled()
      expect(screen.getByRole('button', { name: 'サンプルtxtをダウンロード' })).toBeDisabled()
      expect(screen.getByRole('button', { name: 'クリア' })).toBeDisabled()
    })

    it('isLoading=true で 集計ボタンは isLoading 経由で disabled', () => {
      renderControls({ isLoading: true })
      const button = screen.getByRole('button', { name: '今すぐ集計' })
      expect(button).toBeDisabled()
    })
  })

  describe('file input attributes', () => {
    it('accept=.txt,text/plain', () => {
      const { container } = renderControls()
      const input = container.querySelector('input[type="file"]') as HTMLInputElement
      expect(input.getAttribute('accept')).toBe('.txt,text/plain')
    })

    it('hidden で配置（aria 等を阻害しない）', () => {
      const { container } = renderControls()
      const input = container.querySelector('input[type="file"]') as HTMLInputElement
      expect(input.className).toContain('hidden')
    })
  })
})
