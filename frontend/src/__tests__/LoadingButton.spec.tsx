import { render, screen } from '@testing-library/react'
import { LoadingButton } from '../components/LoadingButton'

describe('LoadingButton', () => {
  test('isLoading=true で disabled/aria-busy とローディング表示', () => {
    render(
      <LoadingButton ariaLabel="送信" isLoading loadingText="送信中…">
        送信
      </LoadingButton>
    )
    const btn = screen.getByRole('button', { name: '送信中…' })
    expect(btn).toBeDisabled()
    expect(btn).toHaveAttribute('aria-busy', 'true')
  })
})

