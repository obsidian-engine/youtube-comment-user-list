import { render, screen } from '@testing-library/react'
import App from '../App.jsx'

describe('App UI', () => {
  test('ヘッダー/カウンタ/テーブルが描画される', async () => {
    render(<App />)
    expect(await screen.findByText(/参加ユーザー/)).toBeInTheDocument()
    expect(screen.getByText(/参加者/)).toBeInTheDocument()
    expect(screen.getByText('名前')).toBeInTheDocument()
  })
})
