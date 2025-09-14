import { render, screen, fireEvent } from '@testing-library/react'
import App from '../App.jsx'

describe('App UI', () => {
  test('表示モード切替（chip ↔ table）', async () => {
    render(<App />)
    // 初期は chip
    expect(await screen.findByText('チップ')).toBeInTheDocument()
    // table に切替
    const modeSelect = screen.getByDisplayValue('チップ') as HTMLSelectElement
    fireEvent.change(modeSelect, { target: { value: '表' } })
    expect(screen.getByDisplayValue('表')).toBeInTheDocument()
  })

  test('状態バッジが WAITING を表示', async () => {
    render(<App />)
    expect(await screen.findByText('WAITING')).toBeInTheDocument()
  })
})
