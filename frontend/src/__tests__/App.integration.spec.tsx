import { render, screen, fireEvent } from '@testing-library/react'
import App from '../App.jsx'

describe('App Integration (擬似)', () => {
  test('切替→ACTIVE、今すぐ取得→人数増加（擬似）', async () => {
    render(<App />)
    // 初期は ACTIVE（擬似データ）
    expect(await screen.findAllByText('ACTIVE')).toBeTruthy()

    // 切替（videoId 入力がないとエラー）
    fireEvent.click(screen.getByRole('button', { name: '切替' }))
    expect(await screen.findByRole('alert')).toBeInTheDocument()

    // 入力して切替（擬似: seed に置き換え）
    const input = screen.getByLabelText('videoId') as HTMLInputElement
    fireEvent.change(input, { target: { value: 'VID123' } })
    fireEvent.click(screen.getByRole('button', { name: '切替' }))
    expect(screen.queryByRole('alert')).toBeNull()

    // 今すぐ取得
    const before = screen.getByText(/参加者/).textContent as string
    fireEvent.click(screen.getByRole('button', { name: '今すぐ取得' }))
    // 参加者数表示が変化（簡易判定）
    expect(screen.getByText(/参加者/).textContent).not.toBe(before)
  })
})
