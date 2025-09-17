import { render, screen } from '@testing-library/react'
import { LoadingButton } from '../components/LoadingButton'

describe('LoadingButton', () => {
  test('ローディング中は spinner が表示される', () => {
    render(
      <LoadingButton ariaLabel="送信" isLoading loadingText="送信中…">
        送信
      </LoadingButton>
    )

    // ローディング中の表示テキストが表示されていることを確認
    expect(screen.getByText('送信中…')).toBeInTheDocument()

    // spinnerが表示されていることを確認 (aria-hidden="true"の要素)
    const spinner = document.querySelector('[aria-hidden="true"]')
    expect(spinner).toBeInTheDocument()
    expect(spinner).toHaveClass('animate-spin')
  })

  test('ローディング中でない場合は通常のボタンが表示される', () => {
    render(
      <LoadingButton ariaLabel="送信">
        送信
      </LoadingButton>
    )

    expect(screen.getByText('送信')).toBeInTheDocument()
    expect(screen.queryByText('送信中…')).not.toBeInTheDocument()

    // spinnerが表示されていないことを確認
    const spinner = document.querySelector('[aria-hidden="true"]')
    expect(spinner).not.toBeInTheDocument()
  })

  test('disabled状態の場合はボタンが無効化される', () => {
    render(
      <LoadingButton ariaLabel="送信" disabled>
        送信
      </LoadingButton>
    )

    const button = screen.getByRole('button')
    expect(button).toBeDisabled()
  })

  test('variant="outline"の場合は適切なクラスが適用される', () => {
    render(
      <LoadingButton ariaLabel="キャンセル" variant="outline">
        キャンセル
      </LoadingButton>
    )

    const button = screen.getByRole('button')
    expect(button).toHaveClass('border')
  })
})