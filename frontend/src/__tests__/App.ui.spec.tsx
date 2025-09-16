import { render, screen } from '@testing-library/react'
import App from '../App.jsx'

describe('App UI', () => {
  test('ヘッダー/カウンタ/テーブルが描画される', async () => {
    render(<App />)
    expect(await screen.findByText(/参加ユーザー/)).toBeInTheDocument()
    expect(screen.getAllByText(/参加者/).length).toBeGreaterThan(0)
    expect(screen.getByText('名前')).toBeInTheDocument()
  })
})
